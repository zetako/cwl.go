package slex

import "github.com/lijiang2014/cwl.go/runner"

const StarlightJobNameLimit = 32

type StarlightExecutor struct {
	alloc JobAllocationModel
}

func (s StarlightExecutor) Run(process *runner.Process) (runID string, retChan <-chan int, err error) {
	submit := NewSubmitModelFrom(s.alloc.Get(process.PathID))

	submit.RuntimeParams.JobName = process.ShortPath(StarlightJobNameLimit)
	submit.RuntimeParams.Env = process.Env()
	submit.RuntimeParams.Stdin, submit.RuntimeParams.Stdout, submit.RuntimeParams.Stderr = process.GetRedirection()

	submit.RuntimeParams.Cmd, err = process.Command()
}

func (s StarlightExecutor) QueryRuntime(limits runner.ResourcesLimites) runner.Runtime {
	// NO, I can't return anything.
	return runner.Runtime{}
}
