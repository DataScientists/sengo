package profilerepository

import (
	"context"
	"sheng-go-backend/ent"
	"sheng-go-backend/ent/profile"
	"sheng-go-backend/pkg/entity/model"
	"sort"
	"strings"
)

func (r *profileRepository) GroupByTitle(
	ctx context.Context,
	searchTerm *string,
	minCount int,
) ([]*model.ProfileTitleGroup, error) {
	query := r.client.Profile.Query()

	// Apply search filter
	if searchTerm != nil && *searchTerm != "" {
		query = query.Where(
			profile.Or(
				profile.NameContainsFold(*searchTerm),
				profile.TitleContainsFold(*searchTerm),
			),
		)
	}

	// Group by title at database level
	var results []struct {
		Title string `json:"title"`
		Count int    `json:"count"`
	}

	err := query.
		GroupBy(profile.FieldTitle).
		Aggregate(ent.Count()).
		Scan(ctx, &results)
	if err != nil {
		if ent.IsNotFound(err) {
			return []*model.ProfileTitleGroup{}, nil
		}
		return nil, model.NewDBError(err)
	}

	// Post-process for case-insensitive grouping and filtering
	titleMap := make(map[string]*model.ProfileTitleGroup)
	for _, result := range results {
		lowerTitle := strings.ToLower(result.Title)

		if existing, ok := titleMap[lowerTitle]; ok {
			existing.Count += result.Count
		} else {
			titleMap[lowerTitle] = &model.ProfileTitleGroup{
				Title: capitalizeTitle(lowerTitle),
				Count: result.Count,
			}
		}
	}

	// Convert to slice and filter by minCount
	var groups []*model.ProfileTitleGroup
	for _, group := range titleMap {
		if group.Count >= minCount {
			groups = append(groups, group)
		}
	}

	// Sort by count descending
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Count > groups[j].Count
	})

	return groups, nil
}

// capitalizeTitle converts a lowercase title to Title Case
func capitalizeTitle(title string) string {
	words := strings.Fields(title)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}
	return strings.Join(words, " ")
}
