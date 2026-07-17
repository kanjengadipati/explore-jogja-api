CREATE TABLE trips (
    id          SERIAL PRIMARY KEY,
    external_id VARCHAR(100)  UNIQUE NOT NULL,
    user_id     INTEGER       NOT NULL,
    title       VARCHAR(255)  NOT NULL DEFAULT 'My Trip',
    start_date  DATE,
    end_date    DATE,
    duration_days INTEGER     NOT NULL DEFAULT 1,
    days        JSONB         NOT NULL DEFAULT '[]',
    notes       TEXT          DEFAULT '',
    status      VARCHAR(50)   NOT NULL DEFAULT 'draft',
    created_at  TIMESTAMP     DEFAULT NOW(),
    updated_at  TIMESTAMP     DEFAULT NOW(),
    deleted_at  TIMESTAMP
);

CREATE INDEX idx_trips_external_id ON trips(external_id);
CREATE INDEX idx_trips_user_id     ON trips(user_id);
CREATE INDEX idx_trips_status      ON trips(status);
CREATE INDEX idx_trips_deleted_at  ON trips(deleted_at);
