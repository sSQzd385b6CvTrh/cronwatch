package notifier

import "testing"

func TestParseSinkType_Known(t *testing.T) {
	known := []string{
		"log", "webhook", "email", "slack",
		"pagerduty", "opsgenie", "victorops",
		"sns", "datadog", "grafana", "splunk", "teams",
	}
	for _, raw := range known {
		t.Run(raw, func(t *testing.T) {
			st, err := ParseSinkType(raw)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", raw, err)
			}
			if string(st) != raw {
				t.Errorf("want %q, got %q", raw, st)
			}
		})
	}
}

func TestParseSinkType_Unknown(t *testing.T) {
	cases := []string{"", "discord", "SLACK", "Teams"}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			_, err := ParseSinkType(raw)
			if err == nil {
				t.Fatalf("expected error for %q", raw)
			}
		})
	}
}
