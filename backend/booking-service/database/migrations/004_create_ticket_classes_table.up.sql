CREATE TABLE IF NOT EXISTS `ticket_classes` (
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
    UNIQUE KEY `idx_ticket_class_per_concert` (`concert_id`, `name`),
    KEY `idx_ticket_classes_deleted_at` (`deleted_at`),
    CONSTRAINT `fk_ticket_classes_concert` FOREIGN KEY (`concert_id`) REFERENCES `concerts` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;