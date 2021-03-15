package db

import (
	"errors"
	"time"

	"github.com/guregu/dynamo"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/golambda"
)

const (
	vulnInfoKeyPrefix     = "vuln_info:"
	vulnInfoKeyTimeFormat = "20060102_150405"
)

func vulnInfoPK(vulnID string) string { return vulnInfoKeyPrefix + vulnID }
func vulnInfoSK() string              { return "-" }
func vulnInfoPK2() string             { return "list:vuln_info" }
func vulnInfoSK2(detectedAt int64, vulnID string) string {
	return time.Unix(detectedAt, 0).UTC().Format(vulnInfoKeyTimeFormat) + "/" + vulnID
}

// PutVulnInfoBatch writes all vulnInfoSet records and returns only new ones as 1st returned value. (NOTE: it does not use transaction and conditional check because of performance, then no guarantee of consistency)
func (x *DynamoClient) PutVulnInfoBatch(vulnInfoSet []*model.VulnInfo) ([]*model.VulnInfo, error) {
	// Weak update check, no guarantee of consistency
	vulnIDs := make([]string, len(vulnInfoSet))
	for i := 0; i < len(vulnIDs); i++ {
		vulnIDs[i] = vulnInfoSet[i].ID
	}
	exists, err := x.GetVulnInfoBatch(vulnIDs)
	if err != nil {
		return nil, err
	}
	existsMap := map[string]*model.VulnInfo{}
	for _, v := range exists {
		existsMap[v.ID] = v
	}

	// Inserting
	const step = 25
	for s := 0; s < len(vulnInfoSet); s += step {
		e := s + step
		if len(vulnInfoSet) < e {
			e = len(vulnInfoSet)
		}
		records := make([]interface{}, e-s)
		for p := 0; p < e-s; p++ {
			v := vulnInfoSet[s+p]
			records[p] = &dynamoRecord{
				PK:  vulnInfoPK(v.ID),
				SK:  vulnInfoSK(),
				PK2: vulnInfoPK2(),
				SK2: vulnInfoSK2(v.DetectedAt, v.ID),
				Doc: v,
			}
		}

		q := x.table.Batch("pk", "sk")
		if _, err := q.Write().Put(records...).Run(); err != nil {
			return nil, golambda.WrapError(err, "Fail to batch put of VulnInfo").With("records", records)
		}
	}

	var newVulnInfo []*model.VulnInfo
	for _, v := range vulnInfoSet {
		if _, ok := existsMap[v.ID]; !ok {
			newVulnInfo = append(newVulnInfo, v)
		}
	}
	return newVulnInfo, nil
}

// GetVulnInfoBatch returns vulnInfo
func (x *DynamoClient) GetVulnInfoBatch(vulnIDs []string) ([]*model.VulnInfo, error) {
	if len(vulnIDs) == 0 {
		return nil, nil
	}

	var records []*dynamoRecord
	keys := make([]dynamo.Keyed, len(vulnIDs))
	for i := 0; i < len(keys); i++ {
		keys[i] = &dynamoRecord{
			PK: vulnInfoPK(vulnIDs[i]),
			SK: vulnInfoSK(),
		}
	}

	if err := x.table.Batch("pk", "sk").Get(keys...).All(&records); err != nil {
		if errors.Is(err, dynamo.ErrNotFound) {
			return nil, nil
		}
		return nil, golambda.WrapError(err, "")
	}

	results := make([]*model.VulnInfo, len(records))
	for i := 0; i < len(results); i++ {
		if err := records[i].Unmarshal(&results[i]); err != nil {
			return nil, err
		}
	}

	return results, nil
}
