CREATE TABLE trending_repositories (
    `id` int NOT NULL AUTO_INCREMENT,
    `full_name` varchar(255) NOT NULL,
    `language` varchar(255) DEFAULT NULL,
    `rank` int NOT NULL,
    `scraped_at` datetime NOT NULL,,
    `trend_date` date NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;