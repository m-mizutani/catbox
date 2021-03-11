package usecase

import (
	"github.com/m-mizutani/catbox/pkg/controller"
	"github.com/m-mizutani/catbox/pkg/model"
)

func UpdateRepoVulnStatuses(ctrl *controller.Controller, statuses []*model.RepoVulnStatus) ([]*model.VulnStatusChange, error) {
	// TODO:
	return nil, nil
}
