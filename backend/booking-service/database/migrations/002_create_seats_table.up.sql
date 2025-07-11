CREATE TABLE IF NOT EXISTS `seats` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `created_at` datetime(3) DEFAULT NULL,
    `updated_at` datetime(3) DEFAULT NULL,
    `deleted_at` datetime(3) DEFAULT NULL,
    `concert_id` bigint unsigned NOT NULL,
    `seat_number` varchar(255) NOT NULL,
    `status` varchar(255) NOT NULL DEFAULT 'available',
    `user_id` bigint unsigned DEFAULT NULL,
    `booking_id` bigint unsigned DEFAULT NULL,
    `ticket_class_id` bigint unsigned NOT NULL, -- <-- TAMBAH
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_seat_number_per_class` (`ticket_class_id`, `seat_number`),
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_seat_number_per_concert` (`concert_id`, `seat_number`), -- Ensure unique seat per concert
    KEY `idx_seats_deleted_at` (`deleted_at`),
    KEY `idx_seats_concert_id` (`concert_id`),
    KEY `idx_seats_status` (`status`), -- Index for filtering by status
    CONSTRAINT `fk_seats_concert` FOREIGN KEY (`concert_id`) REFERENCES `concerts` (`id`) ON DELETE CASCADE
    CONSTRAINT `fk_seats_ticket_class` FOREIGN KEY (`ticket_class_id`) REFERENCES `ticket_classes` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;