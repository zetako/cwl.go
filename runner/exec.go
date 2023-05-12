package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/lijiang2014/cwl.go"
)

// 用于运行 process 的环境
type Executor interface {
	// Run will start a process
	//   - runID can be used to identify the running process, maybe pid or others
	//   - retChan will release the return value of the process
	//   - signalChan can be used to control process. if an impl cannot provide such control, it can just give a nil
	//   - err is runtime error
	Run(process *Process) (runID string, retChan <-chan int, signalChan chan<- Signal, err error)
	// CWL 本身并无中断执行的机制，因此只需要Run 接口即可 <- 现在需要加上这样的功能

	// QueryRuntime will return a runtime limit
	QueryRuntime(limits ResourcesLimites) Runtime
}

// 本地运行环境
type LocalExecutor struct {
}

func (exec LocalExecutor) QueryRuntime(limits ResourcesLimites) Runtime {
	return Runtime{
		Cores: int(limits.CoresMin),
	}
}

func (exe LocalExecutor) Run(process *Process) (runid string, retChan <-chan int, signalChan chan<- Signal, err error) {
	envs := process.Env()
	cmds, err := process.Command()
	// SET INPUTS
	// set stdout
	// set stdin
	// set image
	// migrate inputs

	if err = process.MigrateInputs(); err != nil {
		return "", nil, nil, err
	}
	var r *exec.Cmd
	if os.Getenv("DOCKER") != "off" {
		// if os.Getenv("DOCKER") != "" {
		// docker run
		var dockerR *cwl.DockerRequirement
		// handler docker
		dockerR = process.tool.HitsDocker()
		// 可以先检查是否支持 docker
		if req := process.tool.RequiresDocker(); req != nil {
			dockerR = req
		}
		if dockerR != nil {
			// 挂载代替拷贝？
			var dockerargs []string
			image := dockerR.DockerPull
			workdirInContainer := process.runtime.RootHost
			if dockerR.DockerOutputDirectory != "" {
				workdirInContainer = dockerR.DockerOutputDirectory
			}
			dockerargs = append(dockerargs, "run")
			for k, v := range envs {
				dockerargs = append(dockerargs, "-e", fmt.Sprintf(`%s="%s"`, k, v))
			}
			dockerargs = append(dockerargs, "-v", fmt.Sprintf(`%s:%s`, process.runtime.RootHost, workdirInContainer))
			// 用户数据文件夹映射；否则可能会出现链接文件无法访问
			userhome, _ := os.UserHomeDir()
			dockerargs = append(dockerargs, "-v", fmt.Sprintf(`%s:%s`, userhome, userhome))
			dockerargs = append(dockerargs, "-v", fmt.Sprintf(`%s:%s`, process.runtime.InputsHost, process.runtime.InputsHost))
			// 如果有父运行时，挂载
			if process.parentRuntime.RootHost != "" {
				dockerargs = append(dockerargs, "-v", fmt.Sprintf(`%s:%s`, process.parentRuntime.RootHost, process.parentRuntime.RootHost))
			}
			dockerargs = append(dockerargs, "-w", workdirInContainer)
			if process.stdin != "" {
				// dockerargs = append(dockerargs, "-a", "stdin", "-i")
				dockerargs = append(dockerargs, "-i")
			}
			dockerargs = append(dockerargs, image)
			cmds = append(dockerargs, cmds...)
			r = exec.Command("docker", cmds...)
		}
	}
	if r == nil {
		r = exec.Command(cmds[0], cmds[1:]...)
		for k, v := range envs {
			r.Env = append(r.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}
	r.Dir = process.runtime.RootHost
	// Set Std OUT ERR IN
	if process.stdout != "" {
		outpath := path.Join(process.runtime.RootHost, process.stdout)
		fout, err := os.Create(outpath)
		if err != nil {
			return "", nil, nil, err
		}
		r.Stdout = fout
	}
	if process.stderr != "" {
		errpath := path.Join(process.runtime.RootHost, process.stderr)
		ferr, err := os.Create(errpath)
		if err != nil {
			return "", nil, nil, err
		}
		r.Stderr = ferr
	}
	if process.stdin != "" {
		inPath := process.stdin
		if !path.IsAbs(inPath) {
			inPath = path.Join(process.runtime.RootHost, inPath)
		}
		fin, err := os.OpenFile(inPath, os.O_RDONLY, 0)
		if err != nil {
			return "", nil, nil, err
		}
		r.Stdin = fin
	}
	err = r.Start()
	if err != nil {
		return "", nil, nil, err
	}
	pid := r.Process.Pid
	rChan := make(chan int)
	go func() {
		r.Wait()
		rChan <- r.ProcessState.ExitCode()
		close(rChan)
	}()
	return fmt.Sprint(pid), rChan, nil, nil
}
