package cwl

import (
	"encoding/json"
	"fmt"
)

const version = "v1.0"

// CWLType
// sld:PrimitiveType extends
// cwl:File cwl:Directory

type FileDirI interface {
	filedir()
	Classable
}

type FileDir struct {
	ClassBase `json:",inline"`
	entry     FileDirI
}

func (e *FileDir) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &e.ClassBase)
	if err != nil {
		return err
	}
	if e.Class == "File" {
		entery := &File{}
		e.entry = entery
		return json.Unmarshal(b, e.entry)
	} else if e.Class == "Directory" {
		entery := &Directory{}
		e.entry = entery
		return json.Unmarshal(b, e.entry)
	}
	return fmt.Errorf("class need to be File/Directory")
}

func (e *FileDir) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Entery())
}

func NewFileDir(entry FileDirI) FileDir {
	return FileDir{
		ClassBase{entry.ClassName()},
		entry,
	}
}

func (e *FileDir) Entery() FileDirI {
	return e.entry
}

func (e *FileDir) Value() (*File, *Directory, error) {
	filei, isFile := e.Entery().(*File)
	if !isFile {
		var bean File
		if bean, isFile = e.Entery().(File); isFile {
			filei = &bean
		}
	}
	if isFile {
		return filei, nil, nil
	}
	diri, isDir := e.Entery().(*Directory)
	if !isDir {
		var bean Directory
		if bean, isDir = e.Entery().(Directory); isDir {
			diri = &bean
		}
	}
	if isDir {
		return nil, diri, nil
	}
	return nil, nil, fmt.Errorf("Bad FileDir Entry")
}

// File represents file entry.
// https://www.commonwl.org/v1.2/CommandLineTool.html#File
type File struct {
	ClassBase      `json:",inline"`
	Location       string    `json:"location,omitempty"` // file:// http://
	Path           string    `json:"path,omitempty"`     // runtime local path
	Basename       string    `json:"basename,omitempty"`
	Dirname        string    `json:"dirname,omitempty"`
	Nameroot       string    `json:"nameroot,omitempty"`
	Nameext        string    `json:"nameext,omitempty"`
	Checksum       string    `json:"checksum,omitempty"`
	Size           int64     `json:"size"`
	Format         string    `json:"format,omitempty"`
	Contents       string    `json:"contents,omitempty"` // the file must be a UTF-8 text file 64 KiB or smaller
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
	//Listing   EntryListing `json:"listing,omitempty"`
	//Listing   EntryListing `json:"listing,omitempty"`
}

func (File) filedir()      {}
func (Directory) filedir() {}

type EntryListing []FileDir

//func (l *EntryListing) UnmarshalJSON(b []byte) error {
//	return setField( reflect.TypeOf(l), reflect.ValueOf(l), b, saladTags{}, map[string]*RecordFieldGraph{
//		"File" : &RecordFieldGraph{Example: File{}},
//		"Directory" : &RecordFieldGraph{Example: Directory{}},
//	} )
//}

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
	Doc ArrayString `json:"doc,omitempty"` // comment
}

type FieldBase struct {
	// secondaryFiles v1.0 采用此种表示方法
	//SecondaryFiles ArrayExpression `json:"secondaryFiles,omitempty"`
	// secondaryFiles v1.2 采用此种表示方法
	SecondaryFiles []SecondaryFileSchema `json:"secondaryFiles,omitempty"`
	Streamable     bool                  `json:"streamable,omitempty"` // default False
}

type Parameter struct {
	FieldBase  `json:",inline"`
	Identified `json:",inline"`
	Labeled    `json:",inline"`
	Documented `json:",inline"`
}

type LoadListingEnum string

const (
	NO_LISTING      LoadListingEnum = "no_listing"
	SHALLOW_LISTING LoadListingEnum = "shallow_listing"
	DEEP_LISTING    LoadListingEnum = "deep_listing"
)

// LoadContents
// abstract: true
// v1.0 is InputBinding
type LoadContents struct {
	LoadContents bool            `json:"loadContents,omitempty"`
	LoadListing  LoadListingEnum `json:"loadListing,omitempty"` // By default: `no_listing`
}

type InputFormat struct {
	format ArrayString `json:"format,omitempty"`
}

type OutputFormat struct {
	format Expression `json:"format,omitempty"`
}

// InputParameterBase
// abstract: true
type InputParameterBase struct {
	Parameter    `json:",inline"`
	InputFormat  `json:",inline"`
	LoadContents `json:",inline"`
	Default      Value `json:"default,omitempty" salad:"value"`
}

type InputParameter interface {
	GetInputParameter() InputParameterBase
}

func (i InputParameterBase) GetInputParameter() InputParameterBase {
	return i
}

type OutputParameterBase struct {
	Parameter   `json:",inline"`
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

func (_ BaseRequirement) requirement() {}

// Requirements represents "requirements" field in CWL.
type Requirements []Requirement

type Process interface {
	//process() ProcessBase
	Base() *ProcessBase
}

// ProcessBase
// abstract
type ProcessBase struct {
	CWLVersion   string   `json:"cwlVersion,omitempty"`
	Intent       []string `json:"intent,omitempty"`
	Identified   `json:",inline"`
	Labeled      `json:",inline"`
	Documented   `json:",inline"`
	Requirements Requirements `json:"requirements,omitempty" salad:"mapSubject:class"`
	Hints        Requirements `json:"hints,omitempty" salad:"mapSubject:class"`
	//Inputs       []InputParameter  `json:"inputs,omitempty" salad:"mapSubject:id,mapPredicate:type"`
	Inputs  Inputs            `json:"inputs,omitempty" salad:"mapSubject:id,mapPredicate:type"`
	Outputs []OutputParameter `json:"outputs,omitempty" salad:"mapSubject:id,mapPredicate:type"`
}

func (b *ProcessBase) Base() *ProcessBase {
	return b
}

// InlineJavascriptRequirement is supposed to be embeded to Requirement.
// @see http://www.commonwl.org/v1.0/CommandLineTool.html#InlineJavascriptRequirement
type InlineJavascriptRequirement struct {
	BaseRequirement `json:",inline"`
	ExpressionLib   []string `json:"expressionLib,omitempty"`
}

type CommandInputSchema interface {
	SchemaTypename() string
}

type CommandInputSchemaBase struct {
}

func (_ CommandInputSchemaBase) SchemaTypename() string {
	return ""
}

// SchemaDefRequirement is supposed to be embeded to Requirement.
// @see http://www.commonwl.org/v1.0/CommandLineTool.html#SchemaDefRequirement
type SchemaDefRequirement struct {
	BaseRequirement `json:",inline"`
	Types           []CommandInputSchema `json:"types,omitempty" salad:"type"`
}

type SecondaryFileSchema struct {
	Pattern string `json:"pattern,omitempty"`
	// null ? bool? Expression
	Required bool `json:"required,omitempty"` // Default true for input; false for output
	// RequiredVal *bool
}

type LoadListingRequirement struct {
	BaseRequirement `json:",inline"`
	LoadListing     LoadListingEnum `json:"loadListing,omitempty"`
}
