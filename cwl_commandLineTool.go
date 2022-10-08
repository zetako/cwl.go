package cwl

import (
  "encoding/json"
  "reflect"
)

// Doc: cwltool/schemas/v1.2/CommandLineTool.yml

const (
  STDIN = "stdin"
  STDOUT = "stdout"
  STDERR = "stderr"
)

type EnvironmentDef struct {
  EnvName string `json:"envName"`
  EnvValue Expression `json:"envValue"`
}

// Use of `loadContents` in `InputBinding` is deprecated.
// Preserved for v1.0 backwards compatability.  Will be removed in
// CWL v2.0.  Use `InputParameter.loadContents` instead.
type InputBinding struct {
  LoadContents *bool `json:"loadContents,omitempty"`
}

type CommandLineBinding struct {
  InputBinding  `json:",inline"` // removed
  Position *IntExpression `json:"position,omitempty"`
  Prefix string `json:"prefix"`
  Separate bool `json:"separate" salad:"default:true"`
  ItemSeparator string `json:"itemSeparator,omitempty"`
  ValueFrom Expression `json:"valueFrom,omitempty"`
  ShellQuote bool `json:"shellQuote" salad:"default:true"`
}

type CommandOutputBinding struct {
  LoadContents `json:",inline"`
  Glob ArrayExpression `json:"glob,omitempty"`
  OutputEval Expression `json:"outputEval,omitempty"`
}

type CommandLineBindable struct {
  InputBinding *CommandLineBinding  `json:"inputBinding,omitempty"` // removed
}


//type CommandInputRecordField struct {
//  CommandInputSchemaBase // abstract
//  CommandLineBindable `json:",inline"`
//  //InputRecordField `json:",inline"` // TODO
//}

//type CommandInputEnumSchema struct {
//  CommandInputSchemaBase // abstract
//  CommandLineBindable `json:",inline"`
//  //InputEnumSchema `json:",inline"` // TODO
//}

//type CommandInputArraySchema struct {
//  CommandInputSchemaBase // abstract
//  CommandLineBindable `json:",inline"`
//  //InputArraySchema `json:",inline"` // TODO
//}

//type CommandOutputRecordField struct {
//  //OutputRecordField `json:",inline"` // TODO
//  OutputBinding *CommandOutputBinding `json:"outputBinding,omitempty"`
//}
//
//type CommandOutputRecordSchema struct {
//  //OutputRecordSchema `json:",inline"` // TODO
//}
//
//type CommandOutputEnumSchema struct {
//  //OutputEnumSchema `json:",inline"` // TODO
//}
//
//type CommandOutputArraySchema struct {
//  //OutputArraySchema `json:",inline"` // TODO
//}

// CommandInputType
// a collect for CWLType,stdin, CommandInputRecordSchema, CommandInputEnumSchema, CommandInputArraySchema, string
// and array of them
type CommandInputType struct {
  SaladType
  IsStdin bool // stdin is extends for saladTypes
  Binding *CommandLineBinding // TODO
}


func (recv *CommandInputType) UnmarshalJSON(data []byte) error {
  err := json.Unmarshal(data, &recv.SaladType)
  if err != nil {
    return nil
  }
  if recv.SaladType.name == "stdin" {
    recv.IsStdin = true
  }
  binding := &CommandLineBindable{}
  if err = json.Unmarshal(data, binding) ; err != nil {
    // simple type has no object format
   //return err
    return nil
  }
  if binding.InputBinding != nil {
   recv.Binding = binding.InputBinding
  }
  return nil
}

type CommandInputParameter struct {
  InputParameterBase `json:",inline"`
  Type  CommandInputType `json:"type"`
  InputBinding *CommandLineBinding `json:"inputBinding,omitempty"`
}

type CommandOutputType struct {
  SaladType
  isStdout bool // extends for saladTypes
  isStderr bool // extends for saladTypes
  Binding *CommandOutputBinding // TODO
  // TODO
}

type outputBindable struct {
  OutputBinding *CommandOutputBinding  `json:"outputBinding,omitempty"` // removed
}

func (recv *CommandOutputType) UnmarshalJSON(data []byte) error {
  err := json.Unmarshal(data, &recv.SaladType)
  if err != nil {
    return nil
  }
  if recv.SaladType.name == "stderr" {
    recv.isStderr = true
  }
  if recv.SaladType.name == "stdout" {
    recv.isStdout = true
  }
  binding := &outputBindable{}
  if err = json.Unmarshal(data, binding) ; err != nil {
    // simple type has no object format
    //return err
    return nil
  }
  if binding.OutputBinding != nil {
   recv.Binding = binding.OutputBinding
  }
  return nil
}

type CommandOutputParameter struct {
  OutputParameterBase `json:",inline"`
  Type  CommandOutputType `json:"type"`
  OutputBinding *CommandOutputBinding `json:"outputBinding,omitempty"`
}

type CommandLineTool struct {
  ClassBase          `json:",inline"`
  ProcessBase        `json:",inline"`
  BaseCommands       ArrayString `json:"baseCommand,omitempty"`
  Arguments          Arguments  `json:"arguments,omitempty"`
  Stdin              Expression `json:"stdin,omitempty"`
  Stderr             Expression `json:"stderr,omitempty"`
  Stdout             Expression `json:"stdout,omitempty"`
  SuccessCodes       []int `json:"successCodes,omitempty"`
  TemporaryFailCodes []int `json:"temporaryFailCodes,omitempty"`
  PermanentFailCodes []int `json:"permanentFailCodes,omitempty"`
}

type DockerRequirement struct {
  BaseRequirement `json:",inline"`
  DockerPull string  `json:"dockerPull,omitempty"`
  DockerLoad string  `json:"dockerLoad,omitempty"`
  DockerFile string  `json:"dockerFile,omitempty"`
  DockerImport string  `json:"dockerImport,omitempty"`
  DockerImageId string  `json:"dockerImageId,omitempty"`
  DockerOutputDirectory string  `json:"dockerOutputDirectory,omitempty"`
}

type SoftwarePackage struct {
  Package string `json:"package"`
  Version []string `json:"version,omitempty"`
  Specs []string `json:"specs,omitempty"`
}

type SoftwareRequirement struct {
  BaseRequirement `json:",inline"`
  Packages []SoftwarePackage `json:"packages" salad:"mapSubject:package,mapPredicate:specs"`
}

// Dirent represents ?
// @see http://www.commonwl.org/v1.0/CommandLineTool.html#Dirent
type Dirent struct {
  Entry     Expression `json:"entry,omitempty"`
  EntryName Expression `json:"entryname,omitempty"`
  Writable  bool `json:"writable,omitempty"`
}

// ToSet different
type ListingCollect struct {
  dirents []Dirent
}

type InitialWorkDirRequirement struct {
  BaseRequirement `json:",inline"`
  Listing ListingCollect `json:"listing"`
}

type EnvVarRequirement struct {
  BaseRequirement `json:",inline"`
  EnvDef []EnvironmentDef `json:"envDef"`
}

type ShellCommandRequirement struct {
  BaseRequirement `json:",inline"`
}

type ResourceRequirement struct {
  BaseRequirement `json:",inline"`
  CoresMin LongFloatExpression `json:"coresMin,omitempty"`
  CoresMax LongFloatExpression `json:"coresMax,omitempty"`
  RamMin LongFloatExpression `json:"ramMin,omitempty"` // Minimum reserved RAM in mebibytes (2**20) (default is 256)
  RamMax LongFloatExpression `json:"ramMax,omitempty"`
  TmpdirMin LongFloatExpression `json:"tmpdirMin,omitempty"` // Minimum reserved filesystem based storage for the designated temporary directory, in mebibytes (2**20) (default is 1024)
  TmpdirMax LongFloatExpression `json:"tmpdirMax,omitempty"`
  OutdirMin LongFloatExpression `json:"outdirMin,omitempty"` // Minimum reserved filesystem based storage for the designated output directory, in mebibytes (2**20) (default is 1024)
  OutdirMax LongFloatExpression `json:"outdirMax,omitempty"`
}

type WorkReuse struct {
  BaseRequirement `json:",inline"`
  EnableReuse BoolExpression `json:"enableReuse,omitempty"`
}

type NetworkAccess struct {
  BaseRequirement `json:",inline"`
  NetworkAccess BoolExpression `json:"networkAccess,omitempty"`
}

type InplaceUpdateRequirement struct {
  BaseRequirement `json:",inline"`
  InplaceUpdate bool `json:"inplaceUpdate"`
}

type ToolTimeLimit struct {
  BaseRequirement `json:",inline"`
  Timelimit LongFloatExpression `json:"timelimit"`
}

// Argument represents an element of "arguments" of CWL
// @see http://www.commonwl.org/v1.0/CommandLineTool.html#CommandLineTool
type Argument struct {
  exp Expression
  binding *CommandLineBinding
}

// Arguments represents a list of "Argument"
type Arguments []Argument

//
func (p *CommandLineTool)  UnmarshalJSON(data []byte) error{
  typeOfRecv := reflect.TypeOf(*p)
  valueOfRecv := reflect.ValueOf(p).Elem()
  db := make(map[string]*RecordFieldGraph)
  db["InputParameter"] = &RecordFieldGraph{ Example: CommandInputParameter{} }
  db["OutputParameter"] = &RecordFieldGraph{ Example: CommandOutputParameter{} }
  if err := parseObject(typeOfRecv, valueOfRecv, data,  db); err != nil {
    return err
  }
  return nil
}