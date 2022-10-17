package cwl

type OperationInputParameter struct {
	InputParameterBase `json:"inline"`
	Type               SaladType `json:"type"`
}

type OperationOutputParameter struct {
	OutputParameterBase `json:"inline"`
	Type                SaladType `json:"type"`
}

type Operation struct {
	ProcessBase `json:"inline" salad:"abstract"`
	ClassBase   `json:"inline"`
}
