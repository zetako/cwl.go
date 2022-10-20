package runnertest

import "testing"

// ✅ 1，2, 4
// ❌ 3, 5, 6

func testByID (t *testing.T, id int)  {
  toTests := filterTests(TestDoc{ID: id})
  if len(toTests) == 0{
    t.Fatalf("test not found ")
  }
  doc := toTests[0]
  doTest(t, doc)
}

func TestCWLR2_run1(t *testing.T) {
  testByID(t, 1)
}

func TestCWLR2_run2(t *testing.T) {
  testByID(t, 2)
}

// ✅ 1017 完成
func TestCWLR2_run3(t *testing.T) {
 testByID(t, 3)
}


func TestCWLR2_run4(t *testing.T) {
  testByID(t, 4)
}

// ✅ 1017
func TestCWLR2_run5(t *testing.T) {
  testByID(t, 5)
}

// ✅ 1019
// $include JS FILE
// expressionLib
// InitialWorkDirRequirement
// v1.0/template-tool.cwl v1.0/cat-job.json
func TestCWLR2_run6(t *testing.T) {
  testByID(t, 6)
}


// ✅ 1019
func TestCWLR2_run7(t *testing.T) {
  testByID(t, 7)
}

// ✅ 1019
func TestCWLR2_run8(t *testing.T) {
  testByID(t, 8)
}

// ✅ 1019
// stdName
func TestCWLR2_run9(t *testing.T) {
  testByID(t, 9)
}

// ✅ 1019
// stderr
func TestCWLR2_run10(t *testing.T) {
  testByID(t, 10)
}

func TestCWLR2_run11(t *testing.T) {
  testByID(t, 11)
}

func TestCWLR2_run12(t *testing.T) {
  testByID(t, 12)
}


func TestCWLR2_run13(t *testing.T) {
  testByID(t, 13)
}

// ✅ 1019 expressionTool
func TestCWLR2_run14(t *testing.T) {
  testByID(t, 14)
}

// ✅ 1019 expressionTool input set null
func TestCWLR2_run15(t *testing.T) {
  testByID(t, 15)
}

func TestCWLR2_run16(t *testing.T) {
  testByID(t, 16)
}

func TestCWLR2_run17(t *testing.T) {
  testByID(t, 17)
}

func TestCWLR2_run18(t *testing.T) {
  testByID(t, 18)
}

func TestCWLR2_run19(t *testing.T) {
  testByID(t, 19)
}

//// workflow
//// ❌ any-type-compat.cwl v1.0/any-type-job.json
//func TestCWLR2_run20(t *testing.T) {
//  testByID(t, 20)
//}

func TestCWLR2_run21(t *testing.T) {
  testByID(t, 21)
}

// ✅ 1019 ExpressionTool input loadContents
func TestCWLR2_run22(t *testing.T) {
  testByID(t, 22)
}

// 23,34,44,54,57,58,68
// 69,73,76,84-96,98,100
// 103-109,112,115-118,120,123,127,129

// ✅ outputEval
func TestCWLR2_run23(t *testing.T) {
  testByID(t, 23)
}

// ✅ 1020 envTool
func TestCWLR2_run34(t *testing.T) {
  testByID(t, 34)
}

// ✅ 1020 inputTypeAny
func TestCWLR2_run44(t *testing.T) {
  testByID(t, 44)
}

// ✅ 1020 outputEval
func TestCWLR2_run58(t *testing.T) {
  testByID(t, 58)
}

// ✅ 1020 null input
func TestCWLR2_run68(t *testing.T) {
  testByID(t, 68)
}

func TestCWLR2_run69(t *testing.T) {
  testByID(t, 69)
}

// 69,73,76,84-96,98,100
// 103-109,112,115-118,120,123,127,129
