package irunner

import (
  "github.com/otiai10/cwl.go"
  "github.com/spf13/cast"
  "path/filepath"
)

func (process *Process) bindInput(
  name string,
  types []cwl.Type,
  clb *cwl.Binding,
  secondaryFiles []cwl.SecondaryFile,
  //val interface{},
  val cwl.Parameter,
  key sortKey,
) ([]*Binding, error) {
  
  // If no value was found, check if the type is allowed to be null.
  // If so, return a binding, otherwise fail.
  if val == nil {
    for _, t := range types {
      if t.Type == "null" {
        return []*Binding{
          {clb, t, nil, key, nil, name},
        }, nil
      }
    }
    return nil, process.error("missing value")
  }

Loop:
  
  // An input descriptor describes multiple allowed types.
  // Loop over the types, looking for the best match for the given input value.
  for _, t := range types {
    switch t.Type {
    
    case "array":
      vals, ok := val.([]cwl.Parameter)
      if !ok {
        // input value is not an array.
        continue Loop
      }
      
      // The input array is allowed to be empty,
      // so this must be a non-nil slice.
      out := []*Binding{}
      
      for i, val := range vals {
        subkey := append(key, sortKey{getPos(t.Binding), i}...)
        b, err := process.bindInput("", t.Items, t.Binding, nil, val, subkey)
        if err != nil {
          return nil, err
        }
        if b == nil {
          // array item values did not bind to the array descriptor.
          continue Loop
        }
        out = append(out, b...)
      }
      
      nested := make([]*Binding, len(out))
      copy(nested, out)
      b := &Binding{clb, t, val, key, nested, name}
      // TODO revisit whether creating a nested tree (instead of flat) is always better/ok
      return []*Binding{b}, nil
    
    case "record":
      vals, ok := val.(map[string]cwl.Parameter)
      if !ok {
        // input value is not a record.
        continue Loop
      }
      
      var out []*Binding
      
      for i, field := range t.Fields {
        val, ok := vals[field.Name]
        // TODO lower case?
        if !ok {
          continue Loop
        }
        
        subkey := append(key, sortKey{getPos(field.Binding), i}...)
        b, err := process.bindInput(field.Name, field.Types, field.Binding, nil, val, subkey)
        if err != nil {
          return nil, err
        }
        if b == nil {
          continue Loop
        }
        out = append(out, b...)
      }
      
      if out != nil {
        nested := make([]*Binding, len(out))
        copy(nested, out)
        b := &Binding{clb, t, val, key, nested, name}
        out = append(out, b)
        return out, nil
      }
    
    case "any":
      return []*Binding{
        {clb, t, val, key, nil, name},
      }, nil
    
    case "bool":
      v, err := cast.ToBoolE(val)
      if err != nil {
        continue Loop
      }
      return []*Binding{
        {clb, t, v, key, nil, name},
      }, nil
    
    case "int":
      v, err := cast.ToInt32E(val)
      if err != nil {
        continue Loop
      }
      return []*Binding{
        {clb, t, v, key, nil, name},
      }, nil
    
    case "long":
      v, err := cast.ToInt64E(val)
      if err != nil {
        continue Loop
      }
      return []*Binding{
        {clb, t, v, key, nil, name},
      }, nil
    
    case "float":
      v, err := cast.ToFloat32E(val)
      if err != nil {
        continue Loop
      }
      return []*Binding{
        {clb, t, v, key, nil, name},
      }, nil
    
    case "double":
      v, err := cast.ToFloat64E(val)
      if err != nil {
        continue Loop
      }
      return []*Binding{
        {clb, t, v, key, nil, name},
      }, nil
    
    case "string":
      v, err := cast.ToStringE(val)
      if err != nil {
        continue Loop
      }
      
      return []*Binding{
        {clb, t, v, key, nil, name},
      }, nil
    
    case "File":
      var v cwl.Entry
      def, ok := val.(*cwl.InputDefault)
      if ok {
        v = *def.Entry
      } else {
        v, ok = val.(cwl.Entry)
      }
      if !ok {
        continue Loop
      }
      
      f, err := process.resolveFile(v, getLoadContents(clb))
      if err != nil {
        return nil, err
      }
      // use process.runtime.RootHost as /
      // TODO 这种命名方式可能导致同名文件冲突问题
      f.Path = filepath.Join(process.runtime.RootHost, "/inputs/", f.Path)
      
      //f.Path = "/inputs/" + f.Path
      for _, expr := range secondaryFiles {
        process.resolveSecondaryFiles(f, expr)
      }
      
      return []*Binding{
        {clb, t, f, key, nil, name},
      }, nil
    
    case "Directory":
      v, ok := val.(cwl.Directory)
      if !ok {
        continue Loop
      }
      // TODO resolve directory
      return []*Binding{
        {clb, t, v, key, nil, name},
      }, nil
    default:
    
    //case cwl.TypeRef:
      if rsd, ok := process.RequiresSchemaDef(); ok {
        for _, rsdT := range rsd.Types {
          if rsdT.Name == t.Name || ( rsdT.Name == t.Name[1:] && t.Name[1] == '#')  {
            //do something
            // how to parse Value -> rsdT.Type ?
            st := rsdT.Type
            return []*Binding{
              {clb, t, st, key, nil, name},
            }, nil
          }
        }
      }
      continue Loop
    }
  }
  
  return nil, process.error("missing value")
}

func getLoadContents(clb *cwl.Binding) bool {
  if clb == nil {
    return false
  }
  return clb.LoadContents
}

func (process *Process) MigrateInputs() (err error){
  files := []cwl.Entry{}
  for _, in := range process.bindings {
    if f, ok := in.Value.(cwl.Entry); ok {
      files = append(files, FlattenFiles(f)...)
    }
  }
  // Do migrate
  fs := process.fs
  //if err = fs.EnsureDir(process.runtime.RootHost, 0750); err != nil {
  //  return err
  //}
  if err = fs.EnsureDir(process.runtime.RootHost + "/inputs", 0750); err != nil {
    return err
  }
  for _, filei := range files {
    if filei.Contents != "" && filei.Location == "" && filei.Path != ""{
      if _, err = fs.Create(filei.Path, filei.Contents); err != nil {
        return err
      }
    }
    if err = fs.Migrate(filei.Location, filei.Path); err != nil {
      return err
    }
  }
  return nil
}


func FlattenFiles(f cwl.Entry) []cwl.Entry{
  return []cwl.Entry{ f }
}