ALTER TABLE repositories
ADD `description` varchar(1000) DEFAULT NULL,
ADD `default_branch` varchar(255) DEFAULT NULL;