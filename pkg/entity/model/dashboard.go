package model

import "sheng-go-backend/ent"

// ProfileEntryStats represents statistics about profile entries
type ProfileEntryStats struct {
	TotalCount     int `json:"totalCount"`
	PendingCount   int `json:"pendingCount"`
	FetchingCount  int `json:"fetchingCount"`
	CompletedCount int `json:"completedCount"`
	FailedCount    int `json:"failedCount"`
}

// JobStats represents statistics about job executions
type JobStats struct {
	TotalExecutions    int     `json:"totalExecutions"`
	SuccessRate        float64 `json:"successRate"`
	AverageDuration    int     `json:"averageDuration"`
	TotalProfiles      int     `json:"totalProfiles"`
	TotalAPICallsMade  int     `json:"totalAPICallsMade"`
}

// DashboardOverview provides a complete dashboard overview
type DashboardOverview struct {
	QuotaStatus           *ent.APIQuotaTracker       `json:"quotaStatus"`
	PendingProfilesCount  int                        `json:"pendingProfilesCount"`
	RecentJobExecutions   []*ent.JobExecutionHistory `json:"recentJobExecutions"`
	CronJobsStatus        []*ent.CronJobConfig       `json:"cronJobsStatus"`
	ProfileEntryStats     *ProfileEntryStats         `json:"profileEntryStats"`
}
