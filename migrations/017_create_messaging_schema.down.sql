-- Migration: Drop messaging schema
-- Description: Rollback for messaging tables

-- Drop triggers and functions
DROP TRIGGER IF EXISTS trigger_update_conversation_timestamp ON messages;
DROP FUNCTION IF EXISTS update_conversation_timestamp();
DROP FUNCTION IF EXISTS get_direct_conversation_key(BIGINT, BIGINT);

-- Drop tables in order (respecting foreign keys)
DROP TABLE IF EXISTS message_attachments;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversation_participants;
DROP TABLE IF EXISTS conversations;
