package webalayzer

import "context"

type WebPageStats struct {
	HTMLVersion string
	Title       string

	H1Count int
	H2Count int
	H3Count int
	H4Count int
	H5Count int
	H6Count int

	InternalLinksCount int
	ExternalLinksCount int

	InaccessibleLinksCount int
	AccessibleLinksCount   int

	// used to distinguish registration from login
	PasswordInputsCount int
}

func (wps WebPageStats) HasLoginForm() bool {
	return wps.PasswordInputsCount == 1
}
func (wps WebPageStats) HasRegistrationForm() bool {
	return wps.PasswordInputsCount > 1
}

type WebPageStatsContext struct {
	context.Context

	*WebPageStats
}
