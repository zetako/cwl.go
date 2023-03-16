package runner

import (
	"fmt"

	"github.com/lijiang2014/cwl.go"
)

func (p *Process) bindEnvVar(req *cwl.EnvVarRequirement) (err error) {
	for _, defi := range req.EnvDef {
		value, err := p.eval(defi.EnvValue, nil)
		if err != nil {
			return err
		}
		p.env[defi.EnvName] = fmt.Sprint(value)
	}
	return nil
}
