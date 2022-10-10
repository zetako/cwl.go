package cwl

type InputRecordField struct {
  RecordField `json:",inline"`
  FieldBase `json:",inline"`
  InputFormat `json:",inline"`
  LoadContents `json:",inline"`
}

type InputRecordSchema struct {
  RecordSchema   `json:",inline"`
  InputSchema `json:",inline"`
}

type InputEnumSchema struct {
  EnumSchema `json:",inline"`
  InputSchema `json:",inline"`
}

type InputArraySchema struct {
  ArraySchema `json:",inline"`
  InputSchema `json:",inline"`
}

type OutputRecordField struct {
  RecordField `json:",inline"`
  FieldBase `json:",inline"`
  OutputFormat `json:",inline"`
}

type OutputRecordSchema struct {
  RecordSchema   `json:",inline"`
  OutputSchema `json:",inline"`
}

type OutputEnumSchema struct {
  EnumSchema `json:",inline"`
  OutputSchema `json:",inline"`
}

type OutputArraySchema struct {
  ArraySchema `json:",inline"`
  OutputSchema `json:",inline"`
}



type CommandInputRecordField struct {
  InputRecordField `json:",inline"`
  CommandLineBindable `json:",inline"`
}


type CommandInputRecordSchema struct {
 CommandInputSchemaBase // abstract
 CommandLineBindable `json:",inline"`
 
 //RecordSchema `json:",inline"`
 //InputSchema `json:",inline"`
  InputRecordSchema `json:",inline"`
}

type CommandInputEnumSchema struct {
 CommandInputSchemaBase // abstract
 CommandLineBindable `json:",inline"`
  
  // sld:EnumSchema
  InputEnumSchema `json:",inline"`
  
  //EnumSchema
  //InputSchema
}


type CommandInputArraySchema struct {
  CommandInputSchemaBase // abstract
  CommandLineBindable `json:",inline"`
  
  InputArraySchema `json:",inline"`
}

type CommandOutputRecordField struct {
  OutputRecordField  `json:",inline"`
  OutputBinding *CommandOutputBinding  `json:"outputBinding"`
}

type CommandOutputRecordSchema struct {
  // sld:
  OutputRecordSchema `json:",inline"`
}

type CommandOutputEnumSchema struct {
  OutputEnumSchema `json:",inline"`
}

type CommandOutputArraySchema struct {
  OutputArraySchema `json:",inline"`
}


// CommandInputType
// a collect for CWLType,stdin, CommandInputRecordSchema, CommandInputEnumSchema, CommandInputArraySchema, string
// and array of them
type CommandInputType struct {
  SaladType      `salad:"type"`
}


type CommandOutputType struct {
  SaladType `salad:"type"`
}
