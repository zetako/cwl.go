package irunner

import "github.com/otiai10/cwl.go"

func (process *Process) RequiresSchemaDef() (*cwl.Requirement, bool) {
  for _, ri := range process.tool.Requirements {
    if ri.Class == "SchemaDefRequirement" {
      return &ri, true
    }
  }
  return nil, false
}
