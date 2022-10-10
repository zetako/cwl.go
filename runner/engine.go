package runner

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/lijiang2014/cwl.go"
	"os"
	"os/user"
	"sync"
)

// Engine
// 工作流的解析引擎，管理一个工作流的解析过程
type Engine struct {
	sync.RWMutex `json:"-"`
	// 配置接口
	importer Importer
	//
	inputFS      Filesystem
	outputFS     Filesystem
	root         *cwl.Root // root Documents.
	params       *cwl.Values
	//runtime Runtime
	process *Process // root process
	UserID   string // the userID for the user who requested the workflow run
	RunID    string // the workflow ID
	RootHost string
	Log *MainLog //
}

type EngineConfig struct {
	RunID    string
	UserName string
	Importer
	InputFS       Filesystem
	OutputFS      Filesystem
	Process       []byte
	Params        []byte
	ExtendConfigs map[string]interface{}
	Workdir       string
	RootHost      string
}

func initConfig(c *EngineConfig) {
	if c.RunID == "" {
		c.RunID = "testcwl"
	}
	if c.UserName == "" {
		u, _ := user.Current()
		if u != nil {
			c.UserName = u.Name
		} else {
			c.UserName = "unknown"
		}
	}
	if c.Workdir == "" {
		wd, _ := os.Getwd()
		c.Workdir = wd
	}
	if c.Importer == nil {
		c.Importer = &DefaultImporter{BaseDir: c.Workdir}
	}
	if c.InputFS == nil {
		c.InputFS = NewLocal(c.Workdir)
	}
	if c.RootHost == "" || c.RootHost == "/" {
		c.RootHost = "/tmp/" + c.RunID
	}
	if c.OutputFS == nil {
		c.OutputFS = NewLocal(c.RootHost)
	}

}

// Engine runs an instance of the mariner engine job
func NewEngine(c EngineConfig) (*Engine, error) {
	initConfig(&c)
	e := &Engine{
		params:   cwl.NewValues(),
		RunID:    c.RunID,
		UserID:   c.UserName,
		RootHost: c.RootHost,
		Log: &MainLog{
			ProcessRequest: ProcessRequest{
				Process: c.Process,
				Params:  c.Params,
			},
			Log: logger(),
		},
		inputFS:  c.InputFS,
		outputFS: c.OutputFS,
		importer: c.Importer,
	}
	if err := json.Unmarshal(c.Params, &e.params); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(c.Process, &e.root); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *Engine) RunCommandLine(process *Process, params cwl.Values) (outs cwl.Values ,err error){
	err = e.ResolveProcess(process, params)
	return
}


// 解析但不执行
func (e *Engine) ResolveProcess(process *Process, params cwl.Values) ( error){
	tool , ok := process.tool.Process.(*cwl.CommandLineTool)
	if !ok {
		return  e.errorf("need to be CommandLineTool %s", process.tool.Process.Base().ID)
	}
	setDefault(e.params, tool.Inputs)
	if err := process.initJVM(); err != nil {
		return  err
	}
	// TODO better
	process.runtime.RootHost = e.RootHost
	process.loadRuntime()
	// Bind inputs to values.
	//
	// Since every part of a tool depends on "inputs" being available to expressions,
	// nothing can be done on a ProcessBase without a valid inputs binding,
	// which is why we bind in the ProcessBase constructor.
	for _, inb := range tool.Inputs {
		in := inb.(*cwl.CommandInputParameter)
		val := (*e.params)[in.ID]
		k := sortKey{getPos(in.InputBinding)}
		b, err := process.bindInput(in.ID, in.Type.SaladType, in.InputBinding, in.SecondaryFiles, val, k)
		if err != nil {
			return  e.errorf("binding input %q: %s", in.ID, err)
		}
		if b == nil {
			return e.errorf("no binding found for input: %s", in.ID)
		}
		
		process.bindings = append(process.bindings, b...)
	}
	e.process = process
	
	err := process.loadReqs()
	if err != nil {
		return  err
	}
	
	{
		stdoutI, err := process.eval(tool.Stdout, nil)
		if err != nil {
			return fmt.Errorf("evaluating stdout expression : %s", err)
		}
		
		stderrI, err := process.eval(tool.Stderr, nil)
		if err != nil {
			return fmt.Errorf("evaluating stderr expression : %s", err)
		}
		
		var stdoutStr, stderrStr string
		var ok bool
		
		if stdoutI != nil {
			stdoutStr, ok = stdoutI.(string)
			if !ok {
				return  fmt.Errorf("stdout expression returned a non-string value")
			}
		}
		
		if stderrI != nil {
			stderrStr, ok = stderrI.(string)
			if !ok {
				return fmt.Errorf("stderr expression returned a non-string value")
			}
		}
		
		for _, out := range tool.Outputs {
			outi :=out.(*cwl.CommandOutputParameter)
			if outi.Type.TypeName() == "stdout"  {
				stdoutStr = "stdout-" + uuid.New().String()
			} else if outi.Type.TypeName() == "stderr" {
				stderrStr = "stderr-" + uuid.New().String()
			}
		}
		process.stdout = stdoutStr
		process.stderr = stderrStr
	}
	return nil
}

func (e *Engine) MainProcess() (*Process, error) {
	if e.process != nil {
		return e.process, nil
	}
	process := &Process{
		tool:   e.root,
		inputs: e.params,
		//runtime: e.runtime,
		fs:  e.inputFS,
		env: map[string]string{},
		Log: e.Log.Log,
	}
	//switch expr {
	//
	//}
	inputs := e.root.Process.Base().Inputs
	setDefault(e.params, inputs)

	if err := process.initJVM(); err != nil {
		return nil, err
	}
	process.runtime.RootHost = e.RootHost
	process.loadRuntime()
	return process, nil
}

func (p *Process) loadRuntime() {
	////e.runtime.Cores = "2"
	//// TODO More
	//if res := p.tool.Requirements.RequiresResource(); res != nil {
	//	p.runtime.Cores = fmt.Sprint(res.CoresMin)
	//}
	p.vm.Set("runtime", p.runtime)
}
