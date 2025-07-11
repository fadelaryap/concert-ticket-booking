CREATE TABLE IF NOT EXISTS `payments` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `created_at` datetime(3) DEFAULT NULL,
    `updated_at` datetime(3) DEFAULT NULL,
    `deleted_at` datetime(3) DEFAULT NULL,
    `booking_id` bigint unsigned NOT NULL,
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