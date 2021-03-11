package db

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/golambda"
)

/*


// AssignKeys sets PK (hash key) and SK (range key) from field variables
func (x *ScanReport) AssignKeys() {
	if x.ReportID == "" {
		x.ReportID = uuid.New().String()
	}

	x.PK = ScanReportPK(x.Registry, x.Repo, x.Tag)
	x.SK = string(x.ScanType) + "/" + time.Unix(x.ScannedAt, 0).Format("2006-01-02T15:04:05") + "/" + x.ReportID
	x.PK2 = ScanReportPK2()
	x.SK2 = ScanReportSK2(x.ReportID)
}

*/

func scanReportPK(registry, repo, tag string) string {
	return fmt.Sprintf("report:%s/%s:%s", registry, repo, tag)
}
func scanReportPK2() string {
	return "list:report"
}
func scanReportSK2(reportID string) string {
	return reportID
}

// PutScanReport puts ScanReport to DynamoDB
func (x *DynamoClient) PutScanReport(report *model.ScanReport) error {
	if report.ReportID == "" {
		report.ReportID = uuid.New().String()
	}

	record := dynamoRecord{
		PK:  scanReportPK(report.Registry, report.Repo, report.Tag),
		SK:  string(report.ScannedBy) + "/" + time.Unix(report.ScannedAt, 0).Format("2006-01-02T15:04:05") + "/" + report.ReportID,
		PK2: scanReportPK2(),
		SK2: scanReportSK2(report.ReportID),

		Doc: report,
	}

	if err := x.table.Put(record).Run(); err != nil {
		return golambda.WrapError(err, "PutScanReport").With("report", report)
	}

	return nil
}

// GetScanReportByID returns multiple reports by reportIDs.
func (x *DynamoClient) GetScanReportByID(reportID string) (*model.ScanReport, error) {
	var record dynamoRecord
	pk2 := scanReportPK2()
	sk2 := scanReportSK2(reportID)
	query := x.table.Get("pk2", pk2).Index(dynamoGSIName).Range("sk2", dynamo.Equal, sk2)

	if err := query.One(&record); err != nil {
		if err == dynamo.ErrNotFound {
			return nil, nil
		}
	}

	var resp model.ScanReport
	if err := record.Unmarshal(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetLatestScanReportsByRepo returns latest report of a repository. It returns nil if no report is available.
func (x *DynamoClient) GetLatestScanReportsByRepo(registry, repo, tag string) (*model.ScanReport, error) {
	var record dynamoRecord
	pk := scanReportPK(registry, repo, tag)
	query := x.table.Get("pk", pk).Order(dynamo.Descending).Limit(1)

	if err := query.One(&record); err != nil {
		if err == dynamo.ErrNotFound {
			return nil, nil
		}
		return nil, golambda.WrapError(err, "GetLatestScanReportsByRepo").With("query", query)
	}

	var resp model.ScanReport
	if err := record.Unmarshal(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
