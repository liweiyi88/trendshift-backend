CREATE TABLE repository_monthly_insights (
    `id` INT NOT NULL AUTO_INCREMENT,
    `year` INT NOT NULL,
    `month` INT NOT NULL,
    `stars` INT DEFAULT NULL,
    `forks` INT DEFAULT NULL,
    `merged_prs` INT DEFAULT NULL,
    `issues` INT DEFAULT NULL,
    `closed_issues` INT DEFAULT NULL,
    `completed_at` DATETIME(3) DEFAULT NULL,
    `last_ingested_at` DATETIME(3) DEFAULT NULL,
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `repository_id` INT NOT NULL,
    KEY `IDX_1A2A6C8D8542BDZ5` (`repository_id`),
    UNIQUE KEY unique_month_year (repository_id, `year`, `month`),
    CONSTRAINT `FK_ACEGTDNYZDMEBQFA` FOREIGN KEY (`repository_id`) REFERENCES `repositories` (`id`),
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE repositories
ADD COLUMN skipped BOOLEAN NOT NULL DEFAULT FALSE,
ADD INDEX idx_repositories_skipped (skipped);

ALTER TABLE developers
ADD COLUMN skipped BOOLEAN NOT NULL DEFAULT FALSE,
ADD INDEX idx_developers_skipped (skipped);