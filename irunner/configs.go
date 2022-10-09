package irunner

const (
	// mariner components
	marinerTask   = "task"
	marinerEngine = "engine"
	s3sidecar     = "s3sidecar"
	gen3fuse      = "gen3fuse"

	// default task docker image
	// this should be in the external config
	// not in the codebase
	defaultTaskContainerImage = "ubuntu"

	// volume names
	engineWorkspaceVolumeName = "engine-workspace"
	commonsDataVolumeName     = "commons-data"
	configVolumeName          = "mariner-config"
	conformanceVolumeName     = "conformance-test"

	// location of conformance test input files in s3
	conformanceInputS3Prefix = "conformanceTest/"

	// container name
	taskContainerName = "mariner-task"

	// file path prefixes - used to differentiate COMMONS vs USER vs marinerEngine WORKSPACE file
	// user specifies commons datafile by "COMMONS/<GUID>"
	// user specifies user datafile by "USER/<path>"
	commonsPrefix     = "COMMONS/"
	userPrefix        = "USER/"
	conformancePrefix = "CONFORMANCE/"
	workspacePrefix   = "/" + engineWorkspaceVolumeName
	gatewayPrefix     = "/" + commonsDataVolumeName

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
	logFile = "marinerLog.json"

	// done flag - used by engine
	doneFlag = "done"

	// workflow request file name
	requestFile = "request.json"

	// HTTP
	authHeader = "Authorization"

	// metrics collection sampling period (in seconds)
	metricsSamplingPeriod = 30

	// paths for engine
	//pathToCommonsData = "/commons-data/data/by-guid/"
	pathToCommonsData = "/commons-data/"
	pathToRunf        = "/engine-workspace/workflowRuns/%v/" // fill with runID
	pathToLogf        = pathToRunf + logFile
	pathToDonef       = pathToRunf + doneFlag
	pathToRequestf    = pathToRunf + requestFile
	pathToWorkingDirf = pathToRunf + "%v" // fill with runID

	// paths for server
	pathToUserRunsf   = "%v/workflowRuns/"                // fill with userID
	pathToUserRunLogf = pathToUserRunsf + "%v/" + logFile // fill with runID

	commonsDataPersistentVolumeClaimName = "mariner-nfs-pvc"
)
