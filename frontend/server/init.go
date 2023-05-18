package server

import (
	"github.com/lijiang2014/cwl.go"
	"github.com/lijiang2014/cwl.go/runner"
)

// Variables to run a server
var (
	allSteps     []string
	successSteps []string
	failedSteps  []string
	waitingSteps []string
	Values       cwl.Values

	engine runner.Engine
)
