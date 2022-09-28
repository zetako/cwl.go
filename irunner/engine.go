package irunner

import (
  "encoding/json"
  "fmt"
  "github.com/google/uuid"
  "github.com/otiai10/cwl.go"
  "os"
  "os/user"
  "sync"
)

type Engine struct {
  sync.RWMutex    `json:"-"`
  inputFS Filesystem
  outputFS Filesystem
  root *cwl.Root // root Documents.
  params *cwl.Parameters
  //runtime Runtime
  process *Process // root process
  //TaskSequence    []string            // for testing purposes
  //UnfinishedProcs map[string]bool     // engine's stack of CLT's that are running; (task.Root.ID, Process) pairs
  //FinishedProcs   map[string]bool     // engine's stack of completed processes; (task.Root.ID, Process) pairs
  //CleanupProcs    map[CleanupKey]bool // engine's stack of running cleanup processes
  UserID          string              // the userID for the user who requested the workflow run
  RunID           string              // the workflow ID
  RootHost  string
  //Manifest        *Manifest           // to pass the manifest to the gen3fuse container of each task pod
  Log             *MainLog            //
  //KeepFiles       map[string]bool     // all the paths to not delete during basic file cleanup
  //
  importer Importer
}

type EngineConfig struct {
  RunID string
  UserName string
  Importer
  InputFS       Filesystem
  OutputFS Filesystem
  Process       []byte
  Params        []byte
  ExtendConfigs map[string]interface{}
  Workdir       string
  RootHost      string
}

func initConfig(c *EngineConfig)  {
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
    wd, _ :=os.Getwd()
    c.Workdir = wd
  }
  if c.Importer == nil {
    c.Importer = &DefaultImporter{ BaseDir: c.Workdir }
  }
  if c.InputFS == nil {
    c.InputFS = NewLocal(c.Workdir)
  }
  if c.RootHost == "" ||c.RootHost == "/" {
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
    params: cwl.NewParameters(),
    RunID:           c.RunID,
    UserID:          c.UserName,
    RootHost:        c.RootHost,
    Log:             &MainLog{
      ProcessRequest: ProcessRequest{
        Process: c.Process,
        Params: c.Params,
      },
      Log:      logger(),
    },
    inputFS: c.InputFS,
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

func (e *Engine) MainProcess() (*Process, error) {
  if e.process != nil {
    return e.process, nil
  }
  process := &Process{
    tool: e.root,
    inputs: e.params,
    //runtime: e.runtime,
    fs: e.inputFS,
    env: map[string]string{},
    Log: e.Log.Log,
  }
  setDefault(e.params, e.root.Inputs)
  
  if err := process.initJVM(); err != nil {
    return nil, err
  }
  process.runtime.RootHost = e.RootHost
  process.loadRuntime()
  
  // Bind inputs to values.
  //
  // Since every part of a tool depends on "inputs" being available to expressions,
  // nothing can be done on a Process without a valid inputs binding,
  // which is why we bind in the Process constructor.
  for _, in := range e.root.Inputs {
    val := (*e.params)[in.ID]
    k := sortKey{getPos(in.Binding)}
    b, err := process.bindInput(in.ID, in.Types, in.Binding, in.SecondaryFiles, val, k)
    if err != nil {
      return nil, e.errorf("binding input %q: %s", in.ID, err)
    }
    if b == nil {
      return nil, e.errorf("no binding found for input: %s", in.ID)
    }
    
    process.bindings = append(process.bindings, b...)
  }
  e.process = process
  
  err := process.loadReqs()
  if err != nil {
    return nil, err
  }
  
  {
    stdoutI, err := process.eval(process.tool.Stdout, nil)
    if err != nil {
      return nil, fmt.Errorf("evaluating stdout expression : %s", err)
    }
  
    stderrI, err := process.eval(process.tool.Stderr, nil)
    if err != nil {
      return nil, fmt.Errorf("evaluating stderr expression : %s", err)
    }
  
    var stdoutStr, stderrStr string
    var ok bool
  
    if stdoutI != nil {
      stdoutStr, ok = stdoutI.(string)
      if !ok {
        return nil, fmt.Errorf("stdout expression returned a non-string value")
      }
    }
  
    if stderrI != nil {
      stderrStr, ok = stderrI.(string)
      if !ok {
        return nil, fmt.Errorf("stderr expression returned a non-string value")
      }
    }
  
    for _, out := range process.tool.Outputs {
      if len(out.Types) == 1 {
        if out.Types[0].Type == "stdout" && stdoutStr == "" {
          stdoutStr = "stdout-" + uuid.New().String()
        }
        if out.Types[0].Type == "stderr" && stderrStr == "" {
          stderrStr = "stderr-" + uuid.New().String()
        }
      }
    }
    process.stdout = stdoutStr
    process.stderr = stderrStr
  }

  return process, nil
}

func (p *Process) loadRuntime() {
  //e.runtime.Cores = "2"
  // TODO More
  if res := p.tool.Requirements.RequiresResource(); res != nil {
    p.runtime.Cores = fmt.Sprint( res.CoresMin)
  }
  p.vm.Set("runtime", p.runtime)
}

