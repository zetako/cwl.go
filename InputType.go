package cwl

// 4.1
// primitive or record or enum
// or array of them
type InputType struct {
  name      string // null / boolean / int / long / float / double / string / `definedType`
  primitive string
  record    *InputRecordSchema
  enum      *InputEnumSchema
  array      *InputArraySchema
  multi     []InputType
}

// 4.1.3
type InputRecordSchema struct {
  // sld:
  Type   string
  Fields []InputRecordField
  // From InputRecordSchema
  InputSchema `json:",inline"`
}

// 4.1.4
type InputRecordField struct {
  // sld:RecordField
  Name string
  Documented `json:",inline"`
  Type InputType
  // From InputRecordField
  FieldBase `json:",inline"`
  InputFormat `json:",inline"`
  LoadContents `json:",inline"`
}

// 4.1.4.1
type InputEnumSchema struct {
  // sld:EnumSchema
  Type string // must be enum
  Symbols []string
  // From InputEnumSchema
  InputSchema `json:",inline"`
}

// 4.1.4.2
type InputArraySchema struct {
  // sld:ArraySchema
  Type  string    `json:"type"` // must be array
  Items InputType `json:"items"`
  // From InputArraySchema
  InputSchema `json:",inline"`
}

type OutputType struct {
  name      string // null / boolean / int / long / float / double / string / `definedType`
  primitive string
  record    *OutputRecordSchema
  enum      *OutputEnumSchema
  array      *OutputArraySchema
  multi     []OutputType
}


// 4.1.3
type OutputRecordSchema struct {
  // sld:
  Type   string
  Fields []OutputRecordField
  // From InputRecordSchema
  OutputSchema `json:",inline"`
}

// 4.1.4
type OutputRecordField struct {
  // sld:RecordField
  Name string
  Documented `json:",inline"`
  Type OutputType
  // From InputRecordField
  FieldBase `json:",inline"`
  OutputFormat `json:",inline"`
}

// 4.1.4.1
type OutputEnumSchema struct {
  // sld:EnumSchema
  Type string // must be enum
  Symbols []string
  // From InputEnumSchema
  OutputSchema `json:",inline"`
}

// 4.1.4.2
type OutputArraySchema struct {
  // sld:ArraySchema
  Type  string    `json:"type"` // must be array
  Items OutputType `json:"items"`
  // From InputArraySchema
  OutputSchema `json:",inline"`
}

type CommandInputRecordField struct {
  // sld:RecordField
  Name string
  Documented `json:",inline"`
  Type CommandInputType
  // From InputRecordField
  FieldBase `json:",inline"`
  InputFormat `json:",inline"`
  LoadContents `json:",inline"`
  CommandLineBindable `json:",inline"`
}


type CommandInputRecordSchema struct {
 CommandInputSchemaBase // abstract
 CommandLineBindable `json:",inline"`
  
  Type   string
  Fields []CommandInputRecordField
  // From InputRecordSchema
  InputSchema `json:",inline"`
}

type CommandInputEnumSchema struct {
 CommandInputSchemaBase // abstract
 CommandLineBindable `json:",inline"`
  
  // sld:EnumSchema
  InputEnumSchema `json:",inline"`
}


type CommandInputArraySchema struct {
  CommandInputSchemaBase // abstract
  CommandLineBindable `json:",inline"`
  
  // sld:EnumSchema
  Type  string    `json:"type"` // must be array
  Items CommandInputType `json:"items"`
  // From InputArraySchema
  InputSchema `json:",inline"`
}

//type CommandOutputType struct {
//
//}

type CommandOutputRecordField struct {
  Name string
  Documented `json:",inline"`
  Type CommandOutputType
  // From InputRecordField
  FieldBase `json:",inline"`
  OutputFormat `json:",inline"`
}

type CommandOutputRecordSchema struct {
  // sld:
  Type   string
  Fields []CommandOutputRecordField
  // From InputRecordSchema
  OutputSchema `json:",inline"`
}
