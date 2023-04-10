package runnertest

import (
	"log"
	"strconv"
	"testing"
)

func TestCountCWLR2_workflow_tag(t *testing.T) {
	allWorkflowTest := filterTests(TestDoc{Tags: []string{"workflow"}})
	log.Printf("All workflow count: %d\n", len(allWorkflowTest))
	allWfTestIDStr := ""
	count := 0
	for _, test := range allWorkflowTest {
		count++
		allWfTestIDStr = allWfTestIDStr + strconv.Itoa(test.ID) + ", "
		if count == 10 {
			allWfTestIDStr += "\n"
			count = 0
		}
	}
	log.Printf("All workflow:\n%s\n", allWfTestIDStr)
}

//	var workflowRelatedTest = [...]int{
//		20, 24, 25, 26, 27, 28, 29, 30, 31, 32,
//		33, 35, 36, 37, 38, 39, 40, 41, 42, 43,
//		45, 46, 47, 48, 49, 50, 51, 52, 53, 60,
//		70, 71, 72, 77, 78, 79, 80, 81, 82, 83,
//		97, 99, 110, 111, 113, 114, 122, 126, 128, 131,
//		132,
//	}

func TestCWLR2_run70(t *testing.T) {
	testByID(t, 70)
}
func TestCWLR2_run71(t *testing.T) {
	testByID(t, 71)
}
func TestCWLR2_run72(t *testing.T) {
	testByID(t, 72)
}
func TestCWLR2_run80(t *testing.T) {
	testByID(t, 80)
}
func TestCWLR2_run81(t *testing.T) {
	testByID(t, 81)
}
func TestCWLR2_run82(t *testing.T) {
	testByID(t, 82)
}

// sub-workflow
func TestCWLR2_run99(t *testing.T) {
	testByID(t, 99)
}

// $Graph
func TestCWLR2_run110(t *testing.T) {
	testByID(t, 110)
}

// generate nameroot & nameext
func TestCWLR2_run111(t *testing.T) {
	testByID(t, 111)
}

// Multiple source w/ multiple type
func TestCWLR2_run114(t *testing.T) {
	testByID(t, 114)
}

// InitialWorkDirRequirement base on other step
func TestCWLR2_run122(t *testing.T) {
	testByID(t, 122)
}
