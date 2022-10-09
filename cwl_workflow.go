package cwl

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type ExpressionToolOutputParameter struct {
	OutputParameterBase `json:",inline"`
	Type                SaladType `json:"type"`
}

type WorkflowInputParameter struct {
	InputParameterBase `json:",inline"`
	InputBinding       *InputBinding `json:"inputBinding,omitempty"`
	Type               SaladType     `json:"type"`
}

type ExpressionTool struct {
	ClassBase   `json:",inline"`
	ProcessBase `json:",inline"`
	Expression  Expression `json:"expression"`
}

// LinkMergeMethod enum
// The input link merge method, described in [WorkflowStepInput](#WorkflowStepInput).
type LinkMergeMethod string

const (
	MERGE_NESTED    LinkMergeMethod = "merge_nested"
	MERGE_FLATTENED LinkMergeMethod = "merge_flattened"
)

// PickValueMethod enum
//     Picking non-null values among inbound data links, described in [WorkflowStepInput](#WorkflowStepInput).
type PickValueMethod string

const (
	FIRST_NON_NULL    PickValueMethod = "first_non_null"
	THE_ONLY_NON_NULL PickValueMethod = "the_only_non_null"
	ALL_NON_NULL      PickValueMethod = "all_non_null"
)

type WorkflowOutputParameter struct {
	OutputParameterBase `json:",inline"`
	OutputSource        ArrayString      `json:"outputSource,omitempty"`
	LinkMerge           LinkMergeMethod  `json:"linkMerge,omitempty" salad:"default:merge_nested"`
	PickValue           *PickValueMethod `json:"pickValue,omitempty"`
	Type                SaladType        `json:"type"`
}

// Sink
// abstract struct
type Sink struct {
	Source    ArrayString      `json:"source,omitempty"`
	LinkMerge LinkMergeMethod  `json:"linkMerge,omitempty" salad:"default:merge_nested"`
	PickValue *PickValueMethod `json:"pickValue,omitempty"`
}

type WorkflowStepInput struct {
	Identified   `json:",inline"`
	Sink         `json:",inline"`
	LoadContents `json:",inline"`
	Labeled      `json:",inline"`
	Default      interface{} `json:"default,omitempty"`
	ValueFrom    Expression  `json:"valueFrom,omitempty"`
}

type WorkflowStepOutput struct {
	Identified `json:",inline"`
}

func (e *WorkflowStepOutput) UnmarshalJSON(data []byte) error {
	var bean interface{}
	err := json.Unmarshal(data, &bean)
	if err != nil {
		return err
	}
	switch v := bean.(type) {
	case string:
		e.ID = v
		return nil
	case map[string]interface{}:
		id, got := v["id"]
		if got {
			e.ID = id.(string)
			return nil
		}
	}
	return fmt.Errorf("WorkflowStepOutput Need string / {id: xxxx}")
}

type ScatterMethod string

const (
	DOTPRODUCT          ScatterMethod = "dotproduct"
	NESTED_CROSSPRODUCT ScatterMethod = "nested_crossproduct"
	FLAT_CROSSPRODUCT   ScatterMethod = "flat_crossproduct"
)

type Run struct {
	ID      string
	Process Process
}

func (e *Run) UnmarshalJSON(data []byte) error {
	var id string
	err := json.Unmarshal(data, &id)
	if err == nil {
		e.ID = id
		return nil
	}
	// TODO UnmarshalJSON Process
	return fmt.Errorf("TODO ")
}

type WorkflowStep struct {
	Identified    `json:",inline"`
	Labeled       `json:",inline"`
	Documented    `json:",inline"`
	In            []WorkflowStepInput  `json:"in" salad:"mapSubject:id,mapPredicate:source"`
	Out           []WorkflowStepOutput `json:"out"`
	Requirements  Requirements         `json:"requirements,omitempty" salad:"mapSubject:class"`
	Hits          Requirements         `json:"requirements,omitempty" salad:"mapSubject:class"`
	Run           Run                  `json:"run"`
	When          Expression           `json:"when,omitempty"`
	Scatter       ArrayString          `json:"scatter,omitempty"`
	ScatterMethod ScatterMethod        `json:"scatterMethod,omitempty"`
}

type Workflow struct {
	ProcessBase `json:",inline"`
	ClassBase   `json:",inline"`
	Steps       []WorkflowStep `json:"steps" salad:"mapSubject:id"`
}

type SubworkflowFeatureRequirement struct {
	BaseRequirement `json:",inline"`
}

type ScatterFeatureRequirement struct {
	BaseRequirement `json:",inline"`
}

type MultipleInputFeatureRequirement struct {
	BaseRequirement `json:",inline"`
}

type StepInputExpressionRequirement struct {
	BaseRequirement `json:",inline"`
}

func (p *ExpressionTool) UnmarshalJSON(data []byte) error {
	typeOfRecv := reflect.TypeOf(*p)
	valueOfRecv := reflect.ValueOf(p).Elem()
	db := make(map[string]*RecordFieldGraph)
	db["InputParameter"] = &RecordFieldGraph{Example: WorkflowInputParameter{}}
	db["OutputParameter"] = &RecordFieldGraph{Example: ExpressionToolOutputParameter{}}
	if err := parseObject(typeOfRecv, valueOfRecv, data, db); err != nil {
		return err
	}
	return nil
}

func (p *Workflow) UnmarshalJSON(data []byte) error {
	typeOfRecv := reflect.TypeOf(*p)
	valueOfRecv := reflect.ValueOf(p).Elem()
	db := make(map[string]*RecordFieldGraph)
	db["InputParameter"] = &RecordFieldGraph{Example: WorkflowInputParameter{}}
	db["OutputParameter"] = &RecordFieldGraph{Example: WorkflowOutputParameter{}}
	if err := parseObject(typeOfRecv, valueOfRecv, data, db); err != nil {
		return err
	}
	return nil
}
