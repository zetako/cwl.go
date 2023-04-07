package cwl

// Just For test
func (p *ProcessBase) UnmarshalJSON(data []byte) error {
	graph := &RecordFieldGraph{Example: ProcessBase{}, Fields: make(map[string]*RecordFieldGraph)}
	db := graph.Fields
	db["InputParameter"] = &RecordFieldGraph{Example: InputParameterBase{}}
	db["OutputParameter"] = &RecordFieldGraph{Example: OutputParameterBase{}}
	parser := NewParser(graph, classMap)
	return parser.Unmarshal(data, p)
}

func (p *CommandLineTool) UnmarshalJSON(data []byte) error {
	graph := &RecordFieldGraph{Example: CommandLineTool{}, Fields: make(map[string]*RecordFieldGraph)}
	db := graph.Fields

	db["InputParameter"] = &RecordFieldGraph{Example: CommandInputParameter{}}
	db["OutputParameter"] = &RecordFieldGraph{Example: CommandOutputParameter{}}
	inputFields := map[string]*RecordFieldGraph{
		"ArrayType":  &RecordFieldGraph{Example: CommandInputArraySchema{}},
		"EnumType":   &RecordFieldGraph{Example: CommandInputEnumSchema{}},
		"RecordType": &RecordFieldGraph{Example: CommandInputRecordSchema{}},
		"FieldType":  &RecordFieldGraph{Example: CommandInputRecordField{}},
	}
	//db["SchemaDefRequirement"] = &RecordFieldGraph{Example: SchemaDefRequirement{}, Fields: inputFields}

	db["CommandInputSchema"] = &RecordFieldGraph{Example: CommandInputType{}, Fields: inputFields}
	db["InputBinding"] = &RecordFieldGraph{Example: CommandLineBinding{}}
	//CommandInputType
	db["CommandInputType"] = &RecordFieldGraph{Example: CommandInputType{},
		Fields: inputFields,
	}
	db["CommandOutputType"] = &RecordFieldGraph{Example: CommandOutputType{},
		Fields: map[string]*RecordFieldGraph{
			"ArrayType":  &RecordFieldGraph{Example: CommandOutputArraySchema{}},
			"EnumType":   &RecordFieldGraph{Example: CommandOutputEnumSchema{}},
			"RecordType": &RecordFieldGraph{Example: CommandOutputRecordSchema{}},
			"FieldType":  &RecordFieldGraph{Example: CommandOutputRecordField{}},
		},
	}

	parser := NewParser(graph, classMap)
	return parser.Unmarshal(data, p)
}

func (p *ExpressionTool) UnmarshalJSON(data []byte) error {

	graph := &RecordFieldGraph{Example: ExpressionTool{}, Fields: make(map[string]*RecordFieldGraph)}
	db := graph.Fields

	db["InputParameter"] = &RecordFieldGraph{Example: WorkflowInputParameter{}}
	db["OutputParameter"] = &RecordFieldGraph{Example: ExpressionToolOutputParameter{}}

	db["SaladType"] = &RecordFieldGraph{Example: CommandInputType{},
		Fields: map[string]*RecordFieldGraph{
			"ArrayType":  &RecordFieldGraph{Example: ArraySchema{}},
			"EnumType":   &RecordFieldGraph{Example: EnumSchema{}},
			"RecordType": &RecordFieldGraph{Example: RecordSchema{}},
		},
	}

	parser := NewParser(graph, classMap)
	return parser.Unmarshal(data, p)

}

func (p *Workflow) UnmarshalJSON(data []byte) error {

	graph := &RecordFieldGraph{Example: Workflow{}, Fields: make(map[string]*RecordFieldGraph)}
	db := graph.Fields

	db["InputParameter"] = &RecordFieldGraph{Example: WorkflowInputParameter{}}
	db["OutputParameter"] = &RecordFieldGraph{Example: WorkflowOutputParameter{}}

	inputFields := map[string]*RecordFieldGraph{
		"ArrayType":  &RecordFieldGraph{Example: ArraySchema{}},
		"EnumType":   &RecordFieldGraph{Example: EnumSchema{}},
		"RecordType": &RecordFieldGraph{Example: RecordSchema{}},
		"FieldType":  &RecordFieldGraph{Example: RecordField{}},
	}
	db["SaladType"] = &RecordFieldGraph{Example: CommandInputType{},
		Fields: inputFields,
	}

	db["CommandInputSchema"] = &RecordFieldGraph{Example: CommandInputType{}, Fields: inputFields}

	parser := NewParser(graph, classMap)
	return parser.Unmarshal(data, p)

}

func (p *Operation) UnmarshalJSON(data []byte) error {

	graph := &RecordFieldGraph{Example: Operation{}, Fields: make(map[string]*RecordFieldGraph)}
	db := graph.Fields

	db["InputParameter"] = &RecordFieldGraph{Example: OperationInputParameter{}}
	db["OutputParameter"] = &RecordFieldGraph{Example: OperationOutputParameter{}}

	db["SaladType"] = &RecordFieldGraph{Example: CommandInputType{},
		Fields: map[string]*RecordFieldGraph{
			"ArrayType":  &RecordFieldGraph{Example: ArraySchema{}},
			"EnumType":   &RecordFieldGraph{Example: EnumSchema{}},
			"RecordType": &RecordFieldGraph{Example: RecordSchema{}},
		},
	}

	parser := NewParser(graph, classMap)
	return parser.Unmarshal(data, p)
}

func JsonUnmarshal(data []byte, bean interface{}, graphs ...RecordFieldGraph) error {
	graph := &RecordFieldGraph{Example: Operation{}, Fields: make(map[string]*RecordFieldGraph)}
	db := graph.Fields

	for i, gi := range graphs {
		db[gi.ID] = &graphs[i]
	}
	parser := NewParser(graph, classMap)
	return parser.Unmarshal(data, bean)
}

func (recv *Values) UnmarshalJSON(b []byte) error {
	//return nil
	//
	//if recv == nil {
	//	recv = NewValues()
	//}
	//var any interface{}
	//if err := json.Unmarshal(b, &any); err != nil {
	//	return err
	//}
	//params, ok := any.(map[string]interface{})
	//if !ok {
	//	return fmt.Errorf("not a key-value type")
	//}
	//for key, value := range params {
	//	v, err := ConvertToValue(value)
	//	if err != nil {
	//		return err
	//	}
	//	(*recv)[key] = v
	//}
	parser := NewParser(nil, classMap)
	parser.salad.IsValue = true
	return parser.Unmarshal(b, recv)
	//return nil
}
