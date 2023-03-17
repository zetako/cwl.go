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
	Run(process *Process) (runid string, retChan <-chan int, err error)
	// CWL 本身并无中断执行的机制，因此只需要Run 接口即可
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
	if os.Getenv("DOCKER") != "" {
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
			dockerargs = append(dockerargs, image)
			cmds = append(dockerargs, cmds...)
			r := exec.Command("docker", cmds...)
			err = r.Start()
			if err != nil {
				return "", nil, err
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
	}
	r := exec.Command(cmds[0], cmds[1:]...)
	for k, v := range envs {
		r.Env = append(r.Env, fmt.Sprintf("%s=%s", k, v))
	}
	r.Dir = process.runtime.RootHost
	// Set Std OUT ERR IN
	if process.stdout != "" {
		outpath := path.Join(process.runtime.RootHost, process.stdout)
		fout, err := os.Create(outpath)
		if err != nil {
			return "", nil, err
		}
		r.Stdout = fout
	}
	if process.stderr != "" {
		errpath := path.Join(process.runtime.RootHost, process.stderr)
		ferr, err := os.Create(errpath)
		if err != nil {
			return "", nil, err
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
			return "", nil, err
		}
		r.Stdin = fin
	}
	//if process.tool.Stdout != "" {
	//	stdout , err := process.jsvm.Eval(process.tool.Stdout, nil)
	//	if err != nil {
	//		return "", nil, err
	//	}
	//	outpath, ok := stdout.(string)
	//	if !ok {
	//		return "", nil, fmt.Errorf("stdout should be string")
	//	}
	//	outpath = path.Join(process.runtime.RootHost, outpath)
	//	fout , err := os.Create(outpath)
	//	if err != nil {
	//		return "", nil, err
	//	}
	//	r.Stdout = fout
	//}
	err = r.Start()
	if err != nil {
		return "", nil, err
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
