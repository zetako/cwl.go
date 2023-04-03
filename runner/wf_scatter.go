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
	)
	totalTask, allInputs, err = r.getAllScatterInputs()
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
				output[key] = []cwl.Value{}
				output[key] = make([]cwl.Value, totalTask)
			}
			output[key].([]cwl.Value)[doneCond.scatterID] = value
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

func (r *RegularRunner) getAllScatterInputs() (int, []cwl.Values, error) {
	var (
		scatterCount   int
		scatterTargets []string
		scatterSinks   = map[string]cwl.Sink{}
		scatterSources = map[string]cwl.ArrayString{}
		scatterValues  = map[string][]cwl.Value{}
		scatterInputs  []cwl.Values
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
				return -1, nil, errors.New("没有匹配的输入")
			}
			if tmpList, ok := tmp.([]cwl.Value); ok {
				scatterValues[key] = append(scatterValues[key], tmpList...)
			} else { // 仅有一个source，source又不是数组，就没法分发了
				return -1, nil, errors.New("没有需要分发的输入")
			}
		} else { // 有多个source，查看linkMerge方法
			for _, source := range sources {
				tmp, ok := (*r.parameter)[source]
				if !ok {
					return -1, nil, errors.New("没有匹配的输入")
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
	// 1.3 计算出总scatter量
	//   - 需要根据ScatterMethod方法来实现
	scatterCount = -1
	for _, values := range scatterValues {
		if scatterCount == -1 {
			scatterCount = len(values)
		} else if scatterCount != len(values) {
			return -1, nil, errors.New("不一致的Scatter数量")
		}
	}
	// 2. 产生各scatter任务需要的inputs
	scatterInputs = make([]cwl.Values, scatterCount)
	for idx := range scatterInputs {
		scatterInputs[idx] = cwl.Values{}
	}
	for _, in := range r.step.In {
		if r.needScatter(in.ID) { // 是需要分发的变量
			for index, input := range scatterInputs {
				input[in.ID] = scatterValues[in.ID][index]
			}
		} else { // 不需要分发的变量直接拷贝就行了
			// 相似的，目前只考虑只需要一个值的情况
			for _, input := range scatterInputs {
				input[in.ID] = (*r.parameter)[in.Source[0]]
			}
		}
	}
	return scatterCount, scatterInputs, nil
}
