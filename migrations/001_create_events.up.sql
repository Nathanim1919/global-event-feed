CREATE TABLE events (
    id BIGSERIAL PRIMARY KEY,
    type TEXT,
    title TEXT,
    source TEXT,
    description TEXT,
    occurred_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
)
