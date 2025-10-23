DROP TABLE repository_monthly_insights;

ALTER TABLE repositories
DROP COLUMN repository_created_at,
DROP COLUMN skipped;

ALTER TABLE developers
DROP COLUMN skipped;