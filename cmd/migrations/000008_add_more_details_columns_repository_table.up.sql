ALTER TABLE repositories
ADD `number_of_contributors` INT DEFAULT NULL,
ADD `last_commit_at` DATETIME(3) DEFAULT NULL,
ADD `last_user_commit_at` DATETIME(3) DEFAULT NULL,
ADD `license_key` varchar(500) DEFAULT NULL,
ADD `license_name` varchar(500) DEFAULT NULL;

ALTER TABLE repositories
MODIFY COLUMN `description` TEXT DEFAULT NULL;