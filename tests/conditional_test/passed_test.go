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

func TestCWLWhen_RunIdx10(t *testing.T) {
	test := tests[10]
	t.Log("Test 00: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhen_RunIdx11(t *testing.T) {
	test := tests[11]
	t.Log("Test 11: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhen_RunIdx12(t *testing.T) {
	test := tests[12]
	t.Log("Test 12: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhen_RunIdx13(t *testing.T) {
	test := tests[13]
	t.Log("Test 13: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhen_RunIdx14(t *testing.T) {
	test := tests[14]
	t.Log("Test 14: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhen_RunIdx15(t *testing.T) {
	test := tests[15]
	t.Log("Test 15: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhen_RunIdx16(t *testing.T) {
	test := tests[16]
	t.Log("Test 16: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhen_RunIdx17(t *testing.T) {
	test := tests[17]
	t.Log("Test 17: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhen_RunIdx18(t *testing.T) {
	test := tests[18]
	t.Log("Test 18: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
func TestCWLWhen_RunIdx19(t *testing.T) {
	test := tests[19]
	t.Log("Test 19: ")
	t.Log("ID : ", test.ID)
	t.Log("Doc: ", test.Doc)
	doTest(t, test)
}
