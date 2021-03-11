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
	TargetMeta  ImageMeta

	// Output to
	OutS3Prefix string
}

// S3Key returns S3 key to save the report
func (x *ScanRequestMessage) S3Key(scanner Scanner) string {
	return fmt.Sprintf("%s%s.json.gz", x.OutS3Prefix, scanner)
}

// InspectRequestMessage has only reportID and inspect procedure needs to retrieve report overall from DB
type InspectRequestMessage struct {
	ReportID string
}

// ChangeMessage has vulnerability and status of repository information. This message will be sent to changeTopic and other stack receives the message.
type ChangeMessage struct {
	// NewVuln presents new vulnerabilities, "new" means a first seen vulnerability in the catbox for NewVuln
	NewVuln []*VulnInfo
	// UpdateStatus presents set of RepoVulnStatus updated status
	UpdatedStatus []*RepoVulnStatus
}
