package model

import (
	"fmt"
	"time"
)

type ScanRequestMessage struct {
	RequestID   string    `json:"request_id"`
	RequestedBy string    `json:"requested_by"`
	RequestedAt time.Time `json:"requested_at"`
	Target      Image     `json:"target"`

	// Output to
	S3Region string `json:"s3_region"`
	S3Bucket string `json:"s3_bucket"`
	S3Prefix string `json:"s3_prefix"`
}

// S3Key returns S3 key to save the report
func (x *ScanRequestMessage) S3Key(scanType ScanType) string {
	return fmt.Sprintf("%s%s.json.gz", x.S3Prefix, scanType)
}

type InspectRequestMessage struct {
	ReportID string `json:"report_id"`
}
