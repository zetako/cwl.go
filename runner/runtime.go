package runner

type Mebibyte int

type ResourcesLimites struct {
  CoresMin,
  CoresMax int64
  
  RAMMin,
  RAMMax,
  OutdirMin,
  OutdirMax,
  TmpdirMin,
  TmpdirMax Mebibyte
}

func GetDefaultResourcesLimits() ResourcesLimites {
  return ResourcesLimites{
    CoresMin: 1,
    RAMMin: 256,
    OutdirMin: 1024,
    TmpdirMin: 1024,
  }
}

// TODO this is provided to expressions early on in process processing,
//  but it won't have real values from a scheduler until much later.
type Runtime struct {
  RootHost   string `json:"rootHost"`
  ExitCode *int `json:"exitCode"`
  Outdir string `json:"outdir"`
  Tmpdir string `json:"tmpdir"`
  //  The reported number of CPU cores reserved for the process,
  //        which is available to expressions on the CommandLineTool as
  //        `runtime.cores`, must be a non-zero integer, and may be
  //        calculated by rounding up the cores request to the next whole
  //        number.
  Cores      int `json:"cores"`
  // `runtime.ram`, must be a non-zero integer. RAM in mebibytes (2**20) (default is 256)
  RAM        Mebibyte `json:"ram"`
  // `runtime.tmpdirSize`, must be a non-zero integer. reserved filesystem based storage for the designated temporary directory, in mebibytes (2**20) (default is 1024)
  TmpdirSize Mebibyte `json:"tmpdirSize"`
  // `runtime.outdirSize`, must be a non-zero integer. reserved filesystem based storage for the designated output directory, in mebibytes (2**20) (default is 1024)
  OutdirSize Mebibyte `json:"outdirSize"`
}

var defaultRuntime = Runtime{
  Tmpdir: "/tmp",
  Cores: 1,
  RAM: 256,
  TmpdirSize: 1024,
  OutdirSize: 1024,
}

func (p *Process) SetRuntime(config Runtime) {
  
  if config.RootHost != "" {
    p.runtime.RootHost = config.RootHost
  }
  if config.Outdir != "" {
    p.runtime.Outdir = config.Outdir
  }
  if config.Tmpdir != "" {
    p.runtime.Tmpdir = config.Tmpdir
  }
  if config.Cores >= 0 {
    p.runtime.Cores = config.Cores
  }
  if config.RAM >= 0 {
    p.runtime.RAM = config.RAM
  }
  if config.TmpdirSize >= 0 {
    p.runtime.TmpdirSize = config.TmpdirSize
  }
  if config.OutdirSize >= 0 {
    p.runtime.OutdirSize = config.OutdirSize
  }
  if config.ExitCode != nil {
    p.runtime.ExitCode = config.ExitCode
  }
}