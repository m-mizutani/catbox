package model

import (
	"fmt"
	"time"
)

type ScanRequestMessage struct {
	RequestID   string
	RequestedBy string
	RequestedAt time.Time
	Target      Image

	// Output to
	OutS3Prefix string
}

// S3Key returns S3 key to save the report
func (x *ScanRequestMessage) S3Key(scanType ScanType) string {
	return fmt.Sprintf("%s%s.json.gz", x.OutS3Prefix, scanType)
}

type InspectRequestMessage struct {
	ReportID string
}
