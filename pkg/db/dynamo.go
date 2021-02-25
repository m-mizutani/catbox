package db

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/m-mizutani/catbox/pkg/models"
	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

type DBClient struct {
	table dynamo.Table
	local bool
}

type DBClientFactory func(region, tableName string) (*DBClient, error)

// NewDBClient creates DBClient
func NewDBClient(region, tableName string) (*DBClient, error) {
	ssn, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}

	table := dynamo.New(ssn).Table(tableName)
	return &DBClient{
		table: table,
	}, nil
}

// NewDBClientLocal configures DBClient with local endpoint and create a table for test and return the client.
func NewDBClientLocal(region, tableName string) (*DBClient, error) {
	// Dummy credential
	port := 8000
	if v, ok := os.LookupEnv("DYNAMO_LOCAL_PORT"); ok {
		localPort, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			panic("DYNAMO_LOCAL_PORT can not be parsed: " + v)
		}
		if 65535 < localPort {
			panic("DYNAMO_LOCAL_PORT has invalid port number")
		}
		port = int(localPort)
	}

	os.Setenv("AWS_ACCESS_KEY_ID", "x")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "x")
	ssn, err := session.NewSession(&aws.Config{
		Region:   aws.String(region),
		Endpoint: aws.String(fmt.Sprintf("http://localhost:%d", port)),
		// Credentials: credentials.NewStaticCredentials("dummy_key", "dummy_secret", "dummy_token"),
	})
	if err != nil {
		return nil, err
	}

	db := dynamo.New(ssn)
	if err := db.CreateTable(tableName, models.DBBaseRecord{}).OnDemand(true).Run(); err != nil {
		return nil, golambda.WrapError(err, "Creating local DynamoDB table")
	}

	table := dynamo.New(ssn).Table(tableName)
	return &DBClient{
		local: true,
		table: table,
	}, nil
}

// DestroyTable deletes table in local DynamoDB. It will panic if trying delete of remote DynamoDB table.
func (x *DBClient) DestroyTable() error {
	if !x.local {
		panic("DO NOT call DestroyTable for remote DynamoDB table")
	}

	if err := x.table.DeleteTable().Run(); err != nil {
		return err
	}

	return nil
}

// PutRepoVulnStatusBatch puts multiple RepoVulnStatus to DynamoDB per 25 items iteratively
func (x *DBClient) PutRepoVulnStatusBatch(vulnStatuses []*models.RepoVulnStatus) error {
	const batchSize = 25
	for i := 0; i < len(vulnStatuses); i += batchSize {
		e := i + batchSize
		if len(vulnStatuses) < e {
			e = len(vulnStatuses)
		}
		batchSize := e - i
		items := make([]interface{}, batchSize)

		// Assign hash and range keys
		for p := 0; p < batchSize; p++ {
			vulnStatuses[i+p].AssignKeys()
			items[p] = vulnStatuses[i+p]
		}

		wrote, err := x.table.Batch().Write().Put(items...).Run()
		if err != nil {
			return golambda.WrapError(err, "PutRepoVulnStatusBatch").With("items", items)
		}
		logger.With("wrote", wrote).Debug("Batch put RepoVulnStatus")
	}

	return nil
}

// GetRepoVulnStatusByRepo retrieves all RepoVulnStatus bound to registry, repo and tag
func (x *DBClient) GetRepoVulnStatusByRepo(registry, repo, tag string) ([]*models.RepoVulnStatus, error) {
	var resp []*models.RepoVulnStatus
	pk := models.RepoVulnStatusPK(registry, repo, tag)
	if err := x.table.Get("pk", pk).All(&resp); err != nil {
		return nil, golambda.WrapError(err, "GetRepoVulnStatusByRepo").With("registry", registry).With("repo", repo).With("tag", tag)
	}

	return resp, nil
}

// GetRepoVulnStatusByVulnID retrieves all RepoVulnStatus bound to a vulnID
func (x *DBClient) GetRepoVulnStatusByVulnID(vulnID string) ([]*models.RepoVulnStatus, error) {
	var resp []*models.RepoVulnStatus
	pk2 := models.RepoVulnStatusPK2(vulnID)
	if err := x.table.Get("pk2", pk2).Index("secondary").All(&resp); err != nil {
		return nil, golambda.WrapError(err, "GetRepoVulnStatusByVulnID").With("vulnID", vulnID)
	}

	return resp, nil
}

// PutScanReport puts ScanReport to DynamoDB
func (x *DBClient) PutScanReport(report *models.ScanReport) error {
	report.AssignKeys()
	if err := x.table.Put(report).Run(); err != nil {
		return golambda.WrapError(err, "PutScanReport").With("report", report)
	}
	return nil
}

// GetBatchScanReportByID returns multiple reports by reportIDs.
func (x *DBClient) GetBatchScanReportByID(reportIDs []string) ([]*models.ScanReport, error) {
	return nil, nil
}

// GetLatestScanReportsByRepo returns latest report of a repository. It returns nil if no report is available.
func (x *DBClient) GetLatestScanReportsByRepo(registry, repo, tag string) (*models.ScanReport, error) {
	return nil, nil
}

func (x *DBClient) PutImageLayerDigest(layerDigest *models.ImageLayerIndex) error {
	return nil
}

func (x *DBClient) LookupImageLayerDigests(digests []string) ([]*models.ImageLayerIndex, error) {
	return nil, nil
}
