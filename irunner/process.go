package irunner

import (
  "encoding/json"
  "fmt"
  "github.com/otiai10/cwl.go"
  "sort"
  "strings"
)

type Process struct {
  tool           *cwl.Root
  inputs         *cwl.Parameters
  runtime       Runtime
  fs             Filesystem
  bindings       []*Binding
  expressionLibs []string
  env            map[string]string
  filesToCreate []cwl.Entry
  shell          bool
  resources      Resources
  stdout         string
  stderr         string
  *Log
  *jsvm
}

type Mebibyte int

// TODO this is provided to expressions early on in process processing,
//      but it won't have real values from a scheduler until much later.
type Runtime struct {
  Outdir string
  Tmpdir string
  // TODO make these all strings?
  RootHost   string `json:"rootHost"`
  Cores      string `json:"cores"`
  RAM        Mebibyte
  OutdirSize Mebibyte
  TmpdirSize Mebibyte
}

type Resources struct {
  CoresMin,
  CoresMax int
  
  RAMMin,
  RAMMax,
  OutdirMin,
  OutdirMax,
  TmpdirMin,
  TmpdirMax Mebibyte
}


// Binding binds an input type description (string, array, record, etc)
// to a concrete input value. this information is used while building
// command line args.
type Binding struct {
  clb *cwl.Binding
  // the bound type (resolved by matching the input value to one of many allowed types)
  // can be nil, which means no matching type could be determined.
  Type cwl.Type
  // the value from the input object
  Value cwl.Parameter
  // used to determine the ordering of command line flags.
  // http://www.commonwl.org/v1.0/CommandLineTool.html#Input_binding
  sortKey sortKey
  nested  []*Binding
  name    string
}

type sortKey []interface{}



func setDefault(values *cwl.Parameters, inputs cwl.Inputs) {
  for _, in := range inputs {
    _, ok := (*values)[in.ID]
    if !ok && in.Default != nil {
      (*values)[in.ID] = in.Default
    }
  }
}

func (process *Process) Command() ([]string, error) {
  // Copy "Tool.Inputs" bindings
  args := make([]*Binding, 0, len(process.bindings))
  for _, b := range process.bindings {
    if b.clb != nil {
      args = append(args, b)
    }
  }
  
  // Add "Tool.Arguments"
  for i, arg := range process.tool.Arguments {
    if arg.Binding == nil && arg.Value != "" {
      args = append(args, &Binding{
        arg.Binding, cwl.Type{
          Type: argType, // #TODO Check if Type ok
        }, arg.Value, sortKey{0}, nil, "",
      })
      continue
    } else if arg.Binding == nil {
      return nil, fmt.Errorf("empty argument")
    }
    if arg.Binding != nil && arg.Binding.ValueFrom != nil &&  arg.Binding.ValueFrom.String() == "" {
      return nil, fmt.Errorf("valueFrom is required but missing for argument %d", i)
    }
    args = append(args, &Binding{
      arg.Binding, cwl.Type{
        Type: argType, // #TODO Check if Type ok
      }, nil, sortKey{arg.Binding.Position}, nil, "",
    })
  }
  
  // Evaluate "valueFrom" expression.
  for _, b := range args {
    if b.clb != nil && b.clb.ValueFrom != nil &&  b.clb.ValueFrom.String() != "" {
      val, err := process.eval(b.clb.ValueFrom.String(), b.Value)
      if err != nil {
        return nil, fmt.Errorf("failed to eval argument value: %s", err)
      }
      b.Value = val
    }
  }
  
  sort.Stable(bySortKey(args))
  //debug(args)
  
  // Now collect the input bindings into command line arguments
  cmd := append([]string{}, process.tool.BaseCommands...)
  for _, b := range args {
    cmd = append(cmd, bindArgs(b)...)
  }
  
  if process.tool.RequiresShellCommand() {
    cmd = []string{"/bin/sh", "-c", strings.Join(cmd, " ")}
  }
  
  //debug("COMMAND", cmd)
  return cmd, nil
}


func toJSONMap(v interface{}) (interface{}, error) {
  if v == nil {
    return nil, nil
  }
  
  // Need to convert Go variable naming to JSON. Easiest way to to marshal to JSON,
  // then unmarshal into a map.
  j, err := json.Marshal(v)
  if err != nil {
    return nil, err
  }
  
  var data interface{}
  err = json.Unmarshal(j, &data)
  if err != nil {
    return nil, fmt.Errorf(`marshaling data for JS evaluation %s`, err)
  }
  return data, nil
}

func (process *Process) eval(x string, self interface{}) (interface{}, error) {
  return process.jsvm.Eval(x, self)
}

func (process *Process) Env() map[string]string {
  env := map[string]string{}
  for k, v := range process.env {
    env[k] = v
  }
  return env
}

func (process *Process) loadReqs() error {
  //if req := process.tool.RequiresInlineJavascript(); req != nil {
  //  if err := process.initJVM(); err != nil {
  //    return err
  //  }
  //}
  if req := process.tool.RequiresEnvVar(); req != nil {
    // TODO env 
  }
  if req := process.tool.RequiresResource(); req != nil {
    process.loadRuntime()
  }
  if req := process.tool.RequiresSchemaDef(); req != nil {
    // TODO env 
  }
  if req := process.tool.RequiresInitialWorkDir(); req != nil {
    // TODO env 
  }
  return nil
}