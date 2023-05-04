package cwl

func init() {
	classMap["InlineJavascriptRequirement"] = InlineJavascriptRequirement{}
	classMap["SchemaDefRequirement"] = SchemaDefRequirement{}
	classMap["LoadListingRequirement"] = LoadListingRequirement{}

	classMap["DockerRequirement"] = DockerRequirement{}
	classMap["ResourceRequirement"] = ResourceRequirement{}
	classMap["InitialWorkDirRequirement"] = InitialWorkDirRequirement{}

	classMap["SubworkflowFeatureRequirement"] = SubworkflowFeatureRequirement{}
	classMap["ScatterFeatureRequirement"] = ScatterFeatureRequirement{}
	classMap["MultipleInputFeatureRequirement"] = MultipleInputFeatureRequirement{}
	classMap["StepInputExpressionRequirement"] = StepInputExpressionRequirement{}
	classMap["ShellCommandRequirement"] = ShellCommandRequirement{}
	classMap["EnvVarRequirement"] = EnvVarRequirement{}

	classMap["LoopFeatureRequirement"] = LoopFeatureRequirement{}

	classMap["File"] = File{}
	classMap["Directory"] = Directory{}

}

var (
	NullType = SaladType{primitive: "null"}
	//ArgType = SaladType{ primitive: 	"arg"}
)
