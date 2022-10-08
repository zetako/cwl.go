package cwl

const version  = "v1.0"

// CWLType
// sld:PrimitiveType extends
// cwl:File cwl:Directory

type FileDir interface {
  filedir()
  Classable
}

// File represents file entry.
// @see http://www.commonwl.org/v1.0/CommandLineTool.html#File
type File struct {
  ClassBase      `json:",inline"`
  Location       string    `json:"location,omitempty"`
  Path           string    `json:"path,omitempty"`
  Basename       string    `json:"basename,omitempty"`
  Dirname        string    `json:"dirname,omitempty"`
  Nameroot       string    `json:"nameroot,omitempty"`
  Nameext        string    `json:"nameext,omitempty"`
  Checksum       string    `json:"checksum,omitempty"`
  Size           int64     `json:"size"`
  Format         string    `json:"format,omitempty"`
  Contents       string    `json:"contents,omitempty"`  // the file must be a UTF-8 text file 64 KiB or smaller
  SecondaryFiles []FileDir `json:"secondaryFiles,omitempty"`
}

// Directory represents direcotry entry.
// @see http://www.commonwl.org/v1.0/CommandLineTool.html#Directory
type Directory struct {
  ClassBase `json:",inline"`
  Location  string    `json:"location,omitempty"`
  Path      string    `json:"path,omitempty"`
  Basename  string    `json:"basename,omitempty"`
  Listing   []FileDir `json:"listing,omitempty"`
  Values
}



func (File) filedir()      {}
func (Directory) filedir() {}

// Labeled
// alias SchemaBase in v1.0
//       Labeled in v1.2
type Labeled struct {
  Label string `json:"label,omitempty"`
}

type Identified struct {
  ID string `json:"id,omitempty"`
}

type Documented struct {
  Doc ArrayString  `json:"doc,omitempty"` // comment
}

type FieldBase struct {
  // secondaryFiles v1.0 采用此种表示方法
  //SecondaryFiles ArrayExpression `json:"secondaryFiles,omitempty"`
  // secondaryFiles v1.2 采用此种表示方法
  SecondaryFiles []SecondaryFileSchema `json:"secondaryFiles,omitempty"`
  Streamable bool  `json:"streamable,omitempty"` // default False
}


type Parameter struct {
  FieldBase  `json:",inline"`
  Identified `json:",inline"`
  Labeled    `json:",inline"`
  Documented `json:",inline"`
}


type LoadListingEnum string

const (
  NO_LISTING LoadListingEnum = "no_listing"
  SHALLOW_LISTING LoadListingEnum = "shallow_listing"
  DEEP_LISTING LoadListingEnum = "deep_listing"
)

// LoadContents
// abstract: true
// v1.0 is InputBinding
type LoadContents struct {
  LoadContents bool  `json:"loadContents,omitempty"`
  LoadListing LoadListingEnum `json:"loadListing,omitempty"` // By default: `no_listing`
}

type InputFormat struct {
  format ArrayString  `json:"format,omitempty"`
}

type OutputFormat struct {
  format Expression  `json:"format,omitempty"`
}

// InputParameterBase
// abstract: true
type InputParameterBase struct {
  Parameter `json:",inline"`
  InputFormat `json:",inline"`
  LoadContents `json:",inline"`
  Default interface{} `json:"default,omitempty"`
}

type InputParameter interface {
  GetInputParameter() InputParameterBase
}

func (i InputParameterBase) GetInputParameter() InputParameterBase {
  return i
}

type OutputParameterBase struct {
  Parameter `json:",inline"`
  InputFormat `json:",inline"`
}

type OutputParameter interface {
  GetOutputParameter() OutputParameterBase
}

func (self OutputParameterBase) GetOutputParameter() OutputParameterBase {
  return self
}


type Requirement interface {
  requirement()
  Classable
}

type BaseRequirement struct {
  ClassBase `json:",inline"`
}

func (_ BaseRequirement)  requirement() {}

// Requirements represents "requirements" field in CWL.
type Requirements []Requirement


type Process interface {
  process()
}

// ProcessBase
// abstract
type ProcessBase struct {
  CWLVersion string `json:"cwlVersion,omitempty"`
  Intent []string `json:"intent,omitempty"`
  Identified `json:",inline"`
  Labeled    `json:",inline"`
  Documented `json:",inline"`
  Requirements `json:"requirements,omitempty" salad:"mapSubject:class"`
  Hits []interface{}        `json:"hits,omitempty" salad:"mapSubject:class"`
  Inputs []InputParameter   `json:"inputs,omitempty" salad:"mapSubject:id,mapPredicate:type"`
  Outputs []OutputParameter `json:"outputs,omitempty" salad:"mapSubject:id,mapPredicate:type"`
}

func (_ *ProcessBase) process() {

}

// InlineJavascriptRequirement is supposed to be embeded to Requirement.
// @see http://www.commonwl.org/v1.0/CommandLineTool.html#InlineJavascriptRequirement
type InlineJavascriptRequirement struct {
  BaseRequirement `json:",inline"`
  ExpressionLib []string `json:"expressionLib,omitempty"`
}

type CommandInputSchema interface {
  SchemaTypename() string
}

type  CommandInputSchemaBase struct {
}

func (_ CommandInputSchemaBase)  SchemaTypename() {}


// SchemaDefRequirement is supposed to be embeded to Requirement.
// @see http://www.commonwl.org/v1.0/CommandLineTool.html#SchemaDefRequirement
type SchemaDefRequirement struct {
  BaseRequirement `json:",inline"`
  Types []CommandInputSchema `json:"types,omitempty"`
}

type SecondaryFileSchema struct {
  Pattern string `json:"pattern,omitempty"`
  // null ? bool? Expression
  Required string `json:"required,omitempty"` // Default true for input; false for output
  RequiredVal *bool
}

type LoadListingRequirement struct {
  BaseRequirement `json:",inline"`
  LoadListing LoadListingEnum `json:"loadListing,omitempty"`
}

