package interfaces

import "github.com/m-mizutani/catbox/pkg/model"

type DBClient interface {
	// Management
	Close() error

	// StatusSequence
	RetrieveStatusSequence() (int64, error)

	// RepoVulnStatus
	CreateRepoVulnStatus(status *model.RepoVulnStatus) (bool, error)
	UpdateRepoVulnStatus(changeLog *model.RepoVulnChangeLog) (bool, error)
	GetRepoVulnStatusByRepo(img *model.Image) ([]*model.RepoVulnStatus, error)
	GetRepoVulnStatusByVulnID(vulnID string) ([]*model.RepoVulnStatus, error)

	// RepoVulnChangeLog
	GetRepoVulnChangeLogs(img *model.Image) ([]*model.RepoVulnChangeLog, error)
	GetRepoVulnEntryChangeLogs(img *model.Image, entry *model.RepoVulnEntry) ([]*model.RepoVulnChangeLog, error)

	// ScanReport
	PutScanReport(report *model.ScanReport) error
	GetScanReportByID(reportID string) (*model.ScanReport, error)
	GetLatestScanReportsByRepo(registry, repo, tag string) (*model.ScanReport, error)

	// ImageLayerIndex
	PutImageLayerDigest(layerDigest *model.ImageLayerIndex) error
	LookupImageLayerDigest(digest string) ([]*model.ImageLayerIndex, error)
}

type DBClientFactory func(region, tableName string) (DBClient, error)
