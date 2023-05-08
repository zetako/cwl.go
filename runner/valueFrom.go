package runner

import (
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"github.com/robertkrimen/otto"
	"path"
)

// evalValueFrom 以 cwl.Value 的形式计算表达式
func (vm *jsvm) evalValueFrom(expr cwl.Expression, self cwl.Value) (cwl.Value, error) {
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
func (vm *jsvm) setInputs(inputs cwl.Values) error {
	return vm.setValueAs(inputs, "inputs")
}

// setValuesAs 将一个cwl.Value设置为虚拟机中的map
func (vm *jsvm) setValueAs(values cwl.Values, name string) error {
	// 转换为一般结构
	plain, err := toJSONMap(values)
	if err != nil {
		return err
	}
	plainMap := plain.(map[string]interface{})
	// 将nil值转换为js中的null
	for key, value := range plainMap {
		if value == nil {
			plainMap[key] = otto.NullValue()
		}
	}
	// 设置变量
	err = vm.vm.Set(name, plainMap)
	return err
}

// cwlValuesToJsValue 将cwl.values转换为jsvm的value
func (vm *jsvm) cwlValuesToJsValue(values cwl.Values) (interface{}, error) {
	if values == nil {
		return otto.NullValue(), nil
	}
	// 转换为一般结构
	plain, err := toJSONMap(values)
	if err != nil {
		return otto.UndefinedValue(), err
	}
	plainMap := plain.(map[string]interface{})
	// 将nil值转换为js中的null
	for key, value := range plainMap {
		if value == nil {
			plainMap[key] = otto.NullValue()
		}
	}
	// 设置变量
	return vm.vm.ToValue(plainMap)
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
		// 使用该方法时，即使没有符合条件的值也应该输出一个空列表
		return ret, nil
	case cwl.LAST_NON_NULL:
		for i := len(arr) - 1; i >= 0; i-- {
			if arr[i] != nil {
				return arr[i], nil
			}
		}
		return value, fmt.Errorf("pickValue=last_non_null，但是不存在非空值")
	default:
		// 不应该有default,这里就什么都不做吧！
		return value, nil
	}
}

// TODO: 这个函数需要返回是否尝试使用默认值
func linkMerge(method cwl.LinkMergeMethod, sources cwl.ArrayString, values cwl.Values) (cwl.Value, error) {
	if len(sources) == 1 {
		// FIXME: 这里应该做能否取到值的验证
		return values[sources[0]], nil
	} else {
		switch method {
		case cwl.MERGE_FLATTENED:
			var tmpInput []cwl.Value
			for _, src := range sources {
				tmp, ok := values[src]
				if !ok { // 拿不到可能是When没有执行，设为nil
					tmp = nil
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
				}
				tmpInput = append(tmpInput, tmp)
			}
			return tmpInput, nil
		default: // FIXME: 这里是错的，default就应该是merge_nested
			return values[sources[0]], nil
		}
	}
}

func (vm *jsvm) EvalBoolExpr(expression cwl.Expression) (bool, error) {
	tmp, err := vm.Eval(expression, nil)
	if err != nil {
		return false, err
	}
	tmpBool, ok := tmp.(bool)
	if !ok {
		return false, fmt.Errorf("表达式计算结果不是布尔值")
	}
	return tmpBool, nil
}

func (vm *jsvm) EvalValuesExpr(expression cwl.Expression) (cwl.Values, error) {
	tmp, err := vm.Eval(expression, nil)
	if err != nil {
		return nil, err
	}
	valMap, ok := tmp.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("表达式计算结果不是对象")
	}
	values := cwl.Values{}
	for key, interfaceValue := range valMap {
		value, err := cwl.ConvertToValue(interfaceValue)
		if err != nil {
			return nil, err
		}
		values[key] = value
	}
	return values, nil
}
