package runner

import (
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"github.com/robertkrimen/otto"
	"path"
)

// evalValueFrom 以 cwl.Value 的形式计算表达式
// TODO 重构为jsvm的方法
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

// setInputs 将 cwl.Values 设置为jsvm中的"$inputs"
// TODO 重构为jsvm的方法
func setInputs(vm *jsvm, inputs cwl.Values) error {
	// 如果有nil，需要替换为js的null
	plain, err := toJSONMap(inputs)
	if err != nil {
		return err
	}

	plainMap := plain.(map[string]interface{})
	for key, value := range plainMap {
		if value == nil {
			plainMap[key] = otto.NullValue()
		}
	}
	err = vm.vm.Set("inputs", plainMap)
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

// pickValue 计算pickValue表达式
//   - 将不需要计算的量以及其对应的空Method传进来是安全的
func pickValue(value cwl.Value, method cwl.PickValueMethod) (cwl.Value, error) {
	if method == "" {
		return value, nil
	}
	arr, ok := value.([]cwl.Value)
	if !ok {
		return value, fmt.Errorf("无法对非数组对象执行PickValue")
	}
	switch method {
	case cwl.FIRST_NON_NULL:
		for _, entity := range arr {
			if entity != nil {
				return entity, nil
			}
		}
		return value, fmt.Errorf("pickValue=first_non_null，但是不存在非空值")
	case cwl.THE_ONLY_NON_NULL:
		var ret cwl.Value = nil
		for _, entity := range arr {
			if entity != nil {
				if ret != nil {
					return value, fmt.Errorf("pickValue=the_only_non_null，但是存在多个非空值")
				}
				ret = entity
			}
		}
		if ret != nil {
			return ret, nil
		}
		return ret, fmt.Errorf("pickValue=the_only_non_null，但是不存在非空值")
	case cwl.ALL_NON_NULL:
		var ret []cwl.Value
		for _, entity := range arr {
			if entity != nil {
				ret = append(ret, entity)
			}
		}
		return ret, nil
		// 使用该方法时，即使没有符合条件的值也应该输出一个空列表
		//if len(ret) > 0 {
		//	return ret, nil
		//}
		//return ret, fmt.Errorf("pickValue=all_non_null，但是不存在非空值")
	default:
		// 不应该有default,这里就什么都不做吧！
		return value, nil
	}
}

func linkMerge(method cwl.LinkMergeMethod, sources cwl.ArrayString, values cwl.Values) (cwl.Value, error) {
	if len(sources) == 1 {
		return values[sources[0]], nil
	} else {
		switch method {
		case cwl.MERGE_FLATTENED:
			var tmpInput []cwl.Value
			for _, src := range sources {
				tmp, ok := values[src]
				if !ok { // 拿不到可能是When没有执行，设为nil
					tmp = nil
					//return nil, errors.New("没有匹配 " + src + " 的输入")
				}
				if tmpList, ok := tmp.([]cwl.Value); ok {
					tmpInput = append(tmpInput, tmpList...)
				} else { // 也有可能是单个元素
					tmpInput = append(tmpInput, tmp)
				}
			}
			return tmpInput, nil
		case cwl.MERGE_NESTED:
			var tmpInput []cwl.Value
			for _, src := range sources {
				tmp, ok := values[src]
				if !ok { // 拿不到可能是When没有执行，设为nil
					tmp = nil
					//return nil, errors.New("没有匹配 " + src + " 的输入")
				}
				tmpInput = append(tmpInput, tmp)
			}
			return tmpInput, nil
		default:
			return values[sources[0]], nil
		}
	}
}
