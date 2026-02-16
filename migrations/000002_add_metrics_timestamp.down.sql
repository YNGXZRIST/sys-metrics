DROP TRIGGER IF EXISTS update_metrics_updated_at ON metrics;

DROP FUNCTION IF EXISTS update_updated_at_column();

ALTER TABLE metrics
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS created_at;
