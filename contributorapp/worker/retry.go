package worker

// TODO(retry):
// Implement retry worker to reprocess fail_records with error_type != validation
// - Fetch records with retry_count < max_retry
// - Rebuild ContributorRecord from raw_data
// - Retry upsert via ContributorAdapter
// - On success:
//   - delete fail_record
//   - increment job.success_count
//   - decrement job.fail_count
// - On failure:
//   - increment retry_count
//   - update last_error

// TODO(retry):
// Add repository methods for retry logic:
// - GetRetryable(limit, maxRetry)
// - IncrementRetry(id, lastError)
// - Delete(id)

// TODO(retry):
// After retry cycle, re-evaluate job status:
// - success if fail_count == 0
// - partial_success if success_count > 0 && fail_count > 0
// - failed otherwise

// TODO(retry):
// Decide retry execution strategy:
// - cron-based
// - background goroutine
// - separate queue / stream
