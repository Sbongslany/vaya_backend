-- +goose Up
CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    receiver_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TYPE call_status AS ENUM ('INITIATED', 'RINGING', 'CONNECTED', 'ENDED', 'MISSED', 'DECLINED');

CREATE TABLE call_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    caller_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    receiver_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    status call_status NOT NULL DEFAULT 'INITIATED',
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    duration_seconds INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chat_messages_trip ON chat_messages(trip_id);
CREATE INDEX idx_chat_messages_receiver ON chat_messages(receiver_id, read_at);
CREATE INDEX idx_call_sessions_trip ON call_sessions(trip_id);

-- +goose Down
DROP TABLE IF EXISTS call_sessions;
DROP TYPE IF EXISTS call_status;
DROP TABLE IF EXISTS chat_messages;