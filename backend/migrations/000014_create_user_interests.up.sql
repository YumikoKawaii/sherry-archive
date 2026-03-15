CREATE TABLE user_interests (
    identity_id UUID        NOT NULL,
    dimension   TEXT        NOT NULL,  -- e.g. "tag:action", "author:foo", "category:manga"
    score       FLOAT       NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (identity_id, dimension)
);

CREATE INDEX idx_user_interests_identity_id ON user_interests(identity_id);
