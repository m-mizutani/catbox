package db

import (
	"fmt"

	"github.com/guregu/dynamo"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/golambda"
)

const (
	repoVulnStatusTimeKeyFormat = "20060102_150405"
	repoVulnStatusKeyPrefix     = "repo_vuln_status:"
	repoVulnChangeLogKeyPrefix  = "repo_vuln_changelog:"
)

func repoVulnStatusPK(img *model.Image) string {
	return repoVulnStatusKeyPrefix + img.RegistryRepoTag()
}

func repoVulnStatusSK(entry *model.RepoVulnEntry) string {
	return entry.Key()
}

func repoVulnStatusPK2(vulnID string) string {
	return repoVulnStatusKeyPrefix + vulnID
}

func repoVulnStatusSK2(status *model.RepoVulnStatus) string {
	return status.Image.RegistryRepoTag() + ":" + status.RepoVulnEntry.TypeKey()
}

func repoVulnChangeLogPK(img *model.Image) string {
	return repoVulnChangeLogKeyPrefix + img.RegistryRepoTag()
}

func repoVulnChangeLogSK(entry *model.RepoVulnEntry, seq int64) string {
	return fmt.Sprintf("%s:%016X", repoVulnChangeLogSKPrefix(entry), seq)
}

func repoVulnChangeLogSKPrefix(entry *model.RepoVulnEntry) string {
	return entry.Key() + ":"
}

func repoVulnChangeLogPK2(vulnID string) string {
	return repoVulnChangeLogKeyPrefix + vulnID
}

func repoVulnChangeLogSK2(img *model.Image, entry *model.RepoVulnEntry, seq int64) string {
	return fmt.Sprintf("%s:%016d", repoVulnChangeLogSK2Prefix(img, entry), seq)
}

func repoVulnChangeLogSK2Prefix(img *model.Image, entry *model.RepoVulnEntry) string {
	return img.RegistryRepoTag() + ":" + entry.TypeKey() + ":"
}

// CreateRepoVulnStatus puts only new RepoVulnStatus. If an item already exists, skip it. It returns true if inserted
func (x *DynamoClient) CreateRepoVulnStatus(status *model.RepoVulnStatus) (bool, error) {
	vulnStatItem := dynamoRecord{
		PK:  repoVulnStatusPK(&status.Image),
		SK:  repoVulnStatusSK(&status.RepoVulnEntry),
		PK2: repoVulnStatusPK2(status.VulnID),
		SK2: repoVulnStatusSK2(status),
		Doc: status,
	}
	changeLogItem := dynamoRecord{
		PK:  repoVulnChangeLogPK(&status.Image),
		SK:  repoVulnChangeLogSK(&status.RepoVulnEntry, status.StatusSeq),
		PK2: repoVulnChangeLogPK2(status.VulnID),
		SK2: repoVulnChangeLogSK2(&status.Image, &status.RepoVulnEntry, status.StatusSeq),
		Doc: model.RepoVulnChangeLog{
			Image:         status.Image,
			RepoVulnEntry: status.RepoVulnEntry,
			Status:        model.VulnStatusNew,
			UpdatedAt:     status.UpdatedAt,
			StatusSeq:     status.StatusSeq,
		},
	}

	tx := x.db.WriteTx()
	tx.Put(x.table.Put(vulnStatItem).If("attribute_not_exists(pk) AND attribute_not_exists(sk)"))
	tx.Put(x.table.Put(changeLogItem))
	err := tx.Run()
	if err != nil {
		if isConditionalCheckErr(err) || isTransactionException(err) {
			return false, nil
		}
		return false, golambda.WrapError(err, "PutRepoVulnStatusBatch").With("status", status).With("vulnStatItem", vulnStatItem).With("changeLogItem", changeLogItem)
	}

	return true, nil
}

// UpdateRepoVulnStatus updates Status and sequence if status has been changed AND sequence is greater than old one. It returned true as 1st value if updated or false if update was cancelled. Cancel is occurred by not only conditions are not matched but also PK and SK do not exist.
func (x *DynamoClient) UpdateRepoVulnStatus(changeLog *model.RepoVulnChangeLog) (bool, error) {
	changeLogItem := dynamoRecord{
		PK:  repoVulnChangeLogPK(&changeLog.Image),
		SK:  repoVulnChangeLogSK(&changeLog.RepoVulnEntry, changeLog.StatusSeq),
		PK2: repoVulnChangeLogPK2(changeLog.VulnID),
		SK2: repoVulnChangeLogSK2(&changeLog.Image, &changeLog.RepoVulnEntry, changeLog.StatusSeq),
		Doc: changeLog,
	}

	rvsPK := repoVulnStatusPK(&changeLog.Image)
	rvsSK := repoVulnStatusSK(&changeLog.RepoVulnEntry)
	tx := x.db.WriteTx()
	tx.Update(x.table.Update("pk", rvsPK).Range("sk", rvsSK).
		If("doc.'Status' <> ?", changeLog.Status).
		If("doc.'StatusSeq' < ?", changeLog.StatusSeq).
		Set("doc.'Status'", changeLog.Status).
		Set("doc.'StatusSeq'", changeLog.StatusSeq))
	tx.Put(x.table.Put(changeLogItem))
	if err := tx.Run(); err != nil {
		if isConditionalCheckErr(err) || isTransactionException(err) {
			return false, nil
		}
		return false, golambda.WrapError(err, "dynamo.WriteTx").With("changeLog", changeLog)
	}
	return true, nil
}

// UpdateRepoVulnDescription updates description of RepoVulnStatus
func (x *DynamoClient) UpdateRepoVulnDescription(img *model.Image, entry *model.RepoVulnEntry, descr string) error {
	pk := repoVulnStatusPK(img)
	sk := repoVulnStatusSK(entry)
	query := x.table.Update("pk", pk).Range("sk", sk).Set("doc.'Description'", descr)
	if err := query.Run(); err != nil {
		return golambda.WrapError(err, "dynamo.Update").With("pk", pk).With("sk", sk)
	}
	return nil
}

// GetRepoVulnStatusByRepo retrieves all RepoVulnStatus bound to registry, repo and tag
func (x *DynamoClient) GetRepoVulnStatusByRepo(img *model.Image) ([]*model.RepoVulnStatus, error) {
	pk := repoVulnStatusPK(img)

	var records []*dynamoRecord
	if err := x.table.Get("pk", pk).All(&records); err != nil {
		return nil, golambda.WrapError(err, "GetRepoVulnStatusByRepo").With("img", img)
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

func (x *DynamoClient) GetRepoVulnChangeLogs(img *model.Image) ([]*model.RepoVulnChangeLog, error) {
	pk := repoVulnChangeLogPK(img)
	var records []*dynamoRecord
	if err := x.table.Get("pk", pk).All(&records); err != nil {
		return nil, err
	}

	resp := make([]*model.RepoVulnChangeLog, len(records))
	for i := range records {
		if err := records[i].Unmarshal(&resp[i]); err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func (x *DynamoClient) GetRepoVulnEntryChangeLogs(img *model.Image, entry *model.RepoVulnEntry) ([]*model.RepoVulnChangeLog, error) {
	pk := repoVulnChangeLogPK(img)
	skPrefix := repoVulnChangeLogSKPrefix(entry)

	var records []*dynamoRecord
	if err := x.table.Get("pk", pk).Range("sk", dynamo.BeginsWith, skPrefix).All(&records); err != nil {
		return nil, err
	}

	resp := make([]*model.RepoVulnChangeLog, len(records))
	for i := range records {
		if err := records[i].Unmarshal(&resp[i]); err != nil {
			return nil, err
		}
	}

	return resp, nil
}
