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

	img := scanReport.Image
	if img.Tag != "latest" { // TODO: It will be replaced to check repoConfig
		return nil
	}

	afterStatus, err := reportToRepoVulnStatusMap(ctrl, scanReport)
	if err != nil {
		return err
	}

	beforeStatus, err := ctrl.DB().GetRepoVulnStatusByRepo(&scanReport.Image)
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
			TaggedImage:   status.Image,
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

func reportToRepoVulnStatusMap(ctrl *controller.Controller, report *model.ScanReport) ([]*model.RepoVulnStatus, error) {
	results, err := ctrl.DownloadTrivyReport(&report.OutputTo)
	if err != nil {
		return nil, err
	}

	var statuses []*model.RepoVulnStatus
	for _, source := range results {
		for _, vuln := range source.Vulnerabilities {
			statuses = append(statuses, &model.RepoVulnStatus{
				TaggedImage: report.Image,
				RepoVulnEntry: model.RepoVulnEntry{
					VulnID:    vuln.VulnerabilityID,
					VulnType:  model.VulnPkg,
					PkgSource: source.Target,
					PkgName:   vuln.PkgName,
				},
				UpdatedAt:           report.ScannedAt,
				Status:              model.VulnStatusNew,
				DetectedBy:          report.ScannedBy,
				StatusSeq:           report.StatusSeq,
				PkgType:             source.Type,
				PkgInstalledVersion: vuln.FixedVersion,
				PkgFixedVersion:     vuln.FixedVersion,
				Description:         "",
			})
		}
	}

	return statuses, nil
}

func calcDiff(beforeStatus, afterStatus []*model.RepoVulnStatus) *statusDiff {
	remapStatus := func(statuses []*model.RepoVulnStatus) map[vulnKey]*model.RepoVulnStatus {
		statusMap := make(map[vulnKey]*model.RepoVulnStatus)
		for _, s := range statuses {
			key := vulnKey{
				source:  s.PkgSource,
				pkgName: s.PkgName,
				vulnID:  s.VulnID,
			}
			if exists, ok := statusMap[key]; !ok || exists.StatusSeq < s.StatusSeq {
				statusMap[key] = s
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
