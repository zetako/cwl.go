package conditionaltest

import "testing"

func TestCWLRec_RunId00(t *testing.T) {
	err := SwitchTestSet("/home/zetako/NSCC/personaltechdoc/cwl/DAGTest", "test-index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	test := tests[0]
	t.Log("Test00: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}

func TestCWLRec_RunId01(t *testing.T) {
	err := SwitchTestSet("/home/zetako/NSCC/personaltechdoc/cwl/DAGTest", "test-index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	test := tests[1]
	t.Log("Test01: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}

func TestCWLRec_RunId02(t *testing.T) {
	err := SwitchTestSet("/home/zetako/NSCC/personaltechdoc/cwl/DAGTest", "test-index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	test := tests[2]
	t.Log("Test02: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
