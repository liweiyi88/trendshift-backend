DROP TABLE repository_monthly_insights;

ALTER TABLE repositories
DROP COLUMN skipped;

ALTER TABLE developers
DROP COLUMN skipped;