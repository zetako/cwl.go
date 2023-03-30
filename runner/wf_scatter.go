package runner

import (
	"errors"
	"github.com/lijiang2014/cwl.go"
)

// RunScatter 运行需要分发任务的步骤
func (r *RegularRunner) RunScatter(condition chan<- Condition) (err error) {
	var (
		totalTask    int
		runningTask  int                    = 0
		internalChan chan Condition         = make(chan Condition)
		process      *Process               = nil
		isSuccess    bool                   = true // 标记所有任务是否成功
		allOutputs   []ScatterDoneCondition        // 用以存储每一步的输出
		output       cwl.Values
	)
	totalTask = len((*r.parameter)[r.step.Scatter[0]].([]cwl.Value))
	for i := 0; i < totalTask; i++ {
		// 1. Scatter的每个任务都需要创建Process
		process, err = r.engine.GenerateSubProcess(r.step)
		if err != nil {
			isSuccess = false
			break
		}
		// 2. 分别绑定输入
		process.inputs = &cwl.Values{}
		for _, in := range r.step.In {
			if r.needScatter(in.ID) { // 是需要分发的变量
				(*process.inputs)[in.ID] = (*r.parameter)[in.Source[0]].([]cwl.Value)[i]
			} else { // 不需要分发的变量直接拷贝就行了
				// 相似的，目前只考虑只需要一个值的情况
				(*process.inputs)[in.ID] = (*r.parameter)[in.Source[0]]
			}
		}

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
			}
			output[key] = append(output[key].([]cwl.Value), value)
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
	for _, entry := range r.step.Scatter {
		if entry == key {
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
