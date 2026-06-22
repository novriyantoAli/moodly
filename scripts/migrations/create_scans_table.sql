-- Create scans table for barcode scanning feature
CREATE TABLE IF NOT EXISTS scans (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    barcode VARCHAR(255) NOT NULL,
    timestamp BIGINT NOT NULL,
    transaction_id VARCHAR(255) NOT NULL UNIQUE,
    pin VARCHAR(255) NOT NULL,
    photo LONGTEXT NOT NULL,
    device_info JSON,
    photo_size VARCHAR(50),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes for performance
    INDEX idx_user_id (user_id),
    INDEX idx_barcode (barcode),
    INDEX idx_transaction_id (transaction_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
