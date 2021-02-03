package webalayzer

import (
	"context"
	"sync"
	"time"
)

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

	InaccessibleLinksCount int32
	AccessibleLinksCount   int32

	// used to distinguish registration from login
	PasswordInputsCount int

	ProcessingDuration time.Duration
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

	accessibilityWg sync.WaitGroup
}
