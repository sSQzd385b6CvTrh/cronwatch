package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/cronwatch/internal/config"
	"github.com/example/cronwatch/internal/notifier"
	"github.com/example/cronwatch/internal/tracker"
	"github.com/example/cronwatch/internal/watcher"
)

func main() {
	cfgPath := flag.String("config", "cronwatch.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("cronwatch: failed to load config: %v", err)
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Build notifier with configured sinks.
	n := notifier.New()
	n.AddSink(notifier.NewLogSink(logger))

	if cfg.Webhook.URL != "" {
		ws, err := notifier.NewWebhookSink(cfg.Webhook.URL, 0)
		if err != nil {
			log.Fatalf("cronwatch: webhook sink: %v", err)
		}
		n.AddSink(ws)
	}

	if cfg.Slack.URL != "" {
		ss, err := notifier.NewSlackSink(cfg.Slack.URL, 0)
		if err != nil {
			log.Fatalf("cronwatch: slack sink: %v", err)
		}
		n.AddSink(ss)
	}

	// Build tracker and register jobs.
	t, err := tracker.New(n)
	if err != nil {
		log.Fatalf("cronwatch: tracker: %v", err)
	}
	for _, job := range cfg.Jobs {
		if err := t.Register(job.Name, job.Schedule, time.Duration(job.DriftThreshold)); err != nil {
			log.Fatalf("cronwatch: register job %q: %v", job.Name, err)
		}
	}

	pollInterval := time.Duration(cfg.PollEvery)
	w := watcher.New(t, n, pollInterval)

	ctx, stop := signal.NotifyContext(nil, os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Printf("cronwatch: started — watching %d job(s), poll interval %s", len(cfg.Jobs), pollInterval)
	w.Run(ctx)
	logger.Println("cronwatch: shutdown complete")
}
