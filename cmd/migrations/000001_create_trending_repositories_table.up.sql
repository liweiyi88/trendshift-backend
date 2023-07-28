CREATE TABLE repositories (
    id INT NOT NULL,
    PRIMARY KEY (id)
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
    FOREIGN KEY (repository_id) REFERENCES repositories(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;