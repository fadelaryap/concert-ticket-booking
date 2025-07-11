CREATE DATABASE IF NOT EXISTS booking_db;
USE booking_db;

CREATE TABLE IF NOT EXISTS `concerts` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `created_at` datetime(3) DEFAULT NULL,
    `updated_at` datetime(3) DEFAULT NULL,
    `deleted_at` datetime(3) DEFAULT NULL,
    `name` varchar(255) NOT NULL,
    `artist` varchar(255) NOT NULL,
    `date` datetime(3) NOT NULL,
    `venue` varchar(255) NOT NULL,
    `total_seats` int NOT NULL,
    `available_seats` int NOT NULL,
    `price_per_seat` decimal(10,2) NOT NULL,
    `description` text,
    `status` varchar(255) DEFAULT 'pending_seat_creation',
    PRIMARY KEY (`id`),
    KEY `idx_concerts_deleted_at` (`deleted_at`),
    INDEX `idx_concerts_date` (`date`),
    INDEX `idx_concerts_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;