CREATE TABLE `tracker_cost` (
    `id`              BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id`          BIGINT(20) UNSIGNED NOT NULL,
    `tracker_type`     VARCHAR(80)         NOT NULL,
    `tracker_id`       BIGINT(20) UNSIGNED NOT NULL,
    `p_date`           VARCHAR(20)         NOT NULL,
    `total_cost_7day`  BIGINT(20)          NOT NULL,
    `created_at`       TIMESTAMP(1)           NULL DEFAULT CURRENT_TIMESTAMP(1),
    `updated_at`       TIMESTAMP(1)           NULL DEFAULT CURRENT_TIMESTAMP(1) ON UPDATE CURRENT_TIMESTAMP(1),
    PRIMARY KEY (`id`),
    KEY `idx_user_tracker_date` (`user_id`, `tracker_id`, `p_date`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 12119375
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;