CREATE EXTENSION IF NOT EXISTS postgis;

-- Main measurements table (raw ingestion)
CREATE TABLE IF NOT EXISTS measurements (
    id          SERIAL PRIMARY KEY,
    latitude    DOUBLE PRECISION NOT NULL,
    longitude   DOUBLE PRECISION NOT NULL,
    parameter   TEXT             NOT NULL,
    value       DOUBLE PRECISION NOT NULL,
    time        TIMESTAMPTZ      NOT NULL DEFAULT now()
);

-- Add PostGIS geography column (Point, WGS‑84)
ALTER TABLE measurements
    ADD COLUMN IF NOT EXISTS geom geography(Point, 4326);

-- Back‑fill existing rows
UPDATE measurements
    SET geom = ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)
    WHERE geom IS NULL;

-- Spatial index for fast geo queries
CREATE INDEX IF NOT EXISTS idx_measurements_geom
    ON measurements USING GIST (geom);

-- Convert to Timescale hypertable (creates chunks by time)
SELECT create_hypertable('measurements', 'time', if_not_exists => TRUE);
