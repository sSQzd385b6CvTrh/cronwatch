package notifier

import (
	"fmt"
	"strings"
)

// SinkType identifies a supported notification backend.
type SinkType string

const (
	SinkLog       SinkType = "log"
	SinkWebhook   SinkType = "webhook"
	SinkEmail     SinkType = "email"
	SinkSlack     SinkType = "slack"
	SinkPagerDuty SinkType = "pagerduty"
	SinkOpsGenie  SinkType = "opsgenie"
	SinkVictorOps SinkType = "victorops"
	SinkSNS       SinkType = "sns"
	SinkDatadog   SinkType = "datadog"
)

// KnownSinkTypes lists every sink type recognised by cronwatch.
var KnownSinkTypes = []SinkType{
	SinkLog,
	SinkWebhook,
	SinkEmail,
	SinkSlack,
	SinkPagerDuty,
	SinkOpsGenie,
	SinkVictorOps,
	SinkSNS,
	SinkDatadog,
}

// ParseSinkType converts a raw string to a validated SinkType.
// Matching is case-insensitive, so "Slack" and "SLACK" are both accepted.
func ParseSinkType(s string) (SinkType, error) {
	t := SinkType(strings.ToLower(s))
	for _, k := range KnownSinkTypes {
		if t == k {
			return t, nil
		}
	}
	return "", fmt.Errorf("unknown sink type %q", s)
}
