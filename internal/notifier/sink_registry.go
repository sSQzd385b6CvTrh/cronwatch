package notifier

import (
	"fmt"
	"strings"
)

// SinkType enumerates the supported notification sink backends.
type SinkType string

const (
	SinkLog       SinkType = "log"
	SinkWebhook   SinkType = "webhook"
	SinkSlack     SinkType = "slack"
	SinkEmail     SinkType = "email"
	SinkPagerDuty SinkType = "pagerduty"
	SinkOpsGenie  SinkType = "opsgenie"
	SinkVictorOps SinkType = "victorops"
	SinkSNS       SinkType = "sns"
)

// KnownSinkTypes returns all registered sink type identifiers.
func KnownSinkTypes() []SinkType {
	return []SinkType{
		SinkLog,
		SinkWebhook,
		SinkSlack,
		SinkEmail,
		SinkPagerDuty,
		SinkOpsGenie,
		SinkVictorOps,
		SinkSNS,
	}
}

// ParseSinkType converts a raw string to a SinkType, returning an error if
// the value is not recognised.
func ParseSinkType(raw string) (SinkType, error) {
	norm := SinkType(strings.ToLower(strings.TrimSpace(raw)))
	for _, known := range KnownSinkTypes() {
		if norm == known {
			return norm, nil
		}
	}
	return "", fmt.Errorf("notifier: unknown sink type %q", raw)
}
