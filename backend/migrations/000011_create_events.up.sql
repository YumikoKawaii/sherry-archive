CREATE TABLE events (
    device_id  UUID        NOT NULL,
    user_id    UUID        NULL,
    event      TEXT        NOT NULL,
    properties JSONB       NOT NULL DEFAULT '{}',
    referrer   TEXT        NOT NULL DEFAULT '',
    ip_hash    TEXT        NOT NULL DEFAULT '',
    user_agent TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- BRIN is tiny and fast for append-only time-series
CREATE INDEX idx_events_created_brin ON events USING BRIN (created_at);
CREATE INDEX idx_events_device       ON events (device_id, created_at DESC);
CREATE INDEX idx_events_user         ON events (user_id,   created_at DESC) WHERE user_id IS NOT NULL;
CREATE INDEX idx_events_type         ON events (event,     created_at DESC);
