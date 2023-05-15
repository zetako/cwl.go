package runner

import (
	"errors"
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"path"
	"time"
)

// RunScatter 运行需要分发任务的步骤
func (r *RegularRunner) RunScatter(condition chan<- Condition) (err error) {

	var (
		totalTask    int
		runningTask  int            = 0
		internalChan chan Condition = make(chan Condition)
		process      *Process       = nil
		genSuccess   bool           = true // 标记所有任务是否成功
		allInputs    []cwl.Values
		allOutputs   []ScatterDoneCondition // 用以存储每一步的输出
		output       cwl.Values
		layout       []int
	)
	defer func() {
		if err != nil {
			//log.Printf("[Step \"%s\"] Error:%s", r.step.ID, err)
			r.engine.SendMsg(Message{
				Class:     StepMsg,
				Status:    StatusError,
				TimeStamp: time.Now(),
				ID:        r.step.ID,
				Error:     err,
			})
			condition <- &StepErrorCondition{
				step: r.step,
				err:  err,
			}
		} else {
			//log.Printf("[Step \"%s\"] Scattered", r.step.ID)
			r.engine.SendMsg(Message{
				Class:     StepMsg,
				Status:    StatusScatter,
				TimeStamp: time.Now(),
				ID:        r.step.ID,
			})
			condition <- &StepDoneCondition{
				step:    r.step,
				out:     &output,
				runtime: r.process.runtime,
			}
		}
	}()
	totalTask, allInputs, layout, err = r.getAllScatterInputs()
	if err != nil {
		return err
	}
	defer func() { // 定义任务启动失败的处理
		if !genSuccess { // 任务启动失败，等待收回所有任务
			for runningTask > 0 {
				tmpCondition := <-internalChan
				if _, ok := tmpCondition.(*ScatterDoneCondition); ok { // 该步正常结束
					runningTask--
				} else if _, ok = tmpCondition.(*ScatterErrorCondition); ok { // 该步异常结束
					runningTask--
				}
			}
		}
	}()
	for i := 0; i < totalTask; i++ {
		// 1. Scatter的每个任务都需要创建Process
		process, err = r.engine.GenerateSubProcess(r.step)
		if err != nil {
			genSuccess = false
			return fmt.Errorf("创建Process实例失败")
		}
		process.msgTemplate = Message{
			Class: ScatterMsg,
			ID:    r.step.ID,
			Index: i,
		}
		// 2. 设置输出 & 绑定输入
		// 2.1 一般输入
		process.runtime.RootHost = path.Join(process.runtime.RootHost, fmt.Sprintf("scatter%d", i))
		process.outputFS = &Local{
			workdir:      process.runtime.RootHost,
			CalcChecksum: true,
		}
		process.loadRuntime()
		process.inputs = &allInputs[i]
		// 2.2 ValueFrom输入
		err = preprocessInputs(r.process.inputs)
		if err != nil {
			genSuccess = false
			return fmt.Errorf("预处理inputs失败: %v\n", err)
		}
		err = process.jsvm.setInputs(*process.inputs)
		if err != nil {
			genSuccess = false
			return fmt.Errorf("设置inputs失败: %v\n", err)
		}
		for _, in := range r.step.In {
			//if in.ValueFrom != "" && !r.needScatter(in.ID) {
			if in.ValueFrom != "" {
				tmp := (*process.inputs)[in.ID]
				tmp, err = process.jsvm.evalValueFrom(in.ValueFrom, tmp)
				if err != nil {
					genSuccess = false
					return fmt.Errorf("ValueFrom计算失败: %v\n", err)
				}
				(*process.inputs)[in.ID] = tmp
			}
		}

		// 3. 并行执行（?）
		go r.scatterTaskWrapper(process, internalChan, i)
		runningTask++

		// 4. 如果需要控制最大并行数，可以先回收部分任务
		maxScatter := r.engine.flags.MaxScatterLimit
		if maxScatter > 0 && runningTask > maxScatter {
			// 暂定为只回收一个，未来可以通过flag控制
			tmpCondition := <-internalChan
			if doneCondition, ok := tmpCondition.(*ScatterDoneCondition); ok {
				allOutputs = append(allOutputs, *doneCondition)
				runningTask--
			} else if errCondition, ok := tmpCondition.(*ScatterErrorCondition); ok { // 该步异常结束
				runningTask--
				genSuccess = false
				return errCondition.err
			}
		}
	}
	// 4. 等待结束
	for runningTask > 0 {
		tmpCondition := <-internalChan
		if doneCondition, ok := tmpCondition.(*ScatterDoneCondition); ok {
			allOutputs = append(allOutputs, *doneCondition)
			runningTask--
		} else if errCondition, ok := tmpCondition.(*ScatterErrorCondition); ok { // 该步异常结束
			runningTask--
			genSuccess = false
			return errCondition.err
		}
	}
	// 5. 整理结果
	output = cwl.Values{}
	if totalTask == 0 {
		// 空，需要产生空的输出
		output = constructZeroOutput(layout, r.step.Out)
	} else {
		// 非空，整理
		for _, doneCond := range allOutputs {
			// 如果是空的，说明When没有通过；需要为它的所有结果设置为nil
			if doneCond.out == nil || *doneCond.out == nil {
				for _, out := range r.process.root.Process.Base().Outputs {
					key := out.GetOutputParameter().ID
					if _, ok := output[key]; !ok {
						//output[key] = []cwl.Value{}
						output[key] = make([]cwl.Value, totalTask)
					}
					output[key].([]cwl.Value)[doneCond.scatterID] = nil
				}
				continue
			}
			// 否则，进行结果合并
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
			return
		}
	}
	return nil
}

// scatterTaskWrapper 已分发任务的运行封装，用于在协程中运行；使用channel传递结果
func (r *RegularRunner) scatterTaskWrapper(p *Process, condChan chan Condition, ID int) {
	//log.Printf("[Step \"%s\": Scatter %d] Start", r.step.ID, ID)
	r.engine.SendMsg(Message{
		Class:     ScatterMsg,
		Status:    StatusStart,
		TimeStamp: time.Time{},
		ID:        r.step.ID,
		Index:     ID,
	})
	var (
		err         error
		out         cwl.Values
		pass        interface{}
		passBoolean bool
		ok          bool
	)
	// 用于捕捉错误的退出
	defer func() {
		if err != nil {
			//log.Printf("[Step \"%s\": Scatter %d] Error:%s", r.step.ID, ID, err)
			r.engine.SendMsg(Message{
				Class:     ScatterMsg,
				Status:    StatusError,
				TimeStamp: time.Now(),
				ID:        r.step.ID,
				Index:     ID,
				Error:     err,
			})
			condChan <- &ScatterErrorCondition{
				scatterID: ID,
				err:       err,
			}
		} else if r.step.When != "" && !passBoolean {
			//log.Printf("[Step \"%s\": Scatter %d] Skip", r.step.ID, ID)
			r.engine.SendMsg(Message{
				Class:     ScatterMsg,
				Status:    StatusSkip,
				TimeStamp: time.Now(),
				ID:        r.step.ID,
				Index:     ID,
			})
			// 没有通过测试，直接输出空结果
			condChan <- &ScatterDoneCondition{
				scatterID: ID,
				out:       nil,
			}
		} else {
			//log.Printf("[Step \"%s\": Scatter %d] Finish", r.step.ID, ID)
			r.engine.SendMsg(Message{
				Class:     ScatterMsg,
				Status:    StatusFinish,
				TimeStamp: time.Now(),
				ID:        r.step.ID,
				Index:     ID,
				Values:    out,
			})
			condChan <- &ScatterDoneCondition{
				scatterID: ID,
				out:       &out,
			}
		}
	}()
	if r.step.When != "" {
		// 有运行条件
		err = p.RefreshVMInputs()
		if err != nil {
			return
		}
		pass, err = p.Eval(r.step.When, nil)
		if err != nil {
			return
		}
		if passBoolean, ok = pass.(bool); !ok {
			err = fmt.Errorf("when表达式未输出布尔值")
		} else {
			if !passBoolean {
				return
			}
		}
	}
	// 没有运行条件或者运行条件通过了
	out, err = r.engine.RunProcess(p)
	if err != nil {
		return
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
		if len(sources) <= 0 {
			// 如果没有来源，就使用默认值
			for _, in := range r.step.In {
				if in.ID == key && in.Default != nil {
					if defaultArr, ok := in.Default.([]cwl.Value); ok {
						scatterValues[key] = append(scatterValues[key], defaultArr...)
					}
				}
			}
			if len(scatterValues[key]) <= 0 {
				return -1, nil, nil, fmt.Errorf("输入%s没有绑定值或默认值", key)
			}
		} else if len(sources) == 1 { // 仅有单一source
			source := sources[0]
			tmp, ok := (*r.parameter)[source]
			if !ok {
				return -1, nil, nil, fmt.Errorf("来源%s没有匹配的输入", source)
			}
			if tmpList, ok := tmp.([]cwl.Value); ok {
				for _, entity := range tmpList {
					scatterValues[key] = append(scatterValues[key], entity)
				}
			} else { // 仅有一个source，source又不是数组，就没法分发了
				return -1, nil, nil, errors.New("没有需要分发的输入")
			}
		} else { // 有多个source，查看linkMerge方法
			for _, source := range sources {
				tmp, ok := (*r.parameter)[source]
				if !ok {
					tmp = nil
					// 多来源可以允许空来源
					//return -1, nil, nil, fmt.Errorf("来源%s没有匹配的输入", source)
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
		if len(scatterValues[key]) == 0 {
			for _, in := range r.step.In {
				if in.ID == key && in.Default != nil {
					if defaultArr, ok := in.Default.([]cwl.Value); ok {
						scatterValues[key] = defaultArr
					}
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
		if in.Source == nil || len(in.Source) == 0 { // 需要考虑取值由之后的ValueFrom计算出来的情况
			initInput[in.ID] = nil
		} else {
			initInput[in.ID] = (*r.parameter)[in.Source[0]]
		}
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
				if in.Source == nil || len(in.Source) == 0 { // 需要考虑取值由之后的ValueFrom计算出来的情况
					retInputs[i][in.ID] = nil
				} else {
					retInputs[i][in.ID] = (*r.parameter)[in.Source[0]]
				}
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

// constructZeroOutput 产生符合layout的零输出
func constructZeroOutput(layout []int, outputs []cwl.WorkflowStepOutput) cwl.Values {
	ret := cwl.Values{}
	zeroArr := constructZeroValueRecursive(layout)
	for _, out := range outputs {
		ret[out.ID] = zeroArr
	}
	return ret
}

// constructZeroValueRecursive 递归产生需要的零输出数组
func constructZeroValueRecursive(layout []int) []cwl.Value {
	if layout[0] == 0 {
		return []cwl.Value{}
	}
	ret := make([]cwl.Value, layout[0])
	for i := 0; i < layout[0]; i++ {
		ret[i] = constructZeroValueRecursive(layout[1:])
	}
	return ret
}
