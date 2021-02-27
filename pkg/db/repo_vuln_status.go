package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/golambda"
)

func repoVulnStatusPK(registry, repo, tag string) string {
	return fmt.Sprintf("repo_vuln_status:%s/%s:%s", registry, repo, tag)
}

func repoVulnStatusPK2(vulnID string) string {
	return "repo_vuln_status:" + vulnID
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
			v := vulnStatuses[i+p]
			timekey := time.Unix(v.UpdatedAt, 0).Format("20060102_150405")

			items[p] = dynamoRecord{
				PK:   repoVulnStatusPK(v.Image.Registry, v.Image.Repo, v.Image.Tag),
				SK:   strings.Join([]string{v.VulnID, v.PkgSource, v.PkgName, timekey}, ":"),
				PK2:  repoVulnStatusPK2(v.VulnID),
				SK2:  strings.Join([]string{v.Image.RegistryRepoTag(), v.PkgSource, v.PkgName}, ":"),
				Docs: v,
			}
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
	pk := repoVulnStatusPK(registry, repo, tag)

	var records []*dynamoRecord
	if err := x.table.Get("pk", pk).All(&records); err != nil {
		return nil, golambda.WrapError(err, "GetRepoVulnStatusByRepo").With("registry", registry).With("repo", repo).With("tag", tag)
	}

	resp := make([]*model.RepoVulnStatus, len(records))
	for i := range records {
		if err := records[i].Unmarshal(&resp[i]); err != nil {
			return nil, err
		}
	}

	return resp, nil
}

// GetRepoVulnStatusByVulnID retrieves all RepoVulnStatus bound to a vulnID
func (x *DynamoClient) GetRepoVulnStatusByVulnID(vulnID string) ([]*model.RepoVulnStatus, error) {
	pk2 := repoVulnStatusPK2(vulnID)
	var records []*dynamoRecord

	if err := x.table.Get("pk2", pk2).Index(dynamoGSIName).All(&records); err != nil {
		return nil, golambda.WrapError(err, "GetRepoVulnStatusByVulnID").With("vulnID", vulnID)
	}

	resp := make([]*model.RepoVulnStatus, len(records))
	for i := range records {
		if err := records[i].Unmarshal(&resp[i]); err != nil {
			return nil, err
		}
	}

	return resp, nil
}
