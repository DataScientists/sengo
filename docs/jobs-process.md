# Jobs Process Guide

## Scheduler & Job Registration
- On startup, `pkg/infrastructure/scheduler/cron.go` initializes default cron configs (DB table `cron_job_configs`) if missing, then registers enabled jobs from DB.
- Two jobs are registered:
  - `profile_fetcher` (type `PROFILE_FETCHER`) runs per `cron.profileFetcherSchedule` with batch size `cron.batchSize` (default 10) and `respect_quota=true`.
  - `quota_reset` resets monthly RapidAPI quota per `cron.quotaResetSchedule`.
- Each run updates `cron_job_configs.last_run` before executing.

## Profile Fetcher Flow (`pkg/usecase/usecase/profilefetcher/fetcher.go`)
1) Load job config (`profile_fetcher`) from DB.
2) Loop until done:
   - Call `QuotaManager.CheckAndReserveQuota(batchSize)`; returns allowed batch size (may be smaller if monthly quota nearly exhausted).
   - If quota check fails:
     - If nothing processed yet and `respect_quota` is true → record `QUOTA_EXCEEDED` history and stop.
     - If mid-run and `respect_quota` is true → stop loop, mark job `PARTIAL`, add error note.
     - If `respect_quota` is false → continue with requested batch size.
   - Fetch a batch of pending entries ordered by `created_at` (`GetPendingBatch(ctx, allowedBatchSize)`). If none, exit loop.
3) For each entry in the batch:
   - Mark status `FETCHING`.
   - Fetch from RapidAPI (`linkedinClient.FetchProfileByURN`) through `fetchProfileWithRetry`:
     - Retries on RapidAPI rate-limit (HTTP 429) using exponential backoff.
     - Defaults: `rateLimitMaxRetries=3`, `rateLimitBackoffMs=1000`, `rateLimitBackoffMaxMs=8000` (configurable via `rapidapi.*`).
     - Honors `Retry-After` header when present and caps with max backoff.
     - `APICallsMade` counts every attempt.
   - On success:
     - Increment quota call count (`QuotaManager.IncrementCallCount`).
     - Upload raw JSON to S3 (`profiles/<urn>-<ts>-raw.json`).
     - Extract and clean data (`extractProfileData`) → upload cleaned JSON to S3 (`profiles/<urn>-<ts>-cleaned.json`).
     - Upsert profile record in DB with S3 keys (`ProfileRepository.Upsert`).
     - Mark profile entry `COMPLETED`, bump fetch count, set `last_fetched_at`.
   - On failure (RapidAPI error, S3 error, DB error, exhausted retries):
     - Mark entry `FAILED` with error message.
     - Continue to next entry.
4) After loop:
   - Collect quota remaining.
   - Save `job_execution_history` with counts: `TotalProcessed`, `SuccessfulCount`, `FailedCount`, `APICallsMade`, `QuotaRemaining`, status (`SUCCESS`, `PARTIAL`, `FAILED`, or `QUOTA_EXCEEDED`), plus joined error summary.
   - Send completion email summary (counts + errors).

## Data Cleaning
- `extractProfileData` pulls selected fields from the RapidAPI profile: `urn`, `username`, `firstName`, `lastName`, `headline`, `geo`, `educations`, `fullPositions`, `skills`.
- Cleaned JSON is persisted separately from raw to S3; DB row stores S3 keys for both.

## Persistence
- Raw and cleaned payloads: uploaded via `storage.S3Service.UploadJSON`.
- Profile upsert: `ProfileRepository.Upsert` writes/updates core profile fields and S3 key references.
- Profile entry status updates: `ProfileEntryRepository.UpdateStatus` and `UpdateAfterFetch`.
- Job run history: `JobExecutionHistoryRepository.Create` for observability and audit.

## Failure & Retry Strategy
- Per-entry retries: only for RapidAPI rate limits (HTTP 429). Backoff is exponential, respects `Retry-After`, capped by config.
- After retries exhaust, the entry is marked `FAILED`; it is **not automatically retried in future runs** unless manually re-queued.
- Other errors (S3, DB, parse) fail the entry immediately.
- Quota handling is per batch: monthly quota check can halt the run mid-way (marks job `PARTIAL`) or before any work (marks `QUOTA_EXCEEDED`).

## S3 Upload Details
- Keys include URN and a timestamp for traceability and immutability:
  - Raw: `profiles/<urn>-<unix_ts>-raw.json`
  - Cleaned: `profiles/<urn>-<unix_ts>-cleaned.json`
- Failures to upload or read are captured per-entry and recorded in the job error summary.

## Key Config Knobs (config/config.yml)
- `cron.profileFetcherSchedule`, `cron.batchSize`
- `rapidapi.monthlyQuota`, `rapidapi.timeoutSeconds`
- Rate-limit handling: `rapidapi.rateLimitMaxRetries`, `rapidapi.rateLimitBackoffMs`, `rapidapi.rateLimitBackoffMaxMs`
