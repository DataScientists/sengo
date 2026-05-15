package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"sheng-go-backend/config"
	"sheng-go-backend/ent"
	"sheng-go-backend/ent/profilepost"
	"sheng-go-backend/ent/schema/ulid"
	"sheng-go-backend/pkg/adapter/repository/apiquotatrackerrepository"
	"sheng-go-backend/pkg/infrastructure/email"
	"sheng-go-backend/pkg/infrastructure/external/rapidapi"
	"sheng-go-backend/pkg/infrastructure/storage"
	"sheng-go-backend/pkg/usecase/usecase/apiquota"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

const (
	classificationsFile = "/Users/akbar/Documents/top50k_classifications.json"
	s3Prefix            = "linkedin-posts"
	maxPostsPerProfile  = 5
	maxRateLimitRetries = 3
	progressLogInterval = 100
)

type classificationRecord struct {
	ProfileURL      string `json:"profile_url"`
	TopType         string `json:"top_type"`
	TopScore        string `json:"top_score"`
	Confidence      string `json:"confidence"`
	MatchedKeywords string `json:"matched_keywords"`
}

func main() {
	dbURL := flag.String(
		"db",
		"postgresql://root:root@localhost:5433/sheng_db_new?sslmode=disable",
		"Database connection string",
	)
	flag.Parse()

	config.ReadConfig(config.ReadConfigOption{})

	dbClient, err := ent.Open("postgres", *dbURL)
	if err != nil {
		log.Fatalf("failed to open db connection: %v", err)
	}
	defer dbClient.Close()

	s3Svc, err := storage.NewS3Service()
	if err != nil {
		log.Fatalf("failed to initialize S3 service: %v", err)
	}

	emailSvc := email.NewEmailService()
	linkedinClient := rapidapi.NewLinkedInClient()
	quotaTrackerRepo := apiquotatrackerrepository.NewAPIQuotaTrackerRepository(dbClient)
	quotaManager := apiquota.NewQuotaManager(quotaTrackerRepo, emailSvc)

	records, err := loadClassifications(classificationsFile)
	if err != nil {
		log.Fatalf("failed to load classifications: %v", err)
	}
	log.Printf("Loaded %d profiles from %s", len(records), classificationsFile)

	ctx := context.Background()
	startTime := time.Now()

	var (
		totalProcessed int
		completed      int
		skipped        int
		failed         int
		notFound       int
		apiCallsMade   int
	)

	for i, rec := range records {
		username := usernameFromURL(rec.ProfileURL)
		if username == "" {
			log.Printf(
				"[%d/%d] Skipping blank username for URL: %s",
				i+1,
				len(records),
				rec.ProfileURL,
			)
			skipped++
			continue
		}

		totalProcessed++

		// Check if already completed — skip
		existing, err := dbClient.ProfilePost.
			Query().
			Where(profilepost.ProfileUsernameEQ(username)).
			Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			log.Printf("[%d/%d] DB error looking up %s: %v", i+1, len(records), username, err)
			failed++
			continue
		}
		if existing != nil && existing.FetchStatus == profilepost.FetchStatusCOMPLETED {
			skipped++
			if (i+1)%progressLogInterval == 0 {
				log.Printf(
					"[%d/%d] progress: completed=%d skipped=%d failed=%d notFound=%d apiCalls=%d",
					i+1,
					len(records),
					completed,
					skipped,
					failed,
					notFound,
					apiCallsMade,
				)
			}
			continue
		}

		// Check quota
		_, err = quotaManager.CheckAndReserveQuota(ctx, 1)
		if err != nil {
			log.Printf("Quota exceeded or unavailable: %v. Stopping.", err)
			break
		}

		// Upsert ProfilePost row as PENDING (create if not exists)
		ppRow, err := upsertProfilePost(ctx, dbClient, username, existing)
		if err != nil {
			log.Printf(
				"[%d/%d] Failed to upsert ProfilePost for %s: %v",
				i+1,
				len(records),
				username,
				err,
			)
			failed++
			continue
		}

		// Fetch posts with retry on rate limit
		posts, rawBody, fetchErr := fetchWithRetry(ctx, linkedinClient, username)
		apiCallsMade++

		if fetchErr != nil {
			var rateLimitErr *rapidapi.RateLimitError
			var notFoundErr *rapidapi.NotFoundError

			switch {
			case errors.As(fetchErr, &notFoundErr):
				notFound++
				_ = updateProfilePostStatus(
					ctx,
					dbClient,
					ppRow.ID,
					profilepost.FetchStatusNotFound,
					nil,
					strPtr(fetchErr.Error()),
				)
				log.Printf("[%d/%d] Not found: %s", i+1, len(records), username)

			case errors.As(fetchErr, &rateLimitErr):
				// Rate limit exhausted after retries — stop the run
				log.Printf("Rate limit exhausted after retries. Stopping.")
				_ = updateProfilePostStatus(
					ctx,
					dbClient,
					ppRow.ID,
					profilepost.FetchStatusFAILED,
					nil,
					strPtr(fetchErr.Error()),
				)
				failed++
				break

			default:
				failed++
				_ = updateProfilePostStatus(
					ctx,
					dbClient,
					ppRow.ID,
					profilepost.FetchStatusFAILED,
					nil,
					strPtr(fetchErr.Error()),
				)
				log.Printf("[%d/%d] Fetch error for %s: %v", i+1, len(records), username, fetchErr)
			}

			if (i+1)%progressLogInterval == 0 {
				logProgress(i+1, len(records), completed, skipped, failed, notFound, apiCallsMade)
			}
			continue
		}

		// Limit to maxPostsPerProfile
		if len(posts) > maxPostsPerProfile {
			posts = posts[:maxPostsPerProfile]
		}

		// Save raw JSON to S3 — batch folders of 900, matching profiles/ convention
		folder := completed / 900
		s3Key := fmt.Sprintf("%s/batch-%d/%s.json", s3Prefix, folder, username)
		if err := s3Svc.UploadJSON(ctx, s3Key, rawBody); err != nil {
			log.Printf("[%d/%d] S3 upload failed for %s: %v", i+1, len(records), username, err)
		}

		// Insert ProfilePostItem rows
		for _, post := range posts {
			if err := insertPostItem(ctx, dbClient, ppRow, username, post); err != nil {
				log.Printf(
					"[%d/%d] Failed to insert post item for %s: %v",
					i+1,
					len(records),
					username,
					err,
				)
			}
		}

		// Mark ProfilePost as COMPLETED
		if err := updateProfilePostStatus(ctx, dbClient, ppRow.ID, profilepost.FetchStatusCOMPLETED, &s3Key, nil); err != nil {
			log.Printf(
				"[%d/%d] Failed to update status for %s: %v",
				i+1,
				len(records),
				username,
				err,
			)
		}

		// Increment quota counter
		if err := quotaManager.IncrementCallCount(ctx, 1); err != nil {
			log.Printf("Failed to increment quota count: %v", err)
		}

		completed++

		if (i+1)%progressLogInterval == 0 {
			logProgress(i+1, len(records), completed, skipped, failed, notFound, apiCallsMade)
		}
	}

	elapsed := time.Since(startTime)
	fmt.Printf("\n=== Fetch Profile Posts Complete ===\n")
	fmt.Printf("Duration:       %v\n", elapsed.Round(time.Second))
	fmt.Printf("Total profiles: %d\n", len(records))
	fmt.Printf("Processed:      %d\n", totalProcessed)
	fmt.Printf("Completed:      %d\n", completed)
	fmt.Printf("Skipped:        %d\n", skipped)
	fmt.Printf("Failed:         %d\n", failed)
	fmt.Printf("Not found:      %d\n", notFound)
	fmt.Printf("API calls made: %d\n", apiCallsMade)
}

func loadClassifications(path string) ([]classificationRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	// Strip UTF-8 BOM if present
	bom := make([]byte, 3)
	n, _ := f.Read(bom)
	if n == 3 && bom[0] == 0xEF && bom[1] == 0xBB && bom[2] == 0xBF {
		// BOM present, continue reading from current position
	} else {
		// No BOM — seek back to start
		f.Seek(0, 0)
	}

	var records []classificationRecord
	if err := json.NewDecoder(f).Decode(&records); err != nil {
		return nil, fmt.Errorf("decode JSON: %w", err)
	}
	return records, nil
}

func usernameFromURL(profileURL string) string {
	// "linkedin.com/in/some-username" -> "some-username"
	u := strings.TrimPrefix(profileURL, "https://")
	u = strings.TrimPrefix(u, "http://")
	u = strings.TrimPrefix(u, "www.")
	u = strings.TrimPrefix(u, "linkedin.com/in/")
	u = strings.TrimSuffix(u, "/")
	return strings.TrimSpace(u)
}

func upsertProfilePost(
	ctx context.Context,
	client *ent.Client,
	username string,
	existing *ent.ProfilePost,
) (*ent.ProfilePost, error) {
	if existing != nil {
		// Reset to PENDING for re-fetch attempt
		return client.ProfilePost.
			UpdateOneID(existing.ID).
			SetFetchStatus(profilepost.FetchStatusPending).
			ClearErrorMessage().
			Save(ctx)
	}
	return client.ProfilePost.
		Create().
		SetProfileUsername(username).
		SetFetchStatus(profilepost.FetchStatusPending).
		Save(ctx)
}

func updateProfilePostStatus(
	ctx context.Context,
	client *ent.Client,
	id ulid.ID,
	status profilepost.FetchStatus,
	s3Key *string,
	errMsg *string,
) error {
	upd := client.ProfilePost.UpdateOneID(id).SetFetchStatus(status)
	if s3Key != nil {
		upd = upd.SetS3Key(*s3Key)
	}
	if errMsg != nil {
		upd = upd.SetErrorMessage(*errMsg)
	} else {
		upd = upd.ClearErrorMessage()
	}
	_, saveErr := upd.Save(ctx)
	return saveErr
}

func insertPostItem(
	ctx context.Context,
	client *ent.Client,
	ppRow *ent.ProfilePost,
	username string,
	post map[string]interface{},
) error {
	b := client.ProfilePostItem.
		Create().
		SetProfileUsername(username).
		SetProfilePost(ppRow).
		SetRawData(post)

	if v, ok := stringField(post, "urn"); ok {
		b = b.SetPostUrn(v)
	}
	if v, ok := stringField(post, "postUrl"); ok {
		b = b.SetPostURL(v)
	}
	if v, ok := stringField(post, "text"); ok {
		b = b.SetText(v)
	}
	if v, ok := stringField(post, "contentType"); ok {
		b = b.SetContentType(v)
	}
	if v, ok := boolField(post, "reposted"); ok {
		b = b.SetIsRepost(v)
	}
	if v, ok := intField(post, "totalReactionCount"); ok {
		b = b.SetTotalReactions(v)
	}
	if v, ok := intField(post, "likeCount"); ok {
		b = b.SetLikeCount(v)
	}
	if v, ok := intField(post, "commentsCount"); ok {
		b = b.SetCommentsCount(v)
	}
	if v, ok := intField(post, "repostsCount"); ok {
		b = b.SetRepostsCount(v)
	}
	if v, ok := intField(post, "empathyCount"); ok {
		b = b.SetEmpathyCount(v)
	}
	if v, ok := intField(post, "praiseCount"); ok {
		b = b.SetPraiseCount(v)
	}
	if v, ok := intField(post, "funnyCount"); ok {
		b = b.SetFunnyCount(v)
	}
	if v, ok := intField(post, "InterestCount"); ok {
		b = b.SetInterestCount(v)
	}
	// postedDateTimestamp is ms unix — most reliable date field
	if v, ok := intField(post, "postedDateTimestamp"); ok && v > 0 {
		b = b.SetPostedAt(time.UnixMilli(int64(v)))
	}

	_, err := b.Save(ctx)
	return err
}

func boolField(m map[string]interface{}, key string) (bool, bool) {
	v, ok := m[key]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

// fetchWithRetry calls FetchProfilePosts with exponential backoff on rate limit.
func fetchWithRetry(
	ctx context.Context,
	client *rapidapi.LinkedInClient,
	username string,
) ([]map[string]interface{}, []byte, error) {
	backoff := time.Duration(config.C.RapidAPI.RateLimitBackoffMs) * time.Millisecond
	if backoff == 0 {
		backoff = 1 * time.Second
	}

	for attempt := 0; attempt < maxRateLimitRetries; attempt++ {
		posts, body, err := client.FetchProfilePosts(ctx, username, 0)
		if err == nil {
			return posts, body, nil
		}

		var rlErr *rapidapi.RateLimitError
		if !errors.As(err, &rlErr) {
			return nil, body, err
		}

		wait := rlErr.RetryAfter
		if wait <= 0 {
			wait = backoff * time.Duration(1<<attempt)
		}
		maxWait := time.Duration(config.C.RapidAPI.RateLimitBackoffMaxMs) * time.Millisecond
		if maxWait > 0 && wait > maxWait {
			wait = maxWait
		}

		log.Printf(
			"Rate limited for %s, waiting %v (attempt %d/%d)",
			username,
			wait,
			attempt+1,
			maxRateLimitRetries,
		)
		time.Sleep(wait)
	}

	return nil, nil, &rapidapi.RateLimitError{Message: "max retries exceeded"}
}

func logProgress(current, total, completed, skipped, failed, notFound, apiCalls int) {
	log.Printf("[%d/%d] completed=%d skipped=%d failed=%d notFound=%d apiCalls=%d",
		current, total, completed, skipped, failed, notFound, apiCalls)
}

func strPtr(s string) *string { return &s }

func stringField(m map[string]interface{}, key string) (string, bool) {
	v, ok := m[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok && s != ""
}

func intField(m map[string]interface{}, key string) (int, bool) {
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	}
	return 0, false
}

func timeField(m map[string]interface{}, key string) (time.Time, bool) {
	v, ok := m[key]
	if !ok {
		return time.Time{}, false
	}
	switch t := v.(type) {
	case string:
		parsed, err := time.Parse(time.RFC3339, t)
		if err == nil {
			return parsed, true
		}
	case float64:
		// Unix timestamp in ms or seconds
		if t > 1e10 {
			return time.UnixMilli(int64(t)), true
		}
		return time.Unix(int64(t), 0), true
	}
	return time.Time{}, false
}
