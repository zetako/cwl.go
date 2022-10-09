package cwlgotest

import (
	"testing"

	cwl "github.com/lijiang2014/cwl.go"
	. "github.com/otiai10/mint"
)

func TestDecode_starlight1(t *testing.T) {
	f := load("starlight-1.cwl")
	root := cwl.NewCWL()
	err := root.Decode(f)
	Expect(t, err).ToBe(nil)
	Expect(t, root.Version).ToBe("v1.0")
	Expect(t, root.Class).ToBe("CommandLineTool")
	Expect(t, root.Inputs[0].ID).ToBe("in")
	Expect(t, root.Inputs[0].Types[0].Type).ToBe("Any")
	// TODO check specification for this test ID and Type
	Expect(t, root.Outputs[0].ID).ToBe("out")
	Expect(t, root.Outputs[0].Types[0].Type).ToBe("string")
	Expect(t, root.Outputs[0].Binding.Glob[0]).ToBe("out.txt")
	Expect(t, root.Outputs[0].Binding.LoadContents).ToBe(true)
	Expect(t, root.BaseCommands[0]).ToBe("echo")
	Expect(t, root.Stdout).ToBe("out.txt")

	// Expect(t, root.Requirements[0].ClassBase).ToBe("DockerRequirement")
	Expect(t, root.Requirements[0].Class).ToBe("ScedulerRequirement")
	//Expect(t, root.Requirements[0].Sceduler).ToBe("slurm")
	//Expect(t, root.Requirements[0].Partition).ToBe("work")

}
