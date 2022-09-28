package irunner

import (
  "fmt"
  "os/exec"
)

// 用于运行 process 的环境
type Executor interface {
  Run(process *Process) (runid string, retChan <-chan int, err error)
  // CWL 本身并无中断执行的机制，因此只需要Run 接口即可
}

// 本地运行环境
type LocalExecutor struct {

}


func (exe LocalExecutor) Run(process *Process) (runid string, retChan <-chan int, err error) {
  envs := process.Env()
  cmds, err := process.Command()
  // SET INPUTS
  // set stdout
  // set stdin
  // set image
  // migrate inputs
  if err = process.MigrateInputs(); err != nil {
    return "", nil, err
  }
  r := exec.Command(cmds[0], cmds[1:]...)
  for k ,v := range envs {
    r.Env = append(r.Env, fmt.Sprintf("%s=%s", k, v))
  }
  r.Dir = process.runtime.RootHost
  err = r.Start()
  if err != nil {
    return  "", nil, err
  }
  pid := r.Process.Pid
  rChan := make(chan int)
  go func() {
    r.Wait()
    rChan <- r.ProcessState.ExitCode()
    close(rChan)
  }()
  return fmt.Sprint(pid), rChan, nil
}
