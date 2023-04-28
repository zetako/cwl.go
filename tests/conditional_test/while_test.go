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

func TestCWLWhile_TestAllBaseTool(t *testing.T) {
	err := SwitchTestSet("/home/zetako/NSCC/personaltechdoc/cwl/LoopFeature", "test-index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	tests := filterTests(TestDoc{Tags: []string{"base_tool"}})
	for _, test := range tests {
		t.Log("ID : ", test.ID)
		t.Log("Doc: ", test.Doc)
		doTest(t, test)
	}
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

func TestCWLWhile_RunIdx01(t *testing.T) {
	err := SwitchTestSet("/home/zetako/NSCC/personaltechdoc/cwl/LoopFeature", "test-index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	test := tests[1]
	t.Log("Test 01: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}

func TestCWLWhile_RunIdx02(t *testing.T) {
	err := SwitchTestSet("/home/zetako/NSCC/personaltechdoc/cwl/LoopFeature", "test-index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	test := tests[2]
	t.Log("Test 02: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}

func TestCWLWhile_RunIdx03(t *testing.T) {
	err := SwitchTestSet("/home/zetako/NSCC/personaltechdoc/cwl/LoopFeature", "test-index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	test := tests[3]
	t.Log("Test 03: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}

func TestCWLWhile_RunIdx04(t *testing.T) {
	err := SwitchTestSet("/home/zetako/NSCC/personaltechdoc/cwl/LoopFeature", "test-index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	test := tests[4]
	t.Log("Test 04: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}

func TestCWLWhile_RunIdx05(t *testing.T) {
	err := SwitchTestSet("/home/zetako/NSCC/personaltechdoc/cwl/LoopFeature", "test-index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	test := tests[5]
	t.Log("Test 05: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhile_RunIdx06(t *testing.T) {
	err := SwitchTestSet("/home/zetako/NSCC/personaltechdoc/cwl/LoopFeature", "test-index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	test := tests[6]
	t.Log("Test 06: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}

func TestCWLWhile_RunIdx07(t *testing.T) {
	err := SwitchTestSet("/home/zetako/NSCC/personaltechdoc/cwl/LoopFeature", "test-index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	test := tests[7]
	t.Log("Test 07: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}

func TestCWLWhile_RunId12(t *testing.T) {
	err := SwitchTestSet("/home/zetako/NSCC/personaltechdoc/cwl/LoopFeature", "test-index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	test := tests[12]
	t.Log("Test12: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhile_RunId13(t *testing.T) {
	err := SwitchTestSet("/home/zetako/NSCC/personaltechdoc/cwl/LoopFeature", "test-index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	test := tests[13]
	t.Log("Test13: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
