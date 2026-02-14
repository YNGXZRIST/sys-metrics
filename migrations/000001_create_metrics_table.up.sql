CREATE TABLE IF NOT EXISTS metrics (
    id VARCHAR(255) NOT NULL PRIMARY KEY,
    mtype VARCHAR(50) NOT NULL CHECK (mtype IN ('counter', 'gauge')),
    delta BIGINT,
    value DOUBLE PRECISION,
    hash VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS idx_metrics_id ON metrics(id);
CREATE INDEX IF NOT EXISTS idx_metrics_mtype ON metrics(mtype);
