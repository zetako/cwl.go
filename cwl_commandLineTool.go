package cwl

// Doc: cwltool/schemas/v1.2/CommandLineTool.yml

const (
	STDIN  = "stdin"
	STDOUT = "stdout"
	STDERR = "stderr"
)

type EnvironmentDef struct {
	EnvName  string     `json:"envName"`
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
	Position      *IntExpression   `json:"position,omitempty"`
	Prefix        string           `json:"prefix"`
	Separate      bool             `json:"separate" salad:"default:true"`
	ItemSeparator string           `json:"itemSeparator,omitempty"`
	ValueFrom     Expression       `json:"valueFrom,omitempty"`
	ShellQuote    bool             `json:"shellQuote" salad:"default:true"`
}

type CommandOutputBinding struct {
	LoadContents `json:",inline"`
	Glob         ArrayExpression `json:"glob,omitempty"`
	OutputEval   Expression      `json:"outputEval,omitempty"`
}

type CommandLineBindable struct {
	InputBinding *CommandLineBinding `json:"inputBinding,omitempty"` // removed
}




type CommandInputParameter struct {
	InputParameterBase `json:",inline"`
	Type               CommandInputType    `json:"type" salad:"type"`
	//Type               CommandInputType    `json:"type"`
	InputBinding       *CommandLineBinding `json:"inputBinding,omitempty"`
}

type CommandOutputParameter struct {
	OutputParameterBase `json:",inline"`
	Type                CommandOutputType     `json:"type" salad:"type"`
	OutputBinding       *CommandOutputBinding `json:"outputBinding,omitempty"`
}

type CommandLineTool struct {
	ClassBase          `json:",inline"`
	ProcessBase        `json:",inline" salad:"abstract"`
	BaseCommands       ArrayString `json:"baseCommand,omitempty"`
	Arguments          Arguments   `json:"arguments,omitempty"`
	Stdin              Expression  `json:"stdin,omitempty"`
	Stderr             Expression  `json:"stderr,omitempty"`
	Stdout             Expression  `json:"stdout,omitempty"`
	SuccessCodes       []int       `json:"successCodes,omitempty"`
	TemporaryFailCodes []int       `json:"temporaryFailCodes,omitempty"`
	PermanentFailCodes []int       `json:"permanentFailCodes,omitempty"`
}

type DockerRequirement struct {
	BaseRequirement       `json:",inline"`
	DockerPull            string `json:"dockerPull,omitempty"`
	DockerLoad            string `json:"dockerLoad,omitempty"`
	DockerFile            string `json:"dockerFile,omitempty"`
	DockerImport          string `json:"dockerImport,omitempty"`
	DockerImageId         string `json:"dockerImageId,omitempty"`
	DockerOutputDirectory string `json:"dockerOutputDirectory,omitempty"`
}

type SoftwarePackage struct {
	Package string   `json:"package"`
	Version []string `json:"version,omitempty"`
	Specs   []string `json:"specs,omitempty"`
}

type SoftwareRequirement struct {
	BaseRequirement `json:",inline"`
	Packages        []SoftwarePackage `json:"packages" salad:"mapSubject:package,mapPredicate:specs"`
}

// Dirent represents ?
// @see http://www.commonwl.org/v1.0/CommandLineTool.html#Dirent
type Dirent struct {
	Entry     Expression `json:"entry,omitempty"`
	EntryName Expression `json:"entryname,omitempty"`
	Writable  bool       `json:"writable,omitempty"`
}

// ToSet different
type ListingCollect struct {
	dirents []Dirent
}

type InitialWorkDirRequirement struct {
	BaseRequirement `json:",inline"`
	Listing         ListingCollect `json:"listing"`
}

type EnvVarRequirement struct {
	BaseRequirement `json:",inline"`
	EnvDef          []EnvironmentDef `json:"envDef"`
}

type ShellCommandRequirement struct {
	BaseRequirement `json:",inline"`
}

type ResourceRequirement struct {
	BaseRequirement `json:",inline"`
	CoresMin        LongFloatExpression `json:"coresMin,omitempty"`
	CoresMax        LongFloatExpression `json:"coresMax,omitempty"`
	RamMin          LongFloatExpression `json:"ramMin,omitempty"` // Minimum reserved RAM in mebibytes (2**20) (default is 256)
	RamMax          LongFloatExpression `json:"ramMax,omitempty"`
	TmpdirMin       LongFloatExpression `json:"tmpdirMin,omitempty"` // Minimum reserved filesystem based storage for the designated temporary directory, in mebibytes (2**20) (default is 1024)
	TmpdirMax       LongFloatExpression `json:"tmpdirMax,omitempty"`
	OutdirMin       LongFloatExpression `json:"outdirMin,omitempty"` // Minimum reserved filesystem based storage for the designated output directory, in mebibytes (2**20) (default is 1024)
	OutdirMax       LongFloatExpression `json:"outdirMax,omitempty"`
}

type WorkReuse struct {
	BaseRequirement `json:",inline"`
	EnableReuse     BoolExpression `json:"enableReuse,omitempty"`
}

type NetworkAccess struct {
	BaseRequirement `json:",inline"`
	NetworkAccess   BoolExpression `json:"networkAccess,omitempty"`
}

type InplaceUpdateRequirement struct {
	BaseRequirement `json:",inline"`
	InplaceUpdate   bool `json:"inplaceUpdate"`
}

type ToolTimeLimit struct {
	BaseRequirement `json:",inline"`
	Timelimit       LongFloatExpression `json:"timelimit"`
}

// Argument represents an element of "arguments" of CWL
// @see http://www.commonwl.org/v1.0/CommandLineTool.html#CommandLineTool
type Argument struct {
	Exp     Expression
	Binding *CommandLineBinding
}

// Arguments represents a list of "Argument"
type Arguments []Argument

