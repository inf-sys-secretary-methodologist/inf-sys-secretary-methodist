-- Migration: Create messaging schema
-- Description: Tables for internal messaging system (conversations, messages, participants)

-- Conversations table
CREATE TABLE IF NOT EXISTS conversations (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL DEFAULT 'direct' CHECK (type IN ('direct', 'group')),
    title VARCHAR(255),
    description TEXT,
    avatar_url VARCHAR(500),
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for finding conversations by creator
CREATE INDEX idx_conversations_created_by ON conversations(created_by);
CREATE INDEX idx_conversations_type ON conversations(type);
CREATE INDEX idx_conversations_updated_at ON conversations(updated_at DESC);

-- Conversation participants table
CREATE TABLE IF NOT EXISTS conversation_participants (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'member' CHECK (role IN ('member', 'admin')),
    last_read_at TIMESTAMPTZ,
    last_read_message_id BIGINT,
    is_muted BOOLEAN NOT NULL DEFAULT FALSE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at TIMESTAMPTZ,
    UNIQUE(conversation_id, user_id)
);

-- Indexes for participants
CREATE INDEX idx_conv_participants_user ON conversation_participants(user_id) WHERE left_at IS NULL;
CREATE INDEX idx_conv_participants_conversation ON conversation_participants(conversation_id) WHERE left_at IS NULL;

-- Messages table
CREATE TABLE IF NOT EXISTS messages (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'text' CHECK (type IN ('text', 'image', 'file', 'system')),
    content TEXT NOT NULL,
    reply_to_id BIGINT REFERENCES messages(id) ON DELETE SET NULL,
    is_edited BOOLEAN NOT NULL DEFAULT FALSE,
    edited_at TIMESTAMPTZ,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for messages
CREATE INDEX idx_messages_conversation ON messages(conversation_id, created_at DESC);
CREATE INDEX idx_messages_sender ON messages(sender_id);
CREATE INDEX idx_messages_reply_to ON messages(reply_to_id) WHERE reply_to_id IS NOT NULL;

-- Full-text search for messages
ALTER TABLE messages ADD COLUMN IF NOT EXISTS search_vector tsvector
    GENERATED ALWAYS AS (
        setweight(to_tsvector('russian', coalesce(content, '')), 'A')
    ) STORED;

CREATE INDEX idx_messages_search ON messages USING GIN(search_vector);

-- Message attachments table
CREATE TABLE IF NOT EXISTS message_attachments (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    file_id BIGINT NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL DEFAULT 0,
    mime_type VARCHAR(100) NOT NULL,
    url VARCHAR(500) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_message_attachments_message ON message_attachments(message_id);

-- Function to update conversation updated_at on new message
CREATE OR REPLACE FUNCTION update_conversation_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE conversations
    SET updated_at = NOW()
    WHERE id = NEW.conversation_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update conversation timestamp
DROP TRIGGER IF EXISTS trigger_update_conversation_timestamp ON messages;
CREATE TRIGGER trigger_update_conversation_timestamp
    AFTER INSERT ON messages
    FOR EACH ROW
    EXECUTE FUNCTION update_conversation_timestamp();

-- Unique index for direct conversations (prevent duplicates)
-- We create a function and index to ensure only one direct conversation between any two users
CREATE OR REPLACE FUNCTION get_direct_conversation_key(user1 BIGINT, user2 BIGINT)
RETURNS TEXT AS $$
BEGIN
    IF user1 < user2 THEN
        RETURN user1::TEXT || '-' || user2::TEXT;
    ELSE
        RETURN user2::TEXT || '-' || user1::TEXT;
    END IF;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Comments
COMMENT ON TABLE conversations IS 'Chat conversations (direct or group)';
COMMENT ON TABLE conversation_participants IS 'Participants in conversations';
COMMENT ON TABLE messages IS 'Chat messages';
COMMENT ON TABLE message_attachments IS 'File attachments for messages';
COMMENT ON COLUMN conversations.type IS 'Type of conversation: direct (1:1) or group';
COMMENT ON COLUMN conversation_participants.role IS 'Participant role: member or admin';
COMMENT ON COLUMN conversation_participants.is_muted IS 'Whether notifications are muted for this participant';
COMMENT ON COLUMN messages.type IS 'Message type: text, image, file, or system';
COMMENT ON COLUMN messages.is_edited IS 'Whether the message has been edited';
COMMENT ON COLUMN messages.is_deleted IS 'Soft delete flag for messages';
