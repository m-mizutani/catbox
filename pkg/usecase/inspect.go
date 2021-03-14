package usecase

import (
	"github.com/m-mizutani/catbox/pkg/controller"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/golambda"
)

type vulnKey struct {
	source  string
	pkgName string
	vulnID  string
}

type statusDiff struct {
	Added     []*model.RepoVulnStatus
	Fixed     []*model.RepoVulnStatus
	Regressed []*model.RepoVulnStatus
}

// InspectScanReport is workflow of inspect lambda function
func InspectScanReport(ctrl *controller.Controller, msg *model.InspectRequestMessage) error {
	scanReport, err := ctrl.DB().GetScanReportByID(msg.ReportID)
	if err != nil {
		return err
	}

	if scanReport == nil {
		return golambda.WrapError(ErrReportNotFound).With("msg", msg)
	}

	img := scanReport.TaggedImage
	if img.Tag != "latest" { // TODO: It will be replaced to check repoConfig
		return nil
	}

	// Handle VulnInfo
	vulnInfoSet, err := reportToVulnInfo(ctrl, scanReport)
	if err != nil {
		return err
	}
	if err := InsertVulnInfo(ctrl, vulnInfoSet); err != nil {
		return err
	}

	// Handle RepoVulnStatus

	afterStatus, err := reportToRepoVulnStatusMap(ctrl, scanReport)
	if err != nil {
		return err
	}

	beforeStatus, err := ctrl.DB().GetRepoVulnStatusByRepo(&img)
	if err != nil {
		return err
	}

	diff := calcDiff(beforeStatus, afterStatus)

	if err := InsertRepoVulnStatus(ctrl, diff.Added); err != nil {
		return err
	}
	if err := UpdateRepoVulnStatus(ctrl, diff.Fixed, model.VulnStatusFixed, scanReport.ScannedAt, scanReport.StatusSeq); err != nil {
		return err
	}
	if err := UpdateRepoVulnStatus(ctrl, diff.Regressed, model.VulnStatusRegressed, scanReport.ScannedAt, scanReport.StatusSeq); err != nil {
		return err
	}

	return nil
}

// InsertVulnInfo adds vulnInfo to DB and publish new vulnInfo
func InsertVulnInfo(ctrl *controller.Controller, vulnInfoSet []*model.VulnInfo) error {
	newVulns, err := ctrl.DB().PutVulnInfoBatch(vulnInfoSet)
	if err != nil {
		return err
	}

	if len(newVulns) == 0 {
		return nil
	}

	if err := ctrl.PublishChangeMessage(&model.ChangeMessage{
		NewVuln: newVulns,
	}); err != nil {
		return err
	}

	return nil
}

// InsertRepoVulnStatus adds new RepoVulnStatus set
func InsertRepoVulnStatus(ctrl *controller.Controller, newStatuses []*model.RepoVulnStatus) error {
	var inserted []*model.RepoVulnStatus
	for _, status := range newStatuses {
		done, err := ctrl.DB().CreateRepoVulnStatus(status)
		if err != nil {
			return err
		}
		if done {
			inserted = append(inserted, status)
		}
	}

	if len(inserted) > 0 {
		msg := model.ChangeMessage{
			UpdatedStatus: inserted,
		}
		if err := ctrl.PublishChangeMessage(&msg); err != nil {
			return err
		}
	}

	return nil
}

// UpdateRepoVulnStatus updates status, timestamp (UpdatedAt) and status sequence.
func UpdateRepoVulnStatus(ctrl *controller.Controller, originals []*model.RepoVulnStatus, updateTo model.VulnStatus, ts, seq int64) error {
	var updated []*model.RepoVulnStatus
	for _, status := range originals {
		changeLog := &model.RepoVulnChangeLog{
			TaggedImage:   status.TaggedImage,
			RepoVulnEntry: status.RepoVulnEntry,
			Status:        updateTo,
			UpdatedAt:     ts,
			StatusSeq:     seq,
		}

		done, err := ctrl.DB().UpdateRepoVulnStatus(changeLog)
		if err != nil {
			return golambda.WrapError(err).With("change", changeLog)
		}
		if done {
			status.Status = updateTo
			status.UpdatedAt = ts
			status.StatusSeq = seq

			updated = append(updated, status)
		}
	}

	if len(updated) > 0 {
		msg := model.ChangeMessage{
			UpdatedStatus: updated,
		}
		if err := ctrl.PublishChangeMessage(&msg); err != nil {
			return err
		}
	}

	return nil
}

func reportToVulnInfo(ctrl *controller.Controller, report *model.ScanReport) ([]*model.VulnInfo, error) {
	results, err := ctrl.DownloadTrivyReport(&report.OutputTo)
	if err != nil {
		return nil, err
	}

	vulnMap := make(map[string]*model.VulnInfo)
	for _, source := range results {
		for _, vuln := range source.Vulnerabilities {
			cvssMap := map[string]string{}
			for vendor, cvss := range vuln.CVSS {
				if cvss.V2Vector != "" {
					cvssMap[vendor+"/v2"] = cvss.V2Vector
				}
				if cvss.V3Vector != "" {
					cvssMap[vendor+"/v3"] = cvss.V3Vector
				}
			}

			vulnMap[vuln.VulnerabilityID] = &model.VulnInfo{
				ID:          vuln.VulnerabilityID,
				Type:        model.VulnPkg,
				CVSS:        cvssMap,
				Title:       vuln.Title,
				Description: vuln.Description,
				References:  vuln.References,
				DetectedAt:  report.ScannedAt,
				PkgType:     source.Type,
				PkgName:     vuln.PkgName,
			}
		}
	}

	var vulnSet []*model.VulnInfo
	for _, vuln := range vulnMap {
		vulnSet = append(vulnSet, vuln)
	}
	return vulnSet, nil
}

func reportToRepoVulnStatusMap(ctrl *controller.Controller, report *model.ScanReport) ([]*model.RepoVulnStatus, error) {
	results, err := ctrl.DownloadTrivyReport(&report.OutputTo)
	if err != nil {
		return nil, err
	}

	var statuses []*model.RepoVulnStatus
	for _, source := range results {
		for _, vuln := range source.Vulnerabilities {
			statuses = append(statuses, &model.RepoVulnStatus{
				TaggedImage: report.TaggedImage,
				RepoVulnEntry: model.RepoVulnEntry{
					VulnID:     vuln.VulnerabilityID,
					VulnType:   model.VulnPkg,
					PkgSource:  source.Target,
					PkgName:    vuln.PkgName,
					PkgVersion: vuln.InstalledVersion,
				},
				UpdatedAt:       report.ScannedAt,
				Status:          model.VulnStatusNew,
				DetectedBy:      report.ScannedBy,
				StatusSeq:       report.StatusSeq,
				PkgType:         source.Type,
				PkgFixedVersion: vuln.FixedVersion,
				Description:     "",
			})
		}
	}

	return statuses, nil
}

func calcDiff(beforeStatus, afterStatus []*model.RepoVulnStatus) *statusDiff {

	remapStatus := func(statuses []*model.RepoVulnStatus) map[model.RepoVulnEntry]*model.RepoVulnStatus {
		statusMap := make(map[model.RepoVulnEntry]*model.RepoVulnStatus)
		for _, s := range statuses {
			if exists, ok := statusMap[s.RepoVulnEntry]; !ok || exists.StatusSeq < s.StatusSeq {
				statusMap[s.RepoVulnEntry] = s
			}
		}

		return statusMap
	}

	var results statusDiff
	beforeMap := remapStatus(beforeStatus)
	afterMap := remapStatus(afterStatus)

	for vKey, stat := range beforeMap {
		if _, ok := afterMap[vKey]; !ok {
			results.Fixed = append(results.Fixed, stat)
		}
	}

	for vKey, aStat := range afterMap {
		if bStat, ok := beforeMap[vKey]; !ok {
			results.Added = append(results.Added, aStat)
		} else if bStat.Status == model.VulnStatusFixed {
			results.Regressed = append(results.Regressed, aStat)
		}
	}

	return &results
}
