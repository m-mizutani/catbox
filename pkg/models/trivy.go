package models

import (
	"fmt"
	"strings"

	"github.com/aquasecurity/trivy/pkg/report"
	"github.com/aquasecurity/trivy/pkg/types"
)

// TrivyResults is set of trivy report results
type TrivyResults []report.Result

// Vulnerabilities returns iterator channel of TrivyVulnerability in the report
func (x TrivyResults) Vulnerabilities() chan *TrivyVulnerability {
	ch := make(chan *TrivyVulnerability, 256)

	go func() {
		defer close(ch)
		for _, result := range x {
			for _, vuln := range result.Vulnerabilities {
				// Ignroe debian temporary vulnerability ID
				if strings.HasPrefix(vuln.VulnerabilityID, "TEMP-") {
					continue
				}

				v := TrivyVulnerability{
					PkgSource:             result.Target,
					PkgType:               result.Target,
					DetectedVulnerability: vuln,
				}
				ch <- &v
			}
		}
	}()

	return ch
}

type TrivyVulnerability struct {
	PkgSource string
	PkgType   string
	types.DetectedVulnerability
}

func (x *TrivyVulnerability) Key() string {
	return fmt.Sprintf("%s/%s/%s", x.PkgSource, x.PkgName, x.VulnerabilityID)
}
