package notifier

import (
	"net"
	"net/smtp"
	"strings"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/tracker"
)

func TestNewEmailSink_MissingHost(t *testing.T) {
	_, err := NewEmailSink(EmailConfig{To: []string{"ops@example.com"}})
	if err == nil {
		t.Fatal("expected error for missing host")
	}
}

func TestNewEmailSink_MissingRecipients(t *testing.T) {
	_, err := NewEmailSink(EmailConfig{Host: "smtp.example.com"})
	if err == nil {
		t.Fatal("expected error for missing recipients")
	}
}

func TestNewEmailSink_DefaultPort(t *testing.T) {
	s, err := NewEmailSink(EmailConfig{
		Host: "smtp.example.com",
		To:   []string{"ops@example.com"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.cfg.Port != 587 {
		t.Errorf("expected default port 587, got %d", s.cfg.Port)
	}
}

// TestEmailSink_Send_Success spins up a minimal SMTP stub and verifies the
// sink can connect, authenticate (PLAIN), and deliver a message.
func TestEmailSink_Send_Success(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	received := make(chan string, 1)
	go smtpStub(ln, received)

	addr := ln.Addr().(*net.TCPAddr)
	sink := &emailSink{
		cfg: EmailConfig{
			Host: "127.0.0.1",
			Port: addr.Port,
			From: "cronwatch@example.com",
			To:   []string{"ops@example.com"},
		},
		auth: smtp.PlainAuth("", "", "", "127.0.0.1"),
	}

	a := tracker.Alert{
		JobName: "backup",
		Kind:    tracker.AlertDrift,
		Message: "ran 5m late",
		At:      time.Now(),
	}
	if err := sink.Send(a); err != nil {
		t.Fatalf("Send: %v", err)
	}

	select {
	case msg := <-received:
		if !strings.Contains(msg, "backup") {
			t.Errorf("expected job name in message body, got: %s", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for stub to receive message")
	}
}

// smtpStub accepts a single connection and echoes a minimal SMTP conversation.
func smtpStub(ln net.Listener, out chan<- string) {
	conn, err := ln.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

	buf := make([]byte, 4096)
	conn.Write([]byte("220 stub ready\r\n"))
	var collected strings.Builder
	for {
		n, err := conn.Read(buf)
		if err != nil {
			break
		}
		line := string(buf[:n])
		collected.WriteString(line)
		switch {
		case strings.HasPrefix(line, "EHLO"):
			conn.Write([]byte("250-stub\r\n250 AUTH PLAIN\r\n"))
		case strings.HasPrefix(line, "AUTH"):
			conn.Write([]byte("235 ok\r\n"))
		case strings.HasPrefix(line, "MAIL"):
			conn.Write([]byte("250 ok\r\n"))
		case strings.HasPrefix(line, "RCPT"):
			conn.Write([]byte("250 ok\r\n"))
		case strings.HasPrefix(line, "DATA"):
			conn.Write([]byte("354 go\r\n"))
		case strings.HasSuffix(strings.TrimRight(line, "\r\n"), "."):
			conn.Write([]byte("250 queued\r\n"))
			out <- collected.String()
			return
		case strings.HasPrefix(line, "QUIT"):
			conn.Write([]byte("221 bye\r\n"))
			return
		}
	}
}
