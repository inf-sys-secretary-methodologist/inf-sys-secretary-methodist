-- Reset all AI index statuses to 'pending' so documents get reindexed
-- with text extraction from S3 files (previously content was always NULL).
UPDATE ai_document_index_status SET status = 'pending', chunks_count = 0, error_message = NULL;
