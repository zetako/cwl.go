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
