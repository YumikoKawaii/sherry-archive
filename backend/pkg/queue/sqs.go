package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
)

type Client struct {
	sqs      *sqs.Client
	queueURL string
}

func NewClient(ctx context.Context, region, queueURL string) (*Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	return &Client{
		sqs:      sqs.NewFromConfig(cfg),
		queueURL: queueURL,
	}, nil
}

// UploadMessage is the SQS payload for async zip processing.
// Shared between the server (enqueue) and Lambda (consume).
type UploadMessage struct {
	TaskID    uuid.UUID  `json:"task_id"`
	Type      string     `json:"type"`
	S3Key     string     `json:"s3_key"`
	MangaID   uuid.UUID  `json:"manga_id"`
	ChapterID *uuid.UUID `json:"chapter_id,omitempty"` // nil for oneshot_zip
	OwnerID   uuid.UUID  `json:"owner_id"`
}

func (c *Client) Enqueue(ctx context.Context, msg UploadMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = c.sqs.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(c.queueURL),
		MessageBody: aws.String(string(body)),
	})
	return err
}
