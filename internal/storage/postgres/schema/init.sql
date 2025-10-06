CREATE TABLE IF NOT EXISTS urls (
    id          BIGSERIAL       PRIMARY KEY,
    long_url    VARCHAR(2048)   NOT NULL,
    slug        VARCHAR(10)     NOT NULL UNIQUE,
    clicks      BIGINT          DEFAULT 0,
    created_at  TIMESTAMPTZ     DEFAULT CURRENT_TIMESTAMP,
    expires_at  TIMESTAMPTZ     NOT NULL
);