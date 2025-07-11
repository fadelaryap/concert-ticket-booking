
CREATE DATABASE IF NOT EXISTS user_db;
CREATE DATABASE IF NOT EXISTS booking_db;
CREATE DATABASE IF NOT EXISTS payment_db;


USE user_db;
CREATE TABLE IF NOT EXISTS `users` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `created_at` datetime(3) DEFAULT NULL,
    `updated_at` datetime(3) DEFAULT NULL,
    `deleted_at` datetime(3) DEFAULT NULL,
    `username` varchar(255) NOT NULL UNIQUE,
    `email` varchar(255) NOT NULL UNIQUE,
    `password` varchar(255) NOT NULL,
    `role` varchar(255) DEFAULT 'user',
    `last_login` datetime(3) DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


INSERT IGNORE INTO `users` (`id`, `username`, `email`, `password`, `role`, `created_at`, `updated_at`) VALUES
(1, 'admin', 'admin@example.com', '$2a$10$WpP6Z3E7X2aR7gM1Y5z3D.u8.z5.a6.o9.l0.k7.i5.t4.j2', 'admin', NOW(), NOW());



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
    `description` text,
    `status` varchar(255) DEFAULT 'pending_seat_creation',
    `image_url` varchar(255) DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_concerts_deleted_at` (`deleted_at`),
    INDEX `idx_concerts_date` (`date`),
    INDEX `idx_concerts_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


INSERT IGNORE INTO `concerts` (`id`, `name`, `artist`, `date`, `venue`, `total_seats`, `available_seats`, `description`, `status`, `image_url`, `created_at`, `updated_at`) VALUES
(1, 'Konser Rock Legendaris', 'Band Idola', '2025-08-15 20:00:00', 'Stadion Utama', 600, 600, 'Konser paling ditunggu tahun ini!', 'pending_seat_creation', 'https://example.com/concert_image.jpg', NOW(), NOW());


USE booking_db;
CREATE TABLE IF NOT EXISTS `ticket_classes` ( -- <-- PASTIKAN TABEL INI ADA!
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `created_at` datetime(3) DEFAULT NULL,
    `updated_at` datetime(3) DEFAULT NULL,
    `deleted_at` datetime(3) DEFAULT NULL,
    `concert_id` bigint unsigned NOT NULL,
    `name` varchar(255) NOT NULL,
    `price` decimal(10,2) NOT NULL,
    `total_seats_in_class` int NOT NULL,
    `available_seats_in_class` int NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_ticket_class_per_concert_name` (`concert_id`, `name`),
    KEY `idx_ticket_classes_deleted_at` (`deleted_at`),
    CONSTRAINT `fk_ticket_classes_concert` FOREIGN KEY (`concert_id`) REFERENCES `concerts` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


INSERT IGNORE INTO `ticket_classes` (`id`, `concert_id`, `name`, `price`, `total_seats_in_class`, `available_seats_in_class`, `created_at`, `updated_at`) VALUES
(1, 1, 'Festival', 500000.00, 500, 500, NOW(), NOW()),
(2, 1, 'VIP', 1500000.00, 100, 100, NOW(), NOW());


USE booking_db;
CREATE TABLE IF NOT EXISTS `seats` ( -- <-- PASTIKAN TABEL INI SUDAH DIUBAH!
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `created_at` datetime(3) DEFAULT NULL,
    `updated_at` datetime(3) DEFAULT NULL,
    `deleted_at` datetime(3) DEFAULT NULL,
    `concert_id` bigint unsigned NOT NULL,
    `ticket_class_id` bigint unsigned NOT NULL, -- <-- PASTIKAN KOLOM INI ADA!
    `seat_number` varchar(255) NOT NULL,
    `status` varchar(255) NOT NULL DEFAULT 'available',
    `user_id` bigint unsigned DEFAULT NULL,
    `booking_id` varchar(36) DEFAULT NULL, -- Changed to varchar(36)
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_seat_number_per_class_per_concert` (`concert_id`, `ticket_class_id`, `seat_number`), -- Perhatikan unique key
    KEY `idx_seats_deleted_at` (`deleted_at`),
    KEY `idx_seats_concert_id` (`concert_id`),
    KEY `idx_seats_ticket_class_id` (`ticket_class_id`), -- <-- PASTIKAN INDEX INI ADA!
    KEY `idx_seats_status` (`status`),
    CONSTRAINT `fk_seats_concert` FOREIGN KEY (`concert_id`) REFERENCES `concerts` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_seats_ticket_class` FOREIGN KEY (`ticket_class_id`) REFERENCES `ticket_classes` (`id`) ON DELETE CASCADE -- <-- PASTIKAN CONSTRAINT INI ADA!
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

USE booking_db; -- Pastikan USE lagi jika file sebelumnya tidak ada
CREATE TABLE IF NOT EXISTS `bookings` (
    `id` varchar(36) NOT NULL, -- Changed to varchar(36) for UUID
    `created_at` datetime(3) DEFAULT NULL,
    `updated_at` datetime(3) DEFAULT NULL,
    `deleted_at` datetime(3) DEFAULT NULL,
    `user_id` bigint unsigned NOT NULL,
    `concert_id` bigint unsigned NOT NULL,
    `seat_ids` text NOT NULL, -- Storing as comma-separated string for simplicity. For complex needs, use JSON array type if DB supports or a join table.
    `total_price` decimal(10,2) NOT NULL,
    `status` varchar(255) NOT NULL DEFAULT 'pending',
    `payment_id` bigint unsigned DEFAULT NULL,
    `expires_at` datetime(3) DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_bookings_deleted_at` (`deleted_at`),
    KEY `idx_bookings_user_id` (`user_id`),
    KEY `idx_bookings_concert_id` (`concert_id`),
    KEY `idx_bookings_status` (`status`),
    CONSTRAINT `fk_bookings_concert` FOREIGN KEY (`concert_id`) REFERENCES `concerts` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

USE booking_db; -- Pastikan USE lagi
CREATE TABLE IF NOT EXISTS `booking_seats` (
    `booking_id` varchar(36) NOT NULL, -- Changed to varchar(36)
    `seat_id` bigint unsigned NOT NULL,
    PRIMARY KEY (`booking_id`, `seat_id`),
    KEY `idx_booking_seats_booking_id` (`booking_id`),
    KEY `idx_booking_seats_seat_id` (`seat_id`),
    CONSTRAINT `fk_booking_seats_booking` FOREIGN KEY (`booking_id`) REFERENCES `bookings` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_booking_seats_seat` FOREIGN KEY (`seat_id`) REFERENCES `seats` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


USE payment_db;
CREATE TABLE IF NOT EXISTS `payments` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `created_at` datetime(3) DEFAULT NULL,
    `updated_at` datetime(3) DEFAULT NULL,
    `deleted_at` datetime(3) DEFAULT NULL,
    `booking_id` varchar(36) NOT NULL, -- Changed to varchar(36)
    `amount` decimal(10,2) NOT NULL,
    `payment_method` varchar(255) NOT NULL,
    `transaction_id` varchar(255) DEFAULT NULL,
    `status` varchar(255) NOT NULL DEFAULT 'pending', -- pending, completed, failed, refunded
    `payment_gateway_response` text,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_transaction_id` (`transaction_id`), -- Transaction ID from gateway should be unique
    KEY `idx_payments_deleted_at` (`deleted_at`),
    KEY `idx_payments_booking_id` (`booking_id`), -- Index for quickly finding payments by booking ID
    KEY `idx_payments_status` (`status`) -- Index for filtering by status
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


USE booking_db;
CREATE TABLE IF NOT EXISTS `buyers` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `created_at` datetime(3) DEFAULT NULL,
    `updated_at` datetime(3) DEFAULT NULL,
    `deleted_at` datetime(3) DEFAULT NULL,
    `booking_id` varchar(36) NOT NULL, -- Changed to varchar(36) for UUID
    `full_name` varchar(255) NOT NULL,
    `phone_number` varchar(255) NOT NULL,
    `email` varchar(255) NOT NULL,
    `ktp_number` varchar(255) NOT NULL UNIQUE, -- KTP number unique for a buyer
    PRIMARY KEY (`id`),
    KEY `idx_buyers_deleted_at` (`deleted_at`),
    KEY `idx_buyers_booking_id` (`booking_id`),
    CONSTRAINT `fk_buyers_booking` FOREIGN KEY (`booking_id`) REFERENCES `bookings` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

USE booking_db;
CREATE TABLE IF NOT EXISTS `ticket_holders` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `created_at` datetime(3) DEFAULT NULL,
    `updated_at` datetime(3) DEFAULT NULL,
    `deleted_at` datetime(3) DEFAULT NULL,
    `booking_id` varchar(36) NOT NULL, -- Changed to varchar(36) for UUID
    `full_name` varchar(255) NOT NULL,
    `ktp_number` varchar(255) NOT NULL UNIQUE, -- KTP number unique for a ticket holder
    PRIMARY KEY (`id`),
    KEY `idx_ticket_holders_deleted_at` (`deleted_at`),
    KEY `idx_ticket_holders_booking_id` (`booking_id`),
    CONSTRAINT `fk_ticket_holders_booking` FOREIGN KEY (`booking_id`) REFERENCES `bookings` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;