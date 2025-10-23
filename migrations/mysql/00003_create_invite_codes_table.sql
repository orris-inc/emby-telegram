-- +goose Up
CREATE TABLE IF NOT EXISTS invite_codes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(20) NOT NULL UNIQUE,
    max_uses INT NOT NULL DEFAULT -1,
    current_uses INT NOT NULL DEFAULT 0,
    description VARCHAR(200),
    expire_at TIMESTAMP NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_invite_codes_code (code),
    INDEX idx_invite_codes_status (status),
    INDEX idx_invite_codes_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS invite_code_usage (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    invite_code_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_invite_code_usage_invite_code_id (invite_code_id),
    INDEX idx_invite_code_usage_user_id (user_id),
    FOREIGN KEY (invite_code_id) REFERENCES invite_codes(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS invite_code_usage;
DROP TABLE IF EXISTS invite_codes;
