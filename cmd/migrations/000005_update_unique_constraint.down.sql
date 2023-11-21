ALTER TABLE trending_developers DROP CONSTRAINT `username`;
ALTER TABLE trending_developers ADD UNIQUE (`username`, `language`, `trend_date`);

ALTER TABLE trending_repositories DROP CONSTRAINT `full_name`;
ALTER TABLE trending_repositories ADD UNIQUE (`full_name`, `language`, `trend_date`);