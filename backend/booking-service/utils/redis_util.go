package utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis(addr, password string, db int) error {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	var err error
	var counts uint8 = 1
	const maxRetries = 10
	const retryDelay = 3 * time.Second

	for counts <= maxRetries {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err = client.Ping(ctx).Result()
		cancel()

		if err != nil {
			log.Printf("Attempt %d/%d: Failed to connect to Redis: %v. Retrying in %s...", counts, maxRetries, err, retryDelay)
			time.Sleep(retryDelay)
			counts++
			continue
		} else {
			LogInfo("Connected to Redis successfully for rate limiting!")
			RedisClient = client
			return nil
		}
	}

	if err != nil {
		log.Fatalf("Failed to connect to Redis after %d retries: %v", maxRetries, err)
	}
	return errors.New("failed to connect to Redis after retries")
}

func CloseRedisConnection() {
	if RedisClient != nil {
		RedisClient.Close()
		LogInfo("Redis connection closed.")
	}
}

func SetAvailableSeatsCache(ctx context.Context, concertID uint, availableSeats int) error {
	key := fmt.Sprintf("concert:%d:available_seats", concertID)
	return RedisClient.Set(ctx, key, availableSeats, 0).Err()
}

func SetAvailableSeatsCacheByClass(ctx context.Context, ticketClassID uint, availableSeats int) error {
	key := fmt.Sprintf("ticket_class:%d:available_seats", ticketClassID)
	return RedisClient.Set(ctx, key, availableSeats, 0).Err()
}

func GetAvailableSeatsFromCache(ctx context.Context, concertID uint) (int, error) {
	key := fmt.Sprintf("concert:%d:available_seats", concertID)
	val, err := RedisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, fmt.Errorf("available seats not found in cache for concert ID %d", concertID)
	}
	if err != nil {
		return 0, fmt.Errorf("redis error getting available seats: %w", err)
	}
	seats, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid available seats value in cache for concert ID %d: %w", concertID, err)
	}
	return seats, nil
}

func GetAvailableSeatsCacheByClass(ctx context.Context, ticketClassID uint) (int, error) {
	key := fmt.Sprintf("ticket_class:%d:available_seats", ticketClassID)
	val, err := RedisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, fmt.Errorf("available seats not found in cache for ticket class ID %d", ticketClassID)
	}
	if err != nil {
		return 0, fmt.Errorf("redis error getting available seats for ticket class: %w", err)
	}
	seats, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid available seats value in cache for ticket class ID %d: %w", ticketClassID, err)
	}
	return seats, nil
}

func DecreaseAvailableSeatsAtomically(ctx context.Context, ticketClassID uint, numSeats int) (int64, error) {
	key := fmt.Sprintf("ticket_class:%d:available_seats", ticketClassID)

	txf := func(tx *redis.Tx) error {
		n, err := tx.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			return err
		}
		if err == redis.Nil {

			n = 0
		}

		if n < numSeats {
			return fmt.Errorf("not enough available seats for class. Current: %d, Requested: %d", n, numSeats)
		}

		_, err = tx.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, key, n-numSeats, 0)
			return nil
		})
		return err
	}

	for retries := 0; retries < 5; retries++ {
		err := RedisClient.Watch(ctx, txf, key)
		if err == nil {
			return RedisClient.Get(ctx, key).Int64()
		}
		if err == redis.TxFailedErr {
			LogWarning("Redis transaction failed, retrying for ticket class %d. Attempt: %d", ticketClassID, retries+1)
			continue
		}
		return 0, err
	}
	return 0, errors.New("failed to decrease available seats after multiple retries due to contention")
}

func IncreaseAvailableSeatsAtomically(ctx context.Context, ticketClassID uint, numSeats int) (int64, error) {
	key := fmt.Sprintf("ticket_class:%d:available_seats", ticketClassID)

	txf := func(tx *redis.Tx) error {
		n, err := tx.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			return err
		}
		if err == redis.Nil {
			n = 0
		}

		_, err = tx.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, key, n+numSeats, 0)
			return nil
		})
		return err
	}

	for retries := 0; retries < 5; retries++ {
		err := RedisClient.Watch(ctx, txf, key)
		if err == nil {
			return RedisClient.Get(ctx, key).Int64()
		}
		if err == redis.TxFailedErr {
			LogWarning("Redis transaction failed during seat increase, retrying for ticket class %d. Attempt: %d", ticketClassID, retries+1)
			continue
		}
		return 0, err
	}
	return 0, errors.New("failed to increase available seats after multiple retries due to contention")
}
