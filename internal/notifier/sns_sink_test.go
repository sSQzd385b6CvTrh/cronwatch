package notifier

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type mockSNSPublisher struct {
	captured *sns.PublishInput
	err      error
}

func (m *mockSNSPublisher) Publish(_ context.Context, params *sns.PublishInput, _ ...func(*sns.Options)) (*sns.PublishOutput, error) {
	m.captured = params
	return &sns.PublishOutput{}, m.err
}

func TestNewSNSSink_EmptyARN(t *testing.T) {
	_, err := NewSNSSink(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty topic ARN")
	}
}

func TestSNSSink_Send_Success(t *testing.T) {
	mock := &mockSNSPublisher{}
	s := newSNSSinkWithClient("arn:aws:sns:us-east-1:123456789012:cronwatch", mock)
	a := Alert{
		JobName: "nightly-report",
		Message: "drift detected",
		At:      time.Now(),
	}
	if err := s.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.captured == nil {
		t.Fatal("expected Publish to be called")
	}
	if !strings.Contains(*mock.captured.Subject, "nightly-report") {
		t.Errorf("subject missing job name: %q", *mock.captured.Subject)
	}
	if !strings.Contains(*mock.captured.Message, "drift detected") {
		t.Errorf("message missing alert text: %q", *mock.captured.Message)
	}
}

func TestSNSSink_Send_PublishError(t *testing.T) {
	mock := &mockSNSPublisher{err: errors.New("aws throttled")}
	s := newSNSSinkWithClient("arn:aws:sns:us-east-1:123456789012:cronwatch", mock)
	err := s.Send(Alert{JobName: "job", Message: "msg", At: time.Now()})
	if err == nil {
		t.Fatal("expected error from publish failure")
	}
	if !strings.Contains(err.Error(), "aws throttled") {
		t.Errorf("expected underlying error in message, got: %v", err)
	}
}
