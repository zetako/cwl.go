package runnertest

import "testing"

// time go test -timeout 100s -run ^TestCWLR2_run github.com/lijiang2014/cwl.go/tests/runner_test

func testByID(t *testing.T, id int) {
	toTests := filterTests(TestDoc{ID: id})
	if len(toTests) == 0 {
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

// ✅ 3/14
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

// // workflow
// ✅ workflow any type in
func TestCWLR2_run20(t *testing.T) {
	testByID(t, 20)
}

// ✅ 3/14
func TestCWLR2_run21(t *testing.T) {
	testByID(t, 21)
}

// ✅ 1019 ExpressionTool input loadContents
func TestCWLR2_run22(t *testing.T) {
	testByID(t, 22)
}

// 23,34,44,54,57,58,68 ,69,
// 73,76,84-96,98,100
// 103-109,112,115-118,120,123,127,129

// ✅ outputEval
func TestCWLR2_run23(t *testing.T) {
	testByID(t, 23)
}

// ✅ workflow with cwl file step
func TestCWLR2_run24(t *testing.T) {
	testByID(t, 24)
}

// ✅ workflow with embedded step
func TestCWLR2_run25(t *testing.T) {
	testByID(t, 25)
}

// ✅ workflow with scatter step; multiple parameters
func TestCWLR2_run26(t *testing.T) {
	testByID(t, 26)
}

// ✅ workflow with scatter step; multiple sources
func TestCWLR2_run27(t *testing.T) {
	testByID(t, 27)
}

// ✅ workflow with scatter step; multiple sources in array type
func TestCWLR2_run28(t *testing.T) {
	testByID(t, 28)
}

// ✅ workflow with linkMerge param
func TestCWLR2_run29(t *testing.T) {
	testByID(t, 29)
}

// ✅ workflow no multipleInputFeatureRequirement
func TestCWLR2_run30(t *testing.T) {
	testByID(t, 30)
}

// ✅ workflow empty input
func TestCWLR2_run31(t *testing.T) {
	testByID(t, 31)
}

// ✅ workflow default + has input
func TestCWLR2_run32(t *testing.T) {
	testByID(t, 32)
}

// ✅ workflow default + use default
func TestCWLR2_run33(t *testing.T) {
	testByID(t, 33)
}

// ✅ 1020 envTool 3/14
func TestCWLR2_run34(t *testing.T) {
	testByID(t, 34)
}

// ✅ simple scatter
func TestCWLR2_run35(t *testing.T) {
	testByID(t, 35)
}

// ✅ nest cross product
func TestCWLR2_run36(t *testing.T) {
	testByID(t, 36)
}

// ✅ flat cross product
func TestCWLR2_run37(t *testing.T) {
	testByID(t, 37)
}

// ✅ dot product
func TestCWLR2_run38(t *testing.T) {
	testByID(t, 38)
}

// ✅ empty @ single scatter
func TestCWLR2_run39(t *testing.T) {
	testByID(t, 39)
}

// ✅ empty 2nd param @ nest cross product
func TestCWLR2_run40(t *testing.T) {
	testByID(t, 40)
}

// ✅ empty 1nd param @ nest cross product
func TestCWLR2_run41(t *testing.T) {
	testByID(t, 41)
}

// ✅ empty any param @ flat cross product
func TestCWLR2_run42(t *testing.T) {
	testByID(t, 42)
}

// ✅ empty both param @ dot product
func TestCWLR2_run43(t *testing.T) {
	testByID(t, 43)
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

// ✅ 1021 multi Expr inline
func TestCWLR2_run69(t *testing.T) {
	testByID(t, 69)
}

// 73,76,84-96,98,100
// 103-109,112,115-118,120,123,127,129

// record_output_binding v1.0/record-output.cwl v1.0/record-output-job.json
func TestCWLR2_run73(t *testing.T) {
	testByID(t, 73)
}

// multiple_glob_expr_list  v1.0/glob-expr-list.cwl v1.0/abc.json
// ✅ 3/14
func TestCWLR2_run76(t *testing.T) {
	testByID(t, 76)
}

// dir job 84 ~ 88 ❌ 后置？
// directory_input_param_ref block
func TestCWLR2_run84(t *testing.T) {
	testByID(t, 84)
}

// ✅ 3/20
func TestCWLR2_run85(t *testing.T) {
	testByID(t, 85)
}

func TestCWLR2_run86(t *testing.T) {
	testByID(t, 86)
}

func TestCWLR2_run87(t *testing.T) {
	testByID(t, 87)
}

func TestCWLR2_run88(t *testing.T) {
	testByID(t, 88)
}

// writable_stagedfiles ✅ 3/14
func TestCWLR2_run89(t *testing.T) {
	testByID(t, 89)
}

// input_file_literal ✅ 03/15
func TestCWLR2_run90(t *testing.T) {
	testByID(t, 90)
}

// ✅ 03/15
func TestCWLR2_run91(t *testing.T) {
	testByID(t, 91)
}

func TestCWLR2_run92(t *testing.T) {
	testByID(t, 92)
}

// ✅ 03/15
func TestCWLR2_run93(t *testing.T) {
	testByID(t, 93)
}

func TestCWLR2_run94(t *testing.T) {
	testByID(t, 94)
}

func TestCWLR2_run95(t *testing.T) {
	testByID(t, 95)
}

func TestCWLR2_run96(t *testing.T) {
	testByID(t, 96)
}

// ✅ initial_workdir_output
// v1.0/initialworkdirrequirement-docker-out.cwl v1.0/initialworkdirrequirement-docker-out-job.json
func TestCWLR2_run98(t *testing.T) {
	testByID(t, 98)
}

// ❌ filesarray_secondaryfiles v1.0/docker-array-secondaryfiles.cwl v1.0/docker-array-secondaryfiles-job.json
func TestCWLR2_run100(t *testing.T) {
	testByID(t, 100)
}

// ✅ 3/17 dockeroutputdir v1.0/docker-output-dir.cwl v1.0/empty.json
func TestCWLR2_run103(t *testing.T) {
	testByID(t, 103)
}

// ✅ 3/15 hit imports 3/16 env hits
func TestCWLR2_run104(t *testing.T) {
	testByID(t, 104)
}

func TestCWLR2_run105(t *testing.T) {
	testByID(t, 105)
}

func TestCWLR2_run106(t *testing.T) {
	testByID(t, 105)
}

// ✅ 3/16 input_dir_recurs_copy_writable v1.0/recursive-input-directory.cwl v1.0/recursive-input-directory.yml
func TestCWLR2_run107(t *testing.T) {
	testByID(t, 107)
}

// ✅ 3/16 null_missing_params v1.0/null-defined.cwl v1.0/empty.json
func TestCWLR2_run108(t *testing.T) {
	testByID(t, 108)
}

func TestCWLR2_run109(t *testing.T) {
	testByID(t, 109)
}

func TestCWLR2_run112(t *testing.T) {
	testByID(t, 112)
}

func TestCWLR2_run115(t *testing.T) {
	testByID(t, 115)
}

// ✅ 3/17 shelldir_quoted v1.0/shellchar2.cwl v1.0/empty.json
func TestCWLR2_run116(t *testing.T) {
	testByID(t, 116)
}

// ✅ 3/17 initial_workdir_empty_writable v1.0/writable-dir.cwl v1.0/empty.json
func TestCWLR2_run117(t *testing.T) {
	testByID(t, 117)
}

// ❌ initial_workdir_empty_writable_docker v1.0/writable-dir-docker.cwl v1.0/empty.json
func TestCWLR2_run118(t *testing.T) {
	testByID(t, 118)
}

// fileliteral_input_docker v1.0/cat3-nodocker.cwl v1.0/file-literal.yml
func TestCWLR2_run120(t *testing.T) {
	testByID(t, 120)
}

func TestCWLR2_run123(t *testing.T) {
	testByID(t, 123)
}

func TestCWLR2_run127(t *testing.T) {
	testByID(t, 127)
}

// ✅3/17 valuefrom const overides array inputs
func TestCWLR2_run129(t *testing.T) {
	testByID(t, 129)
}
