package notifier

import (
	"fmt"
	"io"
	"text/template"
	"time"
)

const defaultTemplate = "[{{.OccuredAt.Format \"2006-01-02T15:04:05Z07:00\"}}] {{.Level}} job={{.JobName}} msg={{.Message}}\n"

// LogSink writes formatted alerts to an io.Writer (e.g. os.Stderr or a log file).
type LogSink struct {
	w   io.Writer
	tpl *template.Template
}

// NewLogSink creates a LogSink that writes to w using the default log format.
func NewLogSink(w io.Writer) *LogSink {
	tpl := template.Must(template.New("alert").Funcs(template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
	}).Parse(defaultTemplate))
	return &LogSink{w: w, tpl: tpl}
}

// Send formats the alert and writes it to the underlying writer.
func (l *LogSink) Send(a Alert) error {
	if err := l.tpl.Execute(l.w, a); err != nil {
		return fmt.Errorf("log_sink: template execute: %w", err)
	}
	return nil
}
