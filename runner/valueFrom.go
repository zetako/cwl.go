package runner

import (
	"github.com/lijiang2014/cwl.go"
	"path"
)

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

// preprocessInputs 计算ValueFrom之前预处理输入
//   - 可能需要用到File的Nameroot和Nameext
func preprocessInputs(inputs *cwl.Values) error {
	for key, value := range *inputs {
		if fileValue, ok := value.(cwl.File); ok {
			// 取值，优先Location，然后是Path
			target := fileValue.Location
			if target == "" {
				target = fileValue.Path
			}
			// 如果还是空，就不处理
			if target == "" {
				continue
			}
			fileValue.Basename = path.Base(target)
			fileValue.Dirname = path.Dir(target)
			fileValue.Nameext = path.Ext(target)
			fileValue.Nameroot = fileValue.Basename[:len(fileValue.Basename)-len(fileValue.Nameext)]
			(*inputs)[key] = fileValue
		}
	}
	return nil
}
