CREATE EXTENSION IF NOT EXISTS postgis;

-- Main measurements table (raw ingestion)
CREATE TABLE IF NOT EXISTS measurements (
    id          SERIAL PRIMARY KEY,
    parameter   TEXT             NOT NULL,
    value       DOUBLE PRECISION NOT NULL,
    time        TIMESTAMPTZ      NOT NULL DEFAULT now()
);
-- Table to store detected anomalies
CREATE TABLE IF NOT EXISTS anomalies (
    id          SERIAL PRIMARY KEY,
    parameter   TEXT             NOT NULL,
    value       DOUBLE PRECISION NOT NULL,
    time        TIMESTAMPTZ      NOT NULL DEFAULT now(),
    location    geography(Point, 4326),
    description TEXT             NOT NULL
);

-- Add PostGIS geography column (Point, WGS‑84)
ALTER TABLE measurements
    ADD COLUMN IF NOT EXISTS location geography(Point, 4326);

-- Spatial index for fast geo queries
CREATE INDEX IF NOT EXISTS idx_measurements_geom
    ON measurements USING GIST (location);

-- Convert to Timescale hypertable (creates chunks by time)
SELECT create_hypertable('measurements', 'time', if_not_exists => TRUE);


-- Spatial index for anomalies table
CREATE INDEX IF NOT EXISTS idx_anomalies_geom
    ON anomalies USING GIST (location);
