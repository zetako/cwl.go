package conditionaltest

import "testing"

func TestCWLWhile_GetAllTests(t *testing.T) {
	err := SwitchTestSet("/home/zetako/NSCC/personaltechdoc/cwl/LoopFeature", "test-index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Total: ", len(tests))
	toPrint := "\n"
	for _, test := range tests {
		toPrint += test.ID + "\n"
	}
	t.Log(toPrint)
}
func TestCWLWhile_RunIdx00(t *testing.T) {
	err := SwitchTestSet("/home/zetako/NSCC/personaltechdoc/cwl/LoopFeature", "test-index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	test := tests[0]
	t.Log("Test 00: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
