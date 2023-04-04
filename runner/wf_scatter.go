package runner

import (
	"errors"
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"path"
)

// RunScatter 运行需要分发任务的步骤
func (r *RegularRunner) RunScatter(condition chan<- Condition) (err error) {
	var (
		totalTask    int
		runningTask  int            = 0
		internalChan chan Condition = make(chan Condition)
		process      *Process       = nil
		isSuccess    bool           = true // 标记所有任务是否成功
		allInputs    []cwl.Values
		allOutputs   []ScatterDoneCondition // 用以存储每一步的输出
		output       cwl.Values
		layout       []int
	)
	totalTask, allInputs, layout, err = r.getAllScatterInputs()
	for i := 0; i < totalTask; i++ {
		// 1. Scatter的每个任务都需要创建Process
		process, err = r.engine.GenerateSubProcess(r.step)
		if err != nil {
			isSuccess = false
			break
		}
		// 2. 设置输出 & 绑定输入
		process.runtime.RootHost = path.Join(process.runtime.RootHost, fmt.Sprintf("scatter%d", i))
		process.outputFS = &Local{
			workdir:      process.runtime.RootHost,
			CalcChecksum: false,
		}
		process.loadRuntime()
		process.inputs = &allInputs[i]

		// 3. 并行执行（?）
		go r.scatterTaskWrapper(process, internalChan, i)
		runningTask++
	}
	if !isSuccess { // 任务启动失败，等待收回所有任务
		// a. 等待所有任务结束
		for runningTask > 0 {
			tmpCondition := <-internalChan
			if _, ok := tmpCondition.(*ScatterDoneCondition); ok { // 该步正常结束
				runningTask--
			} else if _, ok = tmpCondition.(*ScatterErrorCondition); ok { // 该步异常结束
				// TODO 需要一种将该步错误传出的方法
				runningTask--
			}
		}
		// b. 返回一个错误
		condition <- &StepErrorCondition{
			step: r.step,
			err:  errors.New("任务分发启动失败"),
		}
		return errors.New("任务分发启动失败")
	}
	// 4. 等待结束
	for runningTask > 0 {
		tmpCondition := <-internalChan
		if doneCondition, ok := tmpCondition.(*ScatterDoneCondition); ok {
			allOutputs = append(allOutputs, *doneCondition)
			runningTask--
		} else if _, ok = tmpCondition.(*ScatterErrorCondition); ok { // 该步异常结束
			runningTask--
			isSuccess = false
			break
		}
	}
	if !isSuccess { // 任务结束失败，等待收回所有任务
		// a. 等待所有任务结束
		for runningTask > 0 {
			tmpCondition := <-internalChan
			if _, ok := tmpCondition.(*ScatterDoneCondition); ok { // 该步正常结束
				runningTask--
			} else if _, ok = tmpCondition.(*ScatterErrorCondition); ok { // 该步异常结束
				// TODO 需要一种将该步错误传出的方法
				runningTask--
			}
		}
		// b. 返回一个错误
		condition <- &StepErrorCondition{
			step: r.step,
			err:  errors.New("分发的任务执行失败"),
		}
		return errors.New("分发的任务执行失败")
	}
	// 5. 整理结果
	output = cwl.Values{}
	for _, doneCond := range allOutputs {
		for key, value := range *doneCond.out {
			if _, ok := output[key]; !ok {
				//output[key] = []cwl.Value{}
				output[key] = make([]cwl.Value, totalTask)
			}
			output[key].([]cwl.Value)[doneCond.scatterID] = value
		}
	}
	output, err = reconstructOutput(output, layout)
	if err != nil {
		condition <- &StepErrorCondition{
			step: r.step,
			err:  errors.New("输出格式化失败"),
		}
	}
	condition <- &StepDoneCondition{
		step: r.step,
		out:  &output,
	}
	return nil
}

// scatterTaskWrapper 已分发任务的运行封装，用于在协程中运行；使用channel传递结果
func (r *RegularRunner) scatterTaskWrapper(p *Process, condChan chan Condition, ID int) {
	out, err := r.engine.RunProcess(p)
	if err != nil {
		condChan <- &ScatterErrorCondition{
			scatterID: ID,
			err:       err,
		}
	}
	condChan <- &ScatterDoneCondition{
		scatterID: ID,
		out:       &out,
	}
	return
}

// needScatter 判断对应ID的变量是否需要分发
func (r *RegularRunner) needScatter(key string) bool {
	for _, entity := range r.step.Scatter {
		if entity == key {
			return true
		}
	}
	return false
}

// ScatterErrorCondition 分发任务结束的状态
type ScatterErrorCondition struct {
	scatterID int
	err       error
}

// Meet 无意义的判断，始终满足
func (ScatterErrorCondition) Meet(condition []Condition) bool {
	return true
}

// ScatterDoneCondition 分发任务结束的状态
type ScatterDoneCondition struct {
	scatterID int
	out       *cwl.Values
}

// Meet 无意义的判断，始终满足
func (ScatterDoneCondition) Meet(condition []Condition) bool {
	return true
}

// getTotalScatterTask 获取Scatter任务的总数
func (r *RegularRunner) getTotalScatterTask() int {
	var (
		scatterTarget string
		scatterSource []string
		ret           int
	)
	scatterTarget = r.step.Scatter[0]
	for _, inEntity := range r.step.In {
		if inEntity.ID == scatterTarget {
			scatterSource = inEntity.Source
		}
	}
	ret = 0
	for _, src := range scatterSource {
		bindSrc := (*r.parameter)[src]
		if bindSrcAsArray, ok := bindSrc.([]cwl.Value); ok {
			ret += len(bindSrcAsArray)
		} else {
			ret++
		}
	}
	return ret
}

func (r *RegularRunner) getAllScatterInputs() (total int, inputs []cwl.Values, layout []int, err error) {
	var (
		scatterCount   int
		scatterTargets []string
		scatterSinks   = map[string]cwl.Sink{}
		scatterSources = map[string]cwl.ArrayString{}
		scatterValues  = map[string][]cwl.Value{}
		scatterInputs  []cwl.Values
		scatterLayout  []int
	)
	// 1. 计算scatter总数
	// 1.1 获取需要scatter的key的所有的source
	scatterTargets = r.step.Scatter
	for _, target := range scatterTargets {
		scatterSources[target] = nil
	}
	for _, inEntity := range r.step.In {
		if _, ok := scatterSources[inEntity.ID]; ok {
			scatterSinks[inEntity.ID] = inEntity.Sink
			scatterSources[inEntity.ID] = inEntity.Source
		}
	}
	// 1.2 查找参数，根据这些source创造一个values映射
	//    - 根据 LinkMerge 的取值来确定是否需要展开数组
	for key, sources := range scatterSources {
		// 先为每个key分配对应的空间
		scatterValues[key] = []cwl.Value{}
		// 然后根据source数量判断行为
		if len(sources) == 1 { // 仅有单一source
			source := sources[0]
			tmp, ok := (*r.parameter)[source]
			if !ok {
				return -1, nil, nil, errors.New("没有匹配的输入")
			}
			if tmpList, ok := tmp.([]cwl.Value); ok {
				scatterValues[key] = append(scatterValues[key], tmpList...)
			} else { // 仅有一个source，source又不是数组，就没法分发了
				return -1, nil, nil, errors.New("没有需要分发的输入")
			}
		} else { // 有多个source，查看linkMerge方法
			for _, source := range sources {
				tmp, ok := (*r.parameter)[source]
				if !ok {
					return -1, nil, nil, errors.New("没有匹配的输入")
				}
				switch scatterSinks[key].LinkMerge {
				case cwl.MERGE_FLATTENED: // 需要拆分数组
					if tmpList, ok := tmp.([]cwl.Value); ok {
						scatterValues[key] = append(scatterValues[key], tmpList...)
					} else { // 也有可能是单个元素
						scatterValues[key] = append(scatterValues[key], tmp)
					}
					break
				case cwl.MERGE_NESTED: // 不需要拆分数组
					fallthrough
				default:
					scatterValues[key] = append(scatterValues[key], tmp)
					break
				}
			}
		}
	}
	// 1.3 计算出总scatter量和layout
	//   - 需要根据ScatterMethod方法来实现
	switch r.step.ScatterMethod {
	case cwl.NESTED_CROSSPRODUCT:
		scatterCount = 1
		scatterLayout = []int{}
		for _, target := range r.step.Scatter {
			length := len(scatterValues[target])
			scatterCount *= length
			scatterLayout = append(scatterLayout, length)
		}
		scatterInputs = r.generateCrossProductInputs(scatterCount, scatterValues)
		break
	case cwl.FLAT_CROSSPRODUCT:
		scatterCount = 1
		for _, target := range r.step.Scatter {
			length := len(scatterValues[target])
			scatterCount *= length
		}
		scatterLayout = []int{scatterCount}
		scatterInputs = r.generateCrossProductInputs(scatterCount, scatterValues)
		break
	case cwl.DOTPRODUCT:
		fallthrough
	default:
		// 标准中似乎没有指定默认行为，这里我们使用点乘作为默认行为
		// 这是原有的代码，它恰好符合点乘的行为
		scatterCount = -1
		for _, values := range scatterValues {
			if scatterCount == -1 {
				scatterCount = len(values)
			} else if scatterCount != len(values) {
				return -1, nil, nil, errors.New("不一致的Scatter数量")
			}
		}
		scatterLayout = []int{scatterCount}
		scatterInputs = r.generateDotProductInputs(scatterCount, scatterValues)
	}
	return scatterCount, scatterInputs, scatterLayout, nil
}

// generateCrossProductInputs 产生点乘的分发输入
func (r *RegularRunner) generateCrossProductInputs(total int, sources map[string][]cwl.Value) []cwl.Values {
	// 0. 防止为0
	if total == 0 {
		return []cwl.Values{}
	}
	// 1. 产生一个基础，具有所有不需要分发的量
	var initInput = cwl.Values{}
	for _, in := range r.step.In {
		initInput[in.ID] = (*r.parameter)[in.Source[0]]
	}
	// 2. 由这个量初始化一个数组
	var inputs = []cwl.Values{initInput}
	// 3. 遍历所有需要Scatter的量，每次都修改数组
	for _, target := range r.step.Scatter {
		var newInputs []cwl.Values
		for _, input := range inputs {
			for _, src := range sources[target] {
				input[target] = src
				copyInput := cwl.Values{}
				for k, v := range input {
					copyInput[k] = v
				}
				newInputs = append(newInputs, copyInput)
			}
		}
		inputs = newInputs
	}
	// 可以返回
	return inputs
}

// generateDotProduceInputs 产生点乘的分发输入
func (r *RegularRunner) generateDotProductInputs(total int, sources map[string][]cwl.Value) []cwl.Values {
	// 0. 防止为0
	if total == 0 {
		return []cwl.Values{}
	}
	// 1. 初始化空间
	retInputs := make([]cwl.Values, total)
	// 2. 针对每一个产生的输入的每一个In
	for i := 0; i < total; i++ {
		// 2.0 先初始化这个map
		retInputs[i] = cwl.Values{}
		for _, in := range r.step.In {
			if valueArr, ok := sources[in.ID]; ok {
				// 2.1 如果在sources里能找到，分发
				retInputs[i][in.ID] = valueArr[i]
			} else {
				// 2.2 否则原样加入
				retInputs[i][in.ID] = (*r.parameter)[in.Source[0]]
			}
		}
	}
	// 返回
	return retInputs
}

// reconstructOutput 重新结构化输出
func reconstructOutput(values cwl.Values, layout []int) (cwl.Values, error) {
	if len(layout) <= 0 {
		return nil, errors.New("无效的输出布局")
	}
	if len(layout) == 1 { //不需要整理
		return values, nil
	}
	for key, value := range values {
		if valueArr, ok := value.([]cwl.Value); ok { // 如果不是数组的话就不用整理了
			// 需要先验证大小符合
			tmp := 1
			for _, layoutI := range layout {
				tmp *= layoutI
			}
			if tmp != len(valueArr) {
				return nil, fmt.Errorf("输出布局与实际输出不匹配")
			}
			values[key] = reshapeRecursive(layout, valueArr)
		}
	}
	return values, nil
}

// reshapeRecursive 结构化输出的递归函数
func reshapeRecursive(layout []int, flatArr []cwl.Value) []cwl.Value {
	// 最后一层，返回
	if len(layout) == 1 {
		return flatArr
	}
	// 否则，分配空间
	thisLayer := layout[0]
	ret := make([]cwl.Value, thisLayer)
	nextLayerLen := len(flatArr) / thisLayer
	// 然后逐一递归
	for i := 0; i < thisLayer; i++ {
		ret[i] = reshapeRecursive(layout[1:], flatArr[i*nextLayerLen:(i+1)*nextLayerLen])
	}
	return ret
}
