CREATE TABLE `analytics_web`
(
    `id`               BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
    `web_code`         VARCHAR(45)         NOT NULL DEFAULT '',
    `name`             VARCHAR(45)         NOT NULL DEFAULT '',
    `created_at`        TIMESTAMP(1)        NOT NULL DEFAULT CURRENT_TIMESTAMP(1),
    `updated_at`        TIMESTAMP(1)        NULL     DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP(1),
    `deleted_at`        TIMESTAMP(1)        NULL     DEFAULT NULL,
    `version`          INT(11)             NOT NULL DEFAULT 0,
    `extra`            VARCHAR(1000)       NOT NULL DEFAULT '{}',
    `event_setup_mode` TINYINT(2)          NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    KEY `idx_web_code` (`web_code`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 12119375
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

CREATE TABLE `analytics_web_advertiser_relation`
(
    `id`          BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
    `web_id`      BIGINT(20) UNSIGNED NOT NULL,
    `adv_id`      BIGINT(20) UNSIGNED NOT NULL,
    `create_date` TIMESTAMP(1)        NOT NULL DEFAULT CURRENT_TIMESTAMP(1),
    PRIMARY KEY (`id`),
    KEY `idx_web_id` (`web_id`),
    KEY `idx_adv_id` (`adv_id`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 12119375
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

CREATE TABLE `analytics_msg`
(
    `id`        BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
    `adv_id`    BIGINT(20) UNSIGNED NOT NULL,
    `user_id`   BIGINT(20) UNSIGNED NOT NULL,
    `name`      VARCHAR(127)        NOT NULL,
    `created_at` TIMESTAMP(1)        NOT NULL DEFAULT CURRENT_TIMESTAMP(1),
    `updated_at` TIMESTAMP(1)        NULL     DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP(1),
    `deleted_at` TIMESTAMP(1)        NULL     DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_adv` (`adv_id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 12119375
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

CREATE TABLE `analytics_data_source_cost`
(
    `id`              BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
    `adv_id`          BIGINT(20) UNSIGNED NOT NULL,
    `source_type`     VARCHAR(80)         NOT NULL,
    `source_id`       BIGINT(20)          NOT NULL,
    `total_cost_7day` BIGINT(20)          NOT NULL,
    `p_date`          VARCHAR(20)         NOT NULL,
    `created_at`       TIMESTAMP           NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`       TIMESTAMP           NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_adv_source_cost` (`p_date`, `source_id`, `total_cost_7day`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 12119375
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;