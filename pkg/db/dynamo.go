package db

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

const (
	dynamoGSIName = "secondary"
)

type DynamoClient struct {
	tableName string
	table     dynamo.Table
	local     bool
}

// TableName is to identify name of table created for local test
func (x *DynamoClient) TableName() string { return x.tableName }

// NewDynamoClient creates DynamoClient
func NewDynamoClient(region, tableName string) (*DynamoClient, error) {
	ssn, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}

	table := dynamo.New(ssn).Table(tableName)
	return &DynamoClient{
		table: table,
	}, nil
}

// NewDynamoClientLocal configures DynamoClient with local endpoint and create a table for test and return the client.
func NewDynamoClientLocal(region, tableName string) (*DynamoClient, error) {
	// Set port number
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

	// Add table name suffix to isolate from other test
	tableName += "-" + uuid.New().String()

	// Dummy credential
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
	if err := db.CreateTable(tableName, model.DBBaseRecord{}).OnDemand(true).Run(); err != nil {
		return nil, golambda.WrapError(err, "Creating local DynamoDB table")
	}

	table := dynamo.New(ssn).Table(tableName)
	return &DynamoClient{
		local: true,
		table: table,
	}, nil
}

// DestroyTable deletes table in local DynamoDB. It will panic if trying delete of remote DynamoDB table.
func (x *DynamoClient) DestroyTable() error {
	if !x.local {
		panic("DO NOT call DestroyTable for remote DynamoDB table")
	}

	if err := x.table.DeleteTable().Run(); err != nil {
		return err
	}

	return nil
}

// PutRepoVulnStatusBatch puts multiple RepoVulnStatus to DynamoDB per 25 items iteratively
func (x *DynamoClient) PutRepoVulnStatusBatch(vulnStatuses []*model.RepoVulnStatus) error {
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
func (x *DynamoClient) GetRepoVulnStatusByRepo(registry, repo, tag string) ([]*model.RepoVulnStatus, error) {
	var resp []*model.RepoVulnStatus
	pk := model.RepoVulnStatusPK(registry, repo, tag)
	if err := x.table.Get("pk", pk).All(&resp); err != nil {
		return nil, golambda.WrapError(err, "GetRepoVulnStatusByRepo").With("registry", registry).With("repo", repo).With("tag", tag)
	}

	return resp, nil
}

// GetRepoVulnStatusByVulnID retrieves all RepoVulnStatus bound to a vulnID
func (x *DynamoClient) GetRepoVulnStatusByVulnID(vulnID string) ([]*model.RepoVulnStatus, error) {
	var resp []*model.RepoVulnStatus
	pk2 := model.RepoVulnStatusPK2(vulnID)
	if err := x.table.Get("pk2", pk2).Index(dynamoGSIName).All(&resp); err != nil {
		return nil, golambda.WrapError(err, "GetRepoVulnStatusByVulnID").With("vulnID", vulnID)
	}

	return resp, nil
}

// PutScanReport puts ScanReport to DynamoDB
func (x *DynamoClient) PutScanReport(report *model.ScanReport) error {
	report.AssignKeys()
	if err := x.table.Put(report).Run(); err != nil {
		return golambda.WrapError(err, "PutScanReport").With("report", report)
	}
	return nil
}

// GetScanReportByID returns multiple reports by reportIDs.
func (x *DynamoClient) GetScanReportByID(reportID string) (*model.ScanReport, error) {
	var resp model.ScanReport
	pk2 := model.ScanReportPK2()
	sk2 := model.ScanReportSK2(reportID)
	query := x.table.Get("pk2", pk2).Index(dynamoGSIName).Range("sk2", dynamo.Equal, sk2)

	if err := query.One(&resp); err != nil {
		if err == dynamo.ErrNotFound {
			return nil, nil
		}
	}

	return &resp, nil
}

// GetLatestScanReportsByRepo returns latest report of a repository. It returns nil if no report is available.
func (x *DynamoClient) GetLatestScanReportsByRepo(registry, repo, tag string) (*model.ScanReport, error) {
	var resp model.ScanReport
	pk := model.ScanReportPK(registry, repo, tag)
	query := x.table.Get("pk", pk).Order(dynamo.Descending).Limit(1)

	if err := query.One(&resp); err != nil {
		if err == dynamo.ErrNotFound {
			return nil, nil
		}
		return nil, golambda.WrapError(err, "GetLatestScanReportsByRepo").With("query", query)
	}

	return &resp, nil
}

func (x *DynamoClient) PutImageLayerDigest(layerDigest *model.ImageLayerIndex) error {
	return nil
}

func (x *DynamoClient) LookupImageLayerDigests(digests []string) ([]*model.ImageLayerIndex, error) {
	return nil, nil
}
