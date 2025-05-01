CREATE EXTENSION IF NOT EXISTS postgis;

-- Main measurements table (raw ingestion)
CREATE TABLE IF NOT EXISTS measurements (
    id          SERIAL PRIMARY KEY,
    parameter   TEXT             NOT NULL,
    value       DOUBLE PRECISION NOT NULL,
    time        TIMESTAMPTZ      NOT NULL DEFAULT now()
);

-- Add PostGIS geography column (Point, WGS‑84)
ALTER TABLE measurements
    ADD COLUMN IF NOT EXISTS location geography(Point, 4326);

-- Back‑fill existing rows
UPDATE measurements
    SET location = ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)
    WHERE location IS NULL;

-- Spatial index for fast geo queries
CREATE INDEX IF NOT EXISTS idx_measurements_geom
    ON measurements USING GIST (location);

-- Convert to Timescale hypertable (creates chunks by time)
SELECT create_hypertable('measurements', 'time', if_not_exists => TRUE);
