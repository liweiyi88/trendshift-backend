CREATE TABLE developers (
    `id` INT NOT NULL AUTO_INCREMENT,
    `gh_id` INT NOT NULL,
    `username` varchar(255) NOT NULL,
    `avatar_url` varchar(255) NOT NULL,
    `name` varchar(255) DEFAULT NULL,
    `company` varchar(255) DEFAULT NULL,
    `blog` varchar(255) DEFAULT NULL,
    `location` varchar(255) DEFAULT NULL,
    `email` varchar(255) DEFAULT NULL,
    `bio` varchar(255) DEFAULT NULL,
    `twitter_username` varchar(255) DEFAULT NULL,
    `public_repos` INT NOT NULL,
    `public_gists` INT NOT NULL,
    `followers` INT NOT NULL,
    `following` INT NOT NULL,
    `created_at` datetime NOT NULL,
    `updated_at` datetime NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE trending_developers (
    `id` INT NOT NULL AUTO_INCREMENT,
    `username` varchar(255) NOT NULL,
    `language` varchar(255) DEFAULT NULL,
    `rank` INT NOT NULL,
    `scraped_at` datetime NOT NULL,
    `trend_date` date NOT NULL,
    `developer_id` INT DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `IDX_2F8A6Z8D8545CDA5` (`developer_id`),
    CONSTRAINT `FK_EZQFXDTYABCENQFI` FOREIGN KEY (`developer_id`) REFERENCES `developers` (`id`),
    UNIQUE (`username`, `language`, `trend_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;