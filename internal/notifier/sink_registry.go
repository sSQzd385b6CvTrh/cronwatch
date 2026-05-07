package notifier

import "fmt"

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
	SinkGrafana   SinkType = "grafana"
	SinkSplunk    SinkType = "splunk"
	SinkTeams     SinkType = "teams"
)

// ParseSinkType validates and normalises a raw string into a SinkType.
func ParseSinkType(raw string) (SinkType, error) {
	switch SinkType(raw) {
	case SinkLog, SinkWebhook, SinkEmail, SinkSlack,
		SinkPagerDuty, SinkOpsGenie, SinkVictorOps,
		SinkSNS, SinkDatadog, SinkGrafana, SinkSplunk, SinkTeams:
		return SinkType(raw), nil
	default:
		return "", fmt.Errorf("notifier: unknown sink type %q", raw)
	}
}
