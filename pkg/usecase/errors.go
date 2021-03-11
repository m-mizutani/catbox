package usecase

import (
	"fmt"
)

var (
	// ErrReportNotFound indicates report is not found by report ID
	ErrReportNotFound = fmt.Errorf("Report not found")
)
