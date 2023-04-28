package runner

import (
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"github.com/robertkrimen/otto"
	"log"
)

const (
	WHILE_PREVIOUS = "previous"
	WHILE_PARENT   = "parent"
)

type iterData struct {
	Inputs, Outputs cwl.Values
}

func (i iterData) addToJSVM(vm *jsvm, name string) error {
	var (
		err error
	)

	obj, err := vm.vm.Object(name + " = {}")
	if err != nil {
		return fmt.Errorf("设置迭代记录%s失败: %v", name, err)
	}
	tmp, err := vm.cwlValuesToJsValue(i.Inputs)
	if err != nil {
		return fmt.Errorf("设置迭代记录%s失败: %v", name+".inputs", err)
	}
	err = obj.Set("inputs", tmp)
	if err != nil {
		return fmt.Errorf("设置迭代记录%s失败: %v", name+".inputs", err)
	}
	tmp, err = vm.cwlValuesToJsValue(i.Outputs)
	if err != nil {
		return fmt.Errorf("设置迭代记录%s失败: %v", name+".outputs", err)
	}
	err = obj.Set("outputs", tmp)
	if err != nil {
		return fmt.Errorf("设置迭代记录%s失败: %v", name+".outputs", err)
	}
	return nil
}

type iterDataArray []iterData

func (a iterDataArray) addToJSVM(vm *jsvm, name string) error {
	//err := vm.vm.Set(name, []iterData{})
	_, err := vm.vm.Object(name + "= []")
	if err != nil {
		return fmt.Errorf("设置迭代记录%s失败: %v", name, err)
	}
	for idx, data := range a {
		tmpName := fmt.Sprintf("%s[%d]", name, idx)
		err = data.addToJSVM(vm, tmpName)
		if err != nil {
			return fmt.Errorf("设置迭代记录%s失败: %v", tmpName, err)
		}
	}
	return nil
}

// setupLoop 为循环表达式的计算设置环境
//   - inputs: 该次计算的inputs
//   - index: 本次迭代的索引
//   - previous: 上次迭代的数据，如果传入nil，js中对应undefined
//   - parents: 已完成的所有迭代的数据，如果传入nil，js中对应undefined
func setupLoop(vm *jsvm, inputs cwl.Values, index int, previous iterData, parent iterDataArray) error {
	var (
		err error
	)

	if err = vm.setInputs(inputs); err != nil {
		return fmt.Errorf("设置输入失败: %v", err)
	}
	if err = vm.vm.Set("index", index); err != nil {
		return fmt.Errorf("设置迭代索引失败: %v", err)
	}
	if index > 0 { // 大于0的迭代才需要设置这些量
		if err = previous.addToJSVM(vm, WHILE_PREVIOUS); err != nil {
			return fmt.Errorf("设置前次迭代失败: %v", err)
		}
		if err = parent.addToJSVM(vm, WHILE_PARENT); err != nil {
			return fmt.Errorf("设置所有迭代失败: %v", err)
		}
	} else {
		if err = vm.vm.Set("previous", otto.UndefinedValue()); err != nil {
			return fmt.Errorf("设置前次迭代失败: %v", err)
		}
		if err = vm.vm.Set("parent", otto.UndefinedValue()); err != nil {
			return fmt.Errorf("设置所有迭代失败: %v", err)
		}
	}

	return nil
}

func (r *RegularRunner) RunLoop() (values cwl.Values, err error) {
	// 应该在上级函数中计算的步骤：
	//   - 检测while不能和scatter共存
	//   - 绑定输入和计算when

	// 计算中使用的变量
	var (
		index    int = 0
		previous iterData
		parent   []iterData
	)
	// 捕捉失败事件
	//defer func() {
	//	if err != nil {
	//		log.Printf("[Step \"%s\": Loop Index %d] Error:%v", r.step.ID, index, err)
	//	}
	//}()

	parent = make([]iterData, 0)
	// 循环
	for {
		log.Printf("[Step \"%s\": Loop Index %d] Start", r.step.ID, index)
		// 0. 设置运行时
		process, err := r.engine.GenerateSubProcess(r.step)
		if err != nil {
			return nil, fmt.Errorf("迭代进程%d创建失败: %v", index, err)
		}
		err = setupLoop(process.jsvm, *r.process.inputs, index, previous, parent)
		if err != nil {
			return nil, fmt.Errorf("迭代进程%d预处理失败: %v", index, err)
		}
		// 1. 计算while
		whileBool, err := process.jsvm.EvalBoolExpr(r.step.While)
		if err != nil {
			return nil, fmt.Errorf("迭代%d循环条件计算失败: %v", index, err)
		}
		if !whileBool {
			break
		}
		// 2. 计算iterationInputs
		inputs, err := process.jsvm.EvalValuesExpr(r.step.IterationInputs)
		if err != nil {
			return nil, fmt.Errorf("迭代%d循环输入计算失败: %v", index, err)
		}
		process.inputs = &inputs
		// 3. Run
		outputs, err := r.engine.RunProcess(process)
		if err != nil {
			return nil, fmt.Errorf("迭代%d运行失败: %v", index, err)
		}
		// 4. 准备下次运行
		log.Printf("[Step \"%s\": Loop Index %d] Finish", r.step.ID, index)
		index++
		previous = iterData{
			Inputs:  inputs,
			Outputs: outputs,
		}
		parent = append(parent, previous)
	}
	// 准备返回值
	if index == 0 { // 一次都没有执行，返回空值
		return cwl.Values{}, nil
	}
	outputs := cwl.Values{}
	for key, _ := range previous.Outputs { // 我们认为每个迭代具有相同的输出key
		arr := make([]cwl.Value, index)
		for idx, iter := range parent {
			arr[idx] = iter.Outputs[key]
		}
		outputs[key] = arr
	}
	return outputs, nil
}
