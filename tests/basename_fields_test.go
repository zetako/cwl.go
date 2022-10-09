package cwlgotest

import (
	"testing"

	cwl "github.com/lijiang2014/cwl.go"
	. "github.com/otiai10/mint"
)

func TestCWL_basename_fields_test_test(t *testing.T) {
	f := load("basename-fields-test.cwl")
	doc := cwl.NewCWL()
	err := doc.Decode(f)
	root := doc.Process.(*cwl.Workflow)
	Expect(t, err).ToBe(nil)
	Expect(t, root.CWLVersion).ToBe("v1.0")
	Expect(t, root.Class).ToBe("Workflow")

	Expect(t, root.Requirements[0].ClassName()).ToBe("StepInputExpressionRequirement")
	Expect(t, root.Requirements[0])
	t.Logf("%#v", root)
	t.Logf("%#v", root.Steps[0])

}
