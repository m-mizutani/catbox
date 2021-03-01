package interfaces

import "github.com/m-mizutani/catbox/pkg/model"

type DBClient interface {
	// Management
	Close() error

	// RepoVulnStatus
	PutRepoVulnStatusBatch(vulnStatuses []*model.RepoVulnStatus) error
	GetRepoVulnStatusByRepo(registry, repo, tag string) ([]*model.RepoVulnStatus, error)
	GetRepoVulnStatusByVulnID(vulnID string) ([]*model.RepoVulnStatus, error)

	// ScanReport
	PutScanReport(report *model.ScanReport) error
	GetScanReportByID(reportID string) (*model.ScanReport, error)
	GetLatestScanReportsByRepo(registry, repo, tag string) (*model.ScanReport, error)

	// ImageLayerIndex
	PutImageLayerDigest(layerDigest *model.ImageLayerIndex) error
	LookupImageLayerDigest(digest string) ([]*model.ImageLayerIndex, error)
}

type DBClientFactory func(region, tableName string) (DBClient, error)
