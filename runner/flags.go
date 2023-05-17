package runner

import "time"

// EngineFlags is a set of flags controlling engine's processing
// The zero value of this set can be used as default
type EngineFlags struct {
	// Modify CWL Feature (+loop)
	DisablePlusLoop    bool // Switch of overall +loop modify
	DisableLoopFeature bool // Switch of LoopFeatureRequirement
	DisableLastNonNull bool // Switch of PickValue method last_non_null
	MaxLoopCount       int  // Max loop count allowed, exceed the limit cause an error

	// Workflow Related
	MaxWorkflowNested int // Max sub-workflow nested count, exceed the limit cause an error
	MaxScatterLimit   int // Max scatter count, engine will split scatter tasks to multiple groups if this exceeded
	MaxParallelLimit  int // Max step running parallel count, exceeded step just wait for next start

	// Runtime Related
	TotalTimeLimit time.Duration // The overall running time limit of whole process
	StepTimeLimit  time.Duration // The seperated running time limit of each step
}

const DefaultWorkflowNested = 15
