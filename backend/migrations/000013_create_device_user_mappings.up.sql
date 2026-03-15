CREATE TABLE device_user_mappings (
    device_id  UUID NOT NULL,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (device_id, user_id)
);

CREATE INDEX idx_device_user_mappings_device_id ON device_user_mappings(device_id);
