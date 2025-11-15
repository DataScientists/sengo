package email

import (
	"fmt"
	"sheng-go-backend/config"
	"time"

	"gopkg.in/mail.v2"
)

// EmailService handles all email operations
type EmailService struct {
	dialer      *mail.Dialer
	fromAddress string
	adminEmail  string
}

// NewEmailService creates a new email service instance
func NewEmailService() *EmailService {
	cfg := config.C.Email

	dialer := mail.NewDialer(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.Username,
		cfg.Password,
	)

	return &EmailService{
		dialer:      dialer,
		fromAddress: cfg.FromAddress,
		adminEmail:  cfg.AdminEmail,
	}
}

// SendQuotaExceededAlert sends an email when API quota is exceeded
func (s *EmailService) SendQuotaExceededAlert(callCount, quotaLimit int, month, year int) error {
	subject := fmt.Sprintf("⚠️ RapidAPI Monthly Quota Exceeded (%d/%d)", callCount, quotaLimit)

	body := fmt.Sprintf(`
<html>
<body>
<h2>API Quota Exceeded</h2>
<p>Hi Admin,</p>

<p>Your RapidAPI quota has been exhausted:</p>
<ul>
  <li><strong>Current Calls:</strong> %d / %d</li>
  <li><strong>Month:</strong> %d/%d</li>
</ul>

<p>The profile fetching cron job has been paused and will automatically resume on the 1st of next month.</p>

<p>To override and continue fetching profiles immediately, visit your dashboard and enable the quota override.</p>

<p>Best regards,<br/>Sheng System</p>
</body>
</html>
	`, callCount, quotaLimit, month, year)

	return s.sendHTML(s.adminEmail, subject, body)
}

// SendJobCompletionSummary sends a summary email after job execution
func (s *EmailService) SendJobCompletionSummary(
	jobName string,
	durationSeconds int,
	totalProcessed, successful, failed int,
	apiCallsMade, quotaRemaining int,
	errors []string,
	nextRunTime time.Time,
) error {
	subject := fmt.Sprintf("✅ %s Job Completed", jobName)

	errorList := ""
	if len(errors) > 0 {
		errorList = "<h3>Failed Entries:</h3><ul>"
		for _, err := range errors {
			errorList += fmt.Sprintf("<li>%s</li>", err)
		}
		errorList += "</ul>"
	}

	body := fmt.Sprintf(`
<html>
<body>
<h2>Job Summary</h2>
<ul>
  <li><strong>Job Name:</strong> %s</li>
  <li><strong>Duration:</strong> %d seconds</li>
  <li><strong>Total Processed:</strong> %d</li>
  <li><strong>Successful:</strong> %d</li>
  <li><strong>Failed:</strong> %d</li>
  <li><strong>API Calls Made:</strong> %d</li>
  <li><strong>Quota Remaining:</strong> %d / 50,000</li>
</ul>

%s

<p><strong>Next scheduled run:</strong> %s</p>

<p>Best regards,<br/>Sheng System</p>
</body>
</html>
	`, jobName, durationSeconds, totalProcessed, successful, failed, apiCallsMade, quotaRemaining, errorList, nextRunTime.Format("Jan 02, 2006 at 15:04 MST"))

	return s.sendHTML(s.adminEmail, subject, body)
}

// SendQuotaResetNotification sends an email when quota is reset
func (s *EmailService) SendQuotaResetNotification(month, year int, quotaLimit int) error {
	subject := "🔄 Monthly API Quota Reset"

	body := fmt.Sprintf(`
<html>
<body>
<h2>Monthly API Quota Reset</h2>
<p>Hi Admin,</p>

<p>Your RapidAPI quota has been reset for the new month:</p>
<ul>
  <li><strong>Month:</strong> %d/%d</li>
  <li><strong>Available Calls:</strong> %d / %d</li>
  <li><strong>Cron Job Status:</strong> Resumed</li>
</ul>

<p>Profile fetching will resume automatically.</p>

<p>Best regards,<br/>Sheng System</p>
</body>
</html>
	`, month, year, quotaLimit, quotaLimit)

	return s.sendHTML(s.adminEmail, subject, body)
}

// SendQuotaOverrideAlert sends an email when admin enables quota override
func (s *EmailService) SendQuotaOverrideAlert(enabled bool) error {
	subject := "🚨 API Quota Override Status Changed"
	status := "enabled"
	if !enabled {
		status = "disabled"
	}

	body := fmt.Sprintf(`
<html>
<body>
<h2>Quota Override Status Changed</h2>
<p>Hi Admin,</p>

<p>The API quota override has been <strong>%s</strong>.</p>

<p>Profile fetching will now %s regardless of the quota limit.</p>

<p>Best regards,<br/>Sheng System</p>
</body>
</html>
	`, status, map[bool]string{true: "continue", false: "respect quota limits"}[enabled])

	return s.sendHTML(s.adminEmail, subject, body)
}

// sendHTML sends an HTML email
func (s *EmailService) sendHTML(to, subject, htmlBody string) error {
	m := mail.NewMessage()
	m.SetHeader("From", s.fromAddress)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	if err := s.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
