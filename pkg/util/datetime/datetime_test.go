package datetime_test

import (
	"sheng-go-backend/pkg/util/datetime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateTime(t *testing.T) {
	tests := []struct {
		name    string
		arrange func() time.Time
		act     func(time time.Time) string
		assert  func(t *testing.T, time string)
	}{{
		name: "Should format time with time zone",
		arrange: func() time.Time {
			bhutanTimeZone := time.FixedZone("BTT", 6*3600)
			return time.Date(2024, time.February, 11, 10, 30, 40, 40, bhutanTimeZone)
		},
		act: func(time time.Time) string {
			return datetime.FormatDate(time)
		},
		assert: func(t *testing.T, time string) {
			assert.Equal(t, time, "2024-02-11T10:30:40+06:00")
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			datetime := tt.arrange()
			formattedDateTime := tt.act(datetime)
			tt.assert(t, formattedDateTime)
		})
	}
}
