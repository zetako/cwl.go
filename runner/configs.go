package runner

const (
	// mariner components

	notStarted = "not-started" // 3
	running    = "running"     // 2
	failed     = "failed"      // 1
	completed  = "completed"   // 0
	unknown    = "unknown"
	success    = "success"
	cancelled  = "cancelled"

	k8sJobAPI     = "k8sJobAPI"
	k8sPodAPI     = "k8sPodAPI"
	k8sMetricsAPI = "k8sMetricsAPI"
	k8sCoreAPI    = "k8sCoreAPI"

	// top-level workflow ID
	mainProcessID = "#main"

	//
	argType = "cwltype#arg"

	// cwl things //
	// parameter type
	CWLNullType      = "null"
	CWLFileType      = "File"
	CWLDirectoryType = "Directory"
	// object class
	CWLWorkflow        = "Workflow"
	CWLCommandLineTool = "CommandLineTool"
	CWLExpressionTool  = "ExpressionTool"
	// requirements
	CWLInitialWorkDirRequirement = "InitialWorkDirRequirement"
	CWLResourceRequirement       = "ResourceRequirement"
	CWLDockerRequirement         = "DockerRequirement"
	CWLEnvVarRequirement         = "EnvVarRequirement"
	// add the rest ..

	// log levels
	infoLogLevel    = "INFO"
	warningLogLevel = "WARNING"
	errorLogLevel   = "ERROR"

	// log file name

	// HTTP
	authHeader = "Authorization"
)
