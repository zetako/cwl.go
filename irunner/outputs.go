package irunner

import (
  "encoding/json"
  "fmt"
  "github.com/otiai10/cwl.go"
  "github.com/spf13/cast"
  "reflect"
)

func (e *Engine) Outputs() (cwl.Values, error) {
  return e.process.Outputs(e.outputFS)
}

// Outputs binds cwl.Tool output descriptors to concrete values.
func (process *Process) Outputs(fs Filesystem) (cwl.Values, error) {
  outdoc, err := fs.Contents("cwl.output.json")
  if err == nil {
    outputs := cwl.Values{}
    err = json.Unmarshal([]byte(outdoc), &outputs)
    return outputs, err
  }
  
  outputs := cwl.Values{}
  for _, out := range process.tool.Outputs {
    v, err := process.bindOutput(fs, out.Types, out.Binding, out.SecondaryFiles, nil)
    if err != nil {
      return nil, fmt.Errorf(`failed to bind value for "%s": %s`, out.ID, err)
    }
    outputs[out.ID] = v
  }
  return outputs, nil
}

// bindOutput binds the output value for a single CommandOutput.
func (process *Process) bindOutput(
  fs Filesystem,
  types []cwl.Type,
  binding *cwl.Binding,
  secondaryFiles []cwl.SecondaryFile,
  val interface{},
) (interface{}, error) {
  var err error
  
  if binding != nil && len(binding.Glob) > 0 {
    // glob patterns may be expressions. evaluate them.
    globs, err := process.evalGlobPatterns(binding.Glob)
    if err != nil {
      return nil, fmt.Errorf("failed to evaluate glob expressions: %s", err)
    }
    
    files, err := process.matchFiles(fs, globs, binding.LoadContents)
    if err != nil {
      return nil, fmt.Errorf("failed to match files: %s", err)
    }
    val = files
  }
  
  if binding != nil && binding.Eval != nil && binding.Eval.Raw != "" {
    val, err = process.eval(binding.Eval.Raw, val)
    if err != nil {
      return nil, fmt.Errorf("failed to evaluate outputEval: %s", err)
    }
  }
  
  if val == nil {
    for _, t := range types {
      if t.Type == "null" {
        return nil, nil
      }
    }
  }
  
  for _, t := range types {
    switch t.Type {
    // TODO validate stdout/err can only be at root
    //      validate that stdout/err doesn't occur more than once
    case "stdout":
      files, err := process.matchFiles(fs, []string{process.stdout}, false)
      if err != nil {
        return nil, fmt.Errorf("failed to match files: %s", err)
      }
      if len(files) == 0 {
        return nil, fmt.Errorf(`failed to match stdout file "%s"`, process.stdout)
      }
      if len(files) > 1 {
        return nil, fmt.Errorf(`matched multiple stdout files "%s"`, process.stdout)
      }
      return files[0], nil
    
    case "stderr":
      files, err := process.matchFiles(fs, []string{process.stderr}, false)
      if err != nil {
        return nil, fmt.Errorf("failed to match files: %s", err)
      }
      if len(files) == 0 {
        return nil, fmt.Errorf(`failed to match stderr file "%s"`, process.stderr)
      }
      if len(files) > 1 {
        return nil, fmt.Errorf(`matched multiple stderr files "%s"`, process.stderr)
      }
      return files[0], nil
    }
  }
  
  if val == nil {
    return nil, fmt.Errorf("missing value")
  }
  
  // Bind the output value to one of the allowed types.
Loop:
  for _, t := range types {
    switch t.Type {
    case "bool":
      v, err := cast.ToBoolE(val)
      if err == nil {
        return v, nil
      }
    case "int":
      v, err := cast.ToInt32E(val)
      if err == nil {
        return v, nil
      }
    case "long":
      v, err := cast.ToInt64E(val)
      if err == nil {
        return v, nil
      }
    case "float":
      v, err := cast.ToFloat32E(val)
      if err == nil {
        return v, nil
      }
    case "double":
      v, err := cast.ToFloat64E(val)
      if err == nil {
        return v, nil
      }
    case "string":
      v, err := cast.ToStringE(val)
      if err == nil {
        return v, nil
      }
    case "File":
      switch y := val.(type) {
      case []cwl.FileDir:
        if len(y) != 1 {
          continue Loop
        }
        f := y[0]
        for _, expr := range secondaryFiles {
          err := process.resolveSecondaryFiles(f, expr)
          if err != nil {
            return nil, fmt.Errorf("resolving secondary files: %s", err)
          }
        }
        return f, nil
      
      case cwl.File:
        return y, nil
      default:
        continue Loop
      }
    case "Directory":
      // TODO
    case "array":
      typ := reflect.TypeOf(val)
      if typ.Kind() != reflect.Slice {
        continue Loop
      }
      
      var res []interface{}
      
      arr := reflect.ValueOf(val)
      for i := 0; i < arr.Len(); i++ {
        item := arr.Index(i)
        if !item.CanInterface() {
          return nil, fmt.Errorf("can't get interface of array item")
        }
        r, err := process.bindOutput(fs, t.Items, t.Binding, nil, item.Interface())
        if err != nil {
          return nil, err
        }
        res = append(res, r)
      }
      return res, nil
    
    case "record":
      // TODO
      
    }
  }
  
  return nil, fmt.Errorf("no type could be matched")
}

// matchFiles executes the list of glob patterns, returning a list of matched files.
// matchFiles must return a non-nil list on success, even if no files are matched.
func (process *Process) matchFiles(fs Filesystem, globs []string, loadContents bool) ([]cwl.FileDir, error) {
  // it's important this slice isn't nil, because the outputEval field
  // expects it to be non-null during expression evaluation.
  files := []cwl.FileDir{}
  
  // resolve all the globs into file objects.
  for _, pattern := range globs {
    matches, err := fs.Glob(pattern)
    if err != nil {
      return nil, fmt.Errorf("failed to execute glob: %s", err)
    }
    
    for _, m := range matches {
      // TODO handle directories
      v := cwl.FileDir{
        Class: "File",
        Location: m.Location,
        Path:     m.Path,
        File: cwl.File{
          Checksum: m.Checksum,
          Size:     m.Size,
        },
      }
      
      f, err := process.resolveFile(v, loadContents)
      if err != nil {
        return nil, err
      }
      files = append(files, f)
    }
  }
  return files, nil
}


// evalGlobPatterns evaluates a list of potential expressions as defined by the CWL
// OutputBinding.glob field. It returns a list of strings, which are glob expression,
// or an error.
//
// cwl spec:
// "If an expression is provided, the expression must return a string or an array
//  of strings, which will then be evaluated as one or more glob patterns."
func (process *Process) evalGlobPatterns(patterns []string) ([]string, error) {
  var out []string
  
  for _, pattern := range patterns {
    // TODO what is "self" here?
    val, err := process.eval(pattern, nil)
    if err != nil {
      return nil, err
    }
    
    switch z := val.(type) {
    case string:
      out = append(out, z)
    case []cwl.Value:
      for _, val := range z {
        z, ok := val.(string)
        if !ok {
          return nil, fmt.Errorf(
            "glob expression returned an invalid type. Only string or []string "+
              "are allowed. Got: %#v", z)
        }
        out = append(out, z)
      }
    default:
      return nil, fmt.Errorf(
        "glob expression returned an invalid type. Only string or []string "+
          "are allowed. Got: %#v", z)
    }
  }
  return out, nil
}


