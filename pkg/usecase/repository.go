package usecase

import (
	"github.com/m-mizutani/catbox/pkg/controller"
	"github.com/m-mizutani/catbox/pkg/model"
)

// ListRepository returns list of
func ListRepository(ctrl *controller.Controller, registry string) ([]*model.Image, error) {
	return nil, nil
}

func GetRepository(ctrl *controller.Controller, registry, repo, tag string) {}

func GetRepoVulnStatus(ctrl *controller.Controller, registry, repo, tag string) {}
