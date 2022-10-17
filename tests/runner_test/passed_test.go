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

// ✅ 1017  TODO
func TestCWLR2_run5(t *testing.T) {
  testByID(t, 5)
}

// FAILED TODO
// $include JS FILE
func TestCWLR2_run6(t *testing.T) {
  testByID(t, 6)
}