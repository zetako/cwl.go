package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"os"
	"os/user"
	"path"
	"sync"

	"github.com/google/uuid"
)

// Engine
// 工作流的解析引擎，管理一个工作流的解析过程
type Engine struct {
	sync.RWMutex `json:"-"`
	// 配置接口
	importer Importer
	executor Executor
	//
	inputFS  Filesystem
	outputFS Filesystem
	root     *cwl.Root // root Documents.
	params   *cwl.Values
	MessageReceiver
	SignalChannel chan Signal
	//runtime Runtime
	process    *Process // root process
	UserID     string   // the userID for the user who requested the workflow run
	RunID      string   // the workflow ID
	RootHost   string
	InputsHost string
	Log        *MainLog //
	// executor
}

type EngineConfig struct {
	DocumentID string
	RunID      string
	UserName   string
	Importer
	InputFS       Filesystem
	OutputFS      Filesystem
	Process       []byte
	Params        []byte
	ExtendConfigs map[string]interface{}
	DocImportDir  string
	RootHost      string
	InputsDir     string
	WorkDir       string
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
	if c.DocImportDir == "" {
		wd, _ := os.Getwd()
		c.DocImportDir = wd
	}
	if c.Importer == nil {
		c.Importer = &DefaultImporter{BaseDir: c.DocImportDir}
	}
	if c.InputFS == nil {
		c.InputFS = NewLocal(c.DocImportDir)
	}
	if c.RootHost == "" || c.RootHost == "/" {
		c.RootHost = "/tmp/" + c.RunID
	}
	if !path.IsAbs(c.WorkDir) {
		c.WorkDir = path.Join(c.RootHost, c.WorkDir)
	}
	if !path.IsAbs(c.InputsDir) {
		c.InputsDir = path.Join(c.RootHost, c.InputsDir)
	}
	if c.OutputFS == nil {
		c.OutputFS = NewLocal(c.WorkDir)
		c.OutputFS.(*Local).CalcChecksum = true
	}

}

// Engine runs an instance of the mariner engine job
func NewEngine(c EngineConfig) (*Engine, error) {
	var err error
	initConfig(&c)
	e := &Engine{
		params:     cwl.NewValues(),
		RunID:      c.RunID,
		UserID:     c.UserName,
		RootHost:   c.WorkDir,
		InputsHost: c.InputsDir,
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
	// import Doc
	if c.Process, err = e.EnsureImportedDoc(c.Process); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(c.Process, &e.root); err != nil {
		if err.Error() == "no Process name" {
			// parse $graph
			if c.DocumentID == "" {
				return nil, errors.New("no Process name and Document name not specified")
			}
			graphs := cwl.SaladRootDoc{}
			err = json.Unmarshal(c.Process, &graphs)
			if err != nil {
				return nil, err
			}
			for _, graph := range graphs.Graph {
				if graph.Process.Base().ID == c.DocumentID {
					e.root = graph
					e.root.SaladRootDoc = graphs // 这一行用来保存其他已解析的文件
					return e, nil
				}
			}
			return nil, errors.New("no matched document found")
		}
		return nil, err
	}
	// set other defaults, can be changed later
	e.MessageReceiver = DefaultMsgReceiver{}
	e.SignalChannel = make(chan Signal)
	return e, nil
}

func (e *Engine) RunCommandLine(process *Process, params cwl.Values) (outs cwl.Values, err error) {
	process.inputs = &params
	err = e.ResolveProcess(process)
	return
}

func (e *Engine) SetDefaultExecutor(exec Executor) {
	e.executor = exec
}

func (e *Engine) RunProcess(p *Process) (outs cwl.Values, err error) {
	if _, isCLT := p.root.Process.(*cwl.CommandLineTool); isCLT {
		limits, err := p.ResourcesLimites()
		runtime := e.executor.QueryRuntime(*limits)
		p.SetRuntime(runtime)
		err = e.ResolveProcess(p)
		if err != nil {
			return nil, err
		}
		pid, ret, err := e.executor.Run(p)
		if err != nil {
			return nil, err
		}
		p.JobID = pid
		retCode, _ := <-ret
		p.SetRuntime(Runtime{ExitCode: &retCode})
		if p.outputFS == nil {
			p.outputFS = e.outputFS
		}
		outputs, err := p.Outputs(p.outputFS)
		return outputs, err
	} else if _, isExpTool := p.root.Process.(*cwl.ExpressionTool); isExpTool {
		outputs, err := p.RunExpression()
		return outputs, err
	} else if _, isWorkflow := p.root.Process.(*cwl.Workflow); isWorkflow {
		outputs, err := p.RunWorkflow(e)
		return outputs, err
	}
	return nil, fmt.Errorf("unknown process class  ")
}

func (e *Engine) Run() (outs cwl.Values, err error) {
	_, err = e.MainProcess()
	if err != nil {
		return nil, err
	}
	return e.RunProcess(e.process)
}

// 解析但不执行
func (e *Engine) ResolveProcess(process *Process) error {
	tool, ok := process.root.Process.(*cwl.CommandLineTool)
	params := process.inputs
	if !ok {
		return e.errorf("need to be CommandLineTool %s", process.root.Process.Base().ID)
	}
	setDefault(params, tool.Inputs)
	if err := process.initJVM(); err != nil {
		return err
	}
	//
	if process.runtime.RootHost == "" {
		process.runtime.RootHost = e.RootHost
	}
	if process.runtime.InputsHost == "" {
		process.runtime.InputsHost = e.InputsHost
	}
	//process.runtime.RootHost = e.RootHost
	//process.runtime.InputsHost = e.InputsHost

	process.loadRuntime()
	// Bind inputs to values.
	//
	// Since every part of a tool depends on "inputs" being available to expressions,
	// nothing can be done on a ProcessBase without a valid inputs binding,
	// which is why we bind in the ProcessBase constructor.
	for _, inb := range tool.Inputs {
		in := inb.(*cwl.CommandInputParameter)
		val := (*params)[in.ID]
		k := sortKey{getPos(in.InputBinding)}
		if val == nil {
			process.bindings = append(process.bindings, &Binding{name: in.ID, Value: nil})
			continue
		}
		b, err := process.bindInput(in.ID, in.Type.SaladType, in.InputBinding, in.SecondaryFiles, val, k)
		if err != nil {
			return e.errorf("binding input %q: %s", in.ID, err)
		}
		if b == nil {
			return e.errorf("no binding found for input: %s", in.ID)
		}

		process.bindings = append(process.bindings, b...)
	}
	e.process = process

	err := process.loadReqs()
	if err != nil {
		return err
	}
	if err := process.initJVM(); err != nil {
		return err
	}

	{ // set stdout stderr
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
				return fmt.Errorf("stdout expression returned a non-string value")
			}
		}

		if stderrI != nil {
			stderrStr, ok = stderrI.(string)
			if !ok {
				return fmt.Errorf("stderr expression returned a non-string value")
			}
		}

		for _, out := range tool.Outputs {
			outi := out.(*cwl.CommandOutputParameter)
			if outi.Type.TypeName() == "stdout" {
				if stdoutStr == "" {
					stdoutStr = "stdout-" + uuid.New().String()
				}
			} else if outi.Type.TypeName() == "stderr" {
				if stderrStr == "" {
					stderrStr = "stderr-" + uuid.New().String()
				}
			}
		}
		process.stdout = stdoutStr
		process.stderr = stderrStr
		{
			stdinI, err := process.eval(tool.Stdin, nil)
			if err != nil {
				return fmt.Errorf("evaluating stdin expression : %s %s", tool.Stdin, err)
			}
			if stdinI != nil {
				stdinStr, ok := stdinI.(string)
				if !ok {
					return fmt.Errorf("stdin expression returned a non-string value")
				}
				process.stdin = stdinStr
			}

		}
	}
	{ // set env
		if req := tool.RequiresEnvVar(); req != nil {
			for _, env := range req.EnvDef {
				envExp, err := process.Eval(env.EnvValue, nil)
				if err != nil {
					return err
				}
				process.env[env.EnvName] = fmt.Sprint(envExp)
			}
		}

	}
	return nil
}

func (e *Engine) MainProcess() (*Process, error) {
	if e.process != nil {
		return e.process, nil
	}
	process := &Process{
		root:   e.root,
		inputs: e.params,
		//runtime: e.runtime,
		fs:  e.inputFS,
		env: map[string]string{},
		Log: e.Log.Log,
	}
	process.SetRuntime(defaultRuntime)
	//switch expr {
	//
	//}
	if tool, ok := e.root.Process.(*cwl.CommandLineTool); ok {
		process.tool = tool
	}
	inputs := e.root.Process.Base().Inputs
	setDefault(process.inputs, inputs)

	if err := process.initJVM(); err != nil {
		return nil, err
	}

	process.runtime.RootHost = e.RootHost
	process.loadRuntime()
	e.process = process
	return process, nil
}

func (e *Engine) GenerateSubProcess(step *cwl.WorkflowStep) (process *Process, err error) {
	// 初始化
	process = &Process{
		fs:  e.inputFS,
		env: map[string]string{},
		Log: e.Log.Log,
	}

	if step.Run.Process != nil {
		if process.root == nil {
			process.root = &cwl.Root{}
		}
		process.root.Process = step.Run.Process
	} else { // 没有已序列化的进程，此时需要读文件
		if len(step.Run.ID) <= 0 {
			return nil, fmt.Errorf("no Process or Run.ID to use")
		}
		cwlFile := step.Run.ID
		if cwlFile[0] == '#' { // 先判断是否#开头，此时是graph模式
			if e.root.SaladRootDoc.Graph == nil {
				return nil, fmt.Errorf("specify a packed doc but not found")
			}
			cwlFile = cwlFile[1:]
			for _, graph := range e.root.SaladRootDoc.Graph {
				if graph.Process.Base().ID == cwlFile {
					process.root = graph
					break
				}
			}
			if process.root == nil {
				return nil, fmt.Errorf("specify a packed doc but not found")
			}
		} else { // 否则就应该去读cwl
			if len(cwlFile) <= 4 || cwlFile[len(cwlFile)-4:] != ".cwl" {
				return nil, errors.New("Run.ID not a cwl file")
				// TODO 可能还需要考虑带#的情况
			}
			// 读文件
			cwlFileReader, err := e.importer.Load(cwlFile)
			if err != nil {
				return nil, err
			}
			cwlFileJSON, err := cwl.Y2J(cwlFileReader)
			if err != nil {
				return nil, err
			}
			cwlFileJSON, err = e.EnsureImportedDoc(cwlFileJSON)
			if err != nil {
				return nil, err
			}

			// 生成
			if err = json.Unmarshal(cwlFileJSON, &process.root); err != nil {
				return nil, err
			}
		}
	}

	// 其他处理（来自MainProcess）
	process.SetRuntime(defaultRuntime)
	process.runtime.RootHost = path.Join(e.RootHost, step.ID)
	process.outputFS = &Local{
		workdir:      process.runtime.RootHost,
		CalcChecksum: true,
	}
	process.runtime.InputsHost = path.Join(e.InputsHost, step.ID)
	process.inputFS = &Local{
		workdir:      process.runtime.InputsHost,
		CalcChecksum: true,
	}
	if tool, ok := process.root.Process.(*cwl.CommandLineTool); ok {
		process.tool = tool
	}
	inputs := process.root.Process.Base().Inputs
	process.inputs = &cwl.Values{}
	setDefault(process.inputs, inputs)

	if err := process.initJVM(); err != nil {
		return nil, err
	}
	process.loadRuntime()
	return
}
