CREATE TABLE repositories (
    `id` INT NOT NULL AUTO_INCREMENT,
    `ghr_id` INT NOT NULL,
    `stars` INT NOT NULL,
    `forks` INT NOT NULL,
    `full_name` varchar(255) NOT NULL,
    `language` varchar(255) DEFAULT NULL,
    `owner` varchar(255) NOT NULL,
    `owner_avatar_url` varchar(255) NOT NULL,
    `created_at` datetime NOT NULL,
    `updated_at` datetime NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (`full_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE trending_repositories (
    `id` INT NOT NULL AUTO_INCREMENT,
    `full_name` varchar(255) NOT NULL,
    `language` varchar(255) DEFAULT NULL,
    `rank` INT NOT NULL,
    `scraped_at` datetime NOT NULL,
    `trend_date` date NOT NULL,
    `repository_id` INT DEFAULT NULL,
    PRIMARY KEY (`id`),
    FOREIGN KEY (repository_id) REFERENCES repositories(id),
    UNIQUE (`full_name`, `language`, `trend_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;