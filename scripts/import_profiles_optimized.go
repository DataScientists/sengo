package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"sheng-go-backend/ent"
	"sheng-go-backend/ent/schema/ulid"
	"sheng-go-backend/pkg/const/globalid"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func main() {
	client, err := ent.Open(
		"postgres",
		"postgresql://root:root@localhost:5433/sheng_db?sslmode=disable",
	)
	if err != nil {
		log.Fatalf("failed connecting to database: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	if err := importProfilesFromFile(ctx, client, "profiles.sql"); err != nil {
		log.Fatalf("failed importing profiles: %v", err)
	}

	fmt.Println("Successfully imported all profiles with ULIDs")
}

func importProfilesFromFile(ctx context.Context, client *ent.Client, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024) // 10MB buffer for large lines

	insertPattern := regexp.MustCompile(
		`^\s*\((\d+),\s*'([^']*)',\s*'([^']*)',\s*'([^']*)',\s*'([^']*)'\),?\s*$`,
	)

	totalImported := 0
	batchSize := 1000
	batch := make([]*ent.ProfileCreate, 0, batchSize)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip non-data lines
		if !strings.HasPrefix(line, "(") {
			continue
		}

		matches := insertPattern.FindStringSubmatch(line)
		if len(matches) != 6 {
			// Try alternative pattern for lines ending with semicolon
			altPattern := regexp.MustCompile(
				`^\s*\((\d+),\s*'([^']*)',\s*'([^']*)',\s*'([^']*)',\s*'([^']*)'\);?\s*$`,
			)
			matches = altPattern.FindStringSubmatch(line)
			if len(matches) != 6 {
				continue
			}
		}

		// Generate new ULID
		newID := ulid.MustNew(globalid.New().Profile.Prefix)

		name := strings.ReplaceAll(matches[2], "\\'", "'")
		title := strings.ReplaceAll(matches[3], "\\'", "'")
		urn := strings.ReplaceAll(matches[4], "\\'", "'")
		sourceFile := strings.ReplaceAll(matches[5], "\\'", "'")

		// Add to batch
		profileCreate := client.Profile.Create().
			SetID(newID).
			SetName(name).
			SetTitle(title).
			SetUrn(urn).
			SetSourceFile(sourceFile)

		batch = append(batch, profileCreate)

		// Execute batch when full
		if len(batch) >= batchSize {
			if err := executeBatch(ctx, batch); err != nil {
				log.Printf("Failed to execute batch: %v", err)
			} else {
				totalImported += len(batch)
				fmt.Printf("Imported %d profiles (total: %d)\n", len(batch), totalImported)
			}
			batch = batch[:0] // Reset batch
		}
	}

	// Execute remaining batch
	if len(batch) > 0 {
		if err := executeBatch(ctx, batch); err != nil {
			log.Printf("Failed to execute final batch: %v", err)
		} else {
			totalImported += len(batch)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	fmt.Printf("Total imported: %d profiles\n", totalImported)
	return nil
}

func executeBatch(ctx context.Context, batch []*ent.ProfileCreate) error {
	for _, create := range batch {
		_, err := create.Save(ctx)
		if err != nil {
			// Log error but continue with other records
			log.Printf("Failed to create profile: %v", err)
		}
	}
	return nil
}

