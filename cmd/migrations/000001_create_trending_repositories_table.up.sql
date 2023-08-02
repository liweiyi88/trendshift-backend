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
    PRIMARY KEY (`id`),
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
    KEY `IDX_5A8A6C8D8545BDF5` (`repository_id`),
    CONSTRAINT `FK_XZPGTDNYZELENQFI` FOREIGN KEY (`repository_id`) REFERENCES `repositories` (`id`),
    UNIQUE (`full_name`, `language`, `trend_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE tags (
    `id` INT NOT NULL AUTO_INCREMENT,
    `name` varchar(255) NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `repositories_tags` (
  `repository_id` int NOT NULL,
  `tag_id` int NOT NULL,
  PRIMARY KEY (`repository_id`,`tag_id`),
  KEY `IDX_CPJZDYMXGPWTCNYA` (`repository_id`),
  KEY `IDX_DIVFKQKQHBOBOYBD` (`tag_id`),
  CONSTRAINT `FK_WSHAIEZHCUJVVCKB` FOREIGN KEY (`repository_id`) REFERENCES `repositories` (`id`) ON DELETE CASCADE,
  CONSTRAINT `FK_BFONQXGBCQOTMHMQ` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;