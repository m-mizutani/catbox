package usecase

import (
	"github.com/m-mizutani/catbox/pkg/controller"
	"github.com/m-mizutani/catbox/pkg/model"
)

// ListECRRepository returns list of ECR repository with environment variables.
func ListECRRepository(ctrl *controller.Controller, registry string) ([]*model.Image, error) {
	return nil, nil
}
