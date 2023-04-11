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

// $Graph as json
// 有奇怪的行为（前置#）
//func TestCWLR2_run110(t *testing.T) {
//	testByID(t, 110)
//}

// InitialWorkDirRequirement base on other step
// 与目前各步骤独立目录的行为有冲突
//func TestCWLR2_run122(t *testing.T) {
//	testByID(t, 122)
//}

// 动态Resources
// 要求在ResourcesLimit阶段计算js表达式
//	func TestCWLR2_run126(t *testing.T) {
//		testByID(t, 126)
//	}
