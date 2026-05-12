// Package utils holds small shared helpers used by command entrypoints (e.g. build metadata printing).
package utils

import (
	"fmt"
	"sys-metrics/internal/common"
)

func PrintBuildInfo(version string, date, commit string) {
	if version == "" {
		version = common.NotApplicable
	}
	if date == "" {
		date = common.NotApplicable
	}
	if commit == "" {
		commit = common.NotApplicable
	}
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", version, date, commit)
}
