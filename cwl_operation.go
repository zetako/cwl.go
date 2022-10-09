package cwl

import "reflect"

type OperationInputParameter struct {
	InputParameterBase `json:"inline"`
	Type               SaladType `json:"type"`
}

type OperationOutputParameter struct {
	OutputParameterBase `json:"inline"`
	Type                SaladType `json:"type"`
}

type Operation struct {
	ProcessBase `json:"inline"`
	ClassBase   `json:"inline"`
}

func (p *Operation) UnmarshalJSON(data []byte) error {
	typeOfRecv := reflect.TypeOf(*p)
	valueOfRecv := reflect.ValueOf(p).Elem()
	db := make(map[string]*RecordFieldGraph)
	db["InputParameter"] = &RecordFieldGraph{Example: OperationInputParameter{}}
	db["OutputParameter"] = &RecordFieldGraph{Example: OperationOutputParameter{}}
	if err := parseObject(typeOfRecv, valueOfRecv, data, db); err != nil {
		return err
	}
	return nil
}
