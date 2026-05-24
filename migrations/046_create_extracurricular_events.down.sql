-- Rollback v0.162.0 #319 — drop extracurricular events module schema.
-- Participants first due к FK к events table.
DROP TABLE IF EXISTS extracurricular_participants;
DROP TABLE IF EXISTS extracurricular_events;
