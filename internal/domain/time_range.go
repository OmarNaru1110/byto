package domain

import (
	"fmt"
	"regexp"
)

type TimeRange struct {
	IsAllowed bool   `json:"is_allowed"`
	Start     string `json:"start,omitempty"`
	End       string `json:"end,omitempty"`
}

var timeRangePattern = regexp.MustCompile(`^\d{2}:[0-5]\d:[0-5]\d$`)

func (tr TimeRange) Validate() error {
	if tr.Start == "" || tr.End == "" {
		return fmt.Errorf("time range requires both start and end")
	}
	if !timeRangePattern.MatchString(tr.Start) {
		return fmt.Errorf("invalid time range start %q: expected hours:minutes:seconds", tr.Start)
	}
	if !timeRangePattern.MatchString(tr.End) {
		return fmt.Errorf("invalid time range end %q: expected hours:minutes:seconds", tr.End)
	}
	return nil
}
