package notifier

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// SNSPublisher is the subset of the SNS client API used by SNSSink.
type SNSPublisher interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

// SNSSink sends alerts to an AWS SNS topic.
type SNSSink struct {
	topicARN string
	client   SNSPublisher
}

// NewSNSSink constructs an SNSSink using the default AWS credential chain.
// topicARN must be a valid SNS topic ARN.
func NewSNSSink(ctx context.Context, topicARN string) (*SNSSink, error) {
	if topicARN == "" {
		return nil, fmt.Errorf("sns: topic ARN must not be empty")
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("sns: load aws config: %w", err)
	}
	return &SNSSink{
		topicARN: topicARN,
		client:   sns.NewFromConfig(cfg),
	}, nil
}

// newSNSSinkWithClient is used in tests to inject a mock publisher.
func newSNSSinkWithClient(topicARN string, client SNSPublisher) *SNSSink {
	return &SNSSink{topicARN: topicARN, client: client}
}

// Send publishes an Alert to the configured SNS topic.
func (s *SNSSink) Send(a Alert) error {
	subject := fmt.Sprintf("cronwatch alert: %s", a.JobName)
	body := fmt.Sprintf("%s\nJob: %s\nTime: %s", a.Message, a.JobName, a.At.Format(time.RFC3339))
	_, err := s.client.Publish(context.Background(), &sns.PublishInput{
		TopicArn: aws.String(s.topicARN),
		Subject:  aws.String(subject),
		Message:  aws.String(body),
	})
	if err != nil {
		return fmt.Errorf("sns: publish: %w", err)
	}
	return nil
}
