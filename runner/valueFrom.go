package runner

import "github.com/lijiang2014/cwl.go"

func evalValueFrom(vm *jsvm, expr cwl.Expression, self cwl.Value) (cwl.Value, error) {
	plain, err := toJSONMap(self)
	if err != nil {
		return nil, err
	}
	plain, err = vm.Eval(expr, plain)
	if err != nil {
		return nil, err
	}
	ret, err := cwl.ConvertToValue(plain)
	return ret, err
}

func setInputs(vm *jsvm, inputs cwl.Values) error {
	plain, err := toJSONMap(inputs)
	if err != nil {
		return err
	}
	err = vm.vm.Set("inputs", plain)
	return err
}
