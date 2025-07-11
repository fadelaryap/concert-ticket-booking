CREATE TABLE IF NOT EXISTS `bookings` (
    `id` varchar(36) NOT NULL, -- Changed to varchar(36) for UUID
    `created_at` datetime(3) DEFAULT NULL,
    `updated_at` datetime(3) DEFAULT NULL,
    `deleted_at` datetime(3) DEFAULT NULL,
    `user_id` bigint unsigned NOT NULL,
    `concert_id` bigint unsigned NOT NULL,
    `seat_ids` text NOT NULL, -- Storing as comma-separated string for simplicity. For complex needs, use JSON array type if DB supports or a join table.
    `total_price` decimal(10,2) NOT NULL,
    `status` varchar(255) NOT NULL DEFAULT 'pending', -- pending, confirmed, cancelled, failed
    `payment_id` bigint unsigned DEFAULT NULL,
    `expires_at` datetime(3) DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_bookings_deleted_at` (`deleted_at`),
    KEY `idx_bookings_user_id` (`user_id`),
    KEY `idx_bookings_concert_id` (`concert_id`),
    KEY `idx_bookings_status` (`status`), -- Index for filtering by status
    CONSTRAINT `fk_bookings_concert` FOREIGN KEY (`concert_id`) REFERENCES `concerts` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


CREATE TABLE IF NOT EXISTS `booking_seats` (
    `booking_id` varchar(36) NOT NULL, -- Changed to varchar(36)
    `seat_id` bigint unsigned NOT NULL,
    PRIMARY KEY (`booking_id`, `seat_id`),
    KEY `idx_booking_seats_booking_id` (`booking_id`),
    KEY `idx_booking_seats_seat_id` (`seat_id`),
    CONSTRAINT `fk_booking_seats_booking` FOREIGN KEY (`booking_id`) REFERENCES `bookings` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_booking_seats_seat` FOREIGN KEY (`seat_id`) REFERENCES `seats` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;