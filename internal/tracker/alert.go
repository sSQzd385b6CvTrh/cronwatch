package tracker

import (
	"fmt"
	"time"
)

// DriftAlert is raised when a job runs significantly earlier or later
// than its scheduled fire time.
type DriftAlert struct {
	Job      string
	Expected time.Time
	Actual   time.Time
	// Drift is positive when the job ran late, negative when early.
	Drift time.Duration
}

func (a DriftAlert) String() string {
	direction := "late"
	d := a.Drift
	if d < 0 {
		direction = "early"
		d = -d
	}
	return fmt.Sprintf(
		"[DRIFT] job=%q expected=%s actual=%s drift=%s (%s)",
		a.Job,
		a.Expected.Format(time.RFC3339),
		a.Actual.Format(time.RFC3339),
		d.Round(time.Second),
		direction,
	)
}

// MissedAlert is raised when a job's expected fire time has passed
// without a recorded execution.
type MissedAlert struct {
	Job      string
	Expected time.Time
	MissedAt time.Time
}

func (a MissedAlert) String() string {
	return fmt.Sprintf(
		"[MISSED] job=%q expected=%s detected-at=%s overdue=%s",
		a.Job,
		a.Expected.Format(time.RFC3339),
		a.MissedAt.Format(time.RFC3339),
		a.MissedAt.Sub(a.Expected).Round(time.Second),
	)
}
