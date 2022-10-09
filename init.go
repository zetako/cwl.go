package cwl

func init() {
	classMap["InlineJavascriptRequirement"] = InlineJavascriptRequirement{}
	classMap["SchemaDefRequirement"] = SchemaDefRequirement{}
	classMap["LoadListingRequirement"] = LoadListingRequirement{}

	classMap["DockerRequirement"] = DockerRequirement{}
	classMap["ResourceRequirement"] = ResourceRequirement{}

	classMap["SubworkflowFeatureRequirement"] = SubworkflowFeatureRequirement{}
	classMap["ScatterFeatureRequirement"] = ScatterFeatureRequirement{}
	classMap["MultipleInputFeatureRequirement"] = MultipleInputFeatureRequirement{}
	classMap["StepInputExpressionRequirement"] = StepInputExpressionRequirement{}

}
