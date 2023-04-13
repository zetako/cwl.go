package conditionaltest

import "testing"

func TestCWLWhen_GetAllTestID(t *testing.T) {
	t.Log("Total: ", len(tests))
	toPrint := "\n"
	for _, test := range tests {
		toPrint += test.ID + "\n"
	}
	t.Log(toPrint)
}

func TestCWLWhen_RunIdx00(t *testing.T) {
	test := tests[0]
	t.Log("Test 00: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhen_RunIdx01(t *testing.T) {
	test := tests[1]
	t.Log("Test 01: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhen_RunIdx02(t *testing.T) {
	test := tests[2]
	t.Log("Test 02: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhen_RunIdx03(t *testing.T) {
	test := tests[3]
	t.Log("Test 03: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhen_RunIdx04(t *testing.T) {
	test := tests[4]
	t.Log("Test 04: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhen_RunIdx05(t *testing.T) {
	test := tests[5]
	t.Log("Test 05: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhen_RunIdx06(t *testing.T) {
	test := tests[6]
	t.Log("Test 06: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhen_RunIdx07(t *testing.T) {
	test := tests[7]
	t.Log("Test 07: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhen_RunIdx08(t *testing.T) {
	test := tests[8]
	t.Log("Test 08: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhen_RunIdx09(t *testing.T) {
	test := tests[9]
	t.Log("Test 09: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
