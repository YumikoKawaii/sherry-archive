CREATE TABLE interest_sync_watermarks (
    identity_id    UUID        NOT NULL PRIMARY KEY,
    last_synced_at TIMESTAMPTZ NOT NULL
);
