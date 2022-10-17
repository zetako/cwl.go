package cwl

type Inputs []InputParameter


// Len for sorting.
func (ins Inputs) Len() int {
	return len(ins)
}

// Less for sorting.
func (ins Inputs) Less(i, j int) bool {
	prev, next := ins[i].GetInputParameter().ID , ins[j].GetInputParameter().ID
	return prev <= next
}

// Swap for sorting.
func (ins Inputs) Swap(i, j int) {
	ins[i], ins[j] = ins[j], ins[i]
}

// abstract types

type IOSchema struct {
	Name string `json:"name"`
	Labeled `json:",inline"`
	Documented `json:",inline"`
}

type InputSchema struct {
	IOSchema `json:",inline"`
}

type OutputSchema struct {
	IOSchema `json:",inline"`
}

//type InputRecordField struct {
//	RecordField `json:",inline"`
//	FieldBase `json:",inline"`
//	InputFormat `json:",inline"`
//	LoadContents `json:",inline"`
//}
