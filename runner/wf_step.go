package runner

import (
	"errors"
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"github.com/robertkrimen/otto"
	"time"
)

// StepRunner 对应CWL的一个任务，即
//   - CommandLineTool
//   - Expression
//   - Workflow
type StepRunner interface {
	MeetConditions(now []Condition) bool
	Run(chan<- Condition) error
	RunAtMeetConditions(now []Condition, channel chan<- Condition) (run bool)
	SendCtrlSignal(signal Signal)
}

// RegularRunner 常规Step的执行器
//   - 用来执行CommandLineTool和Expression
//   - 原理上，应该也可以作为Workflow的执行器
type RegularRunner struct {
	parent             *WorkflowRunner
	neededCondition    []Condition
	engine             *Engine
	step               *cwl.WorkflowStep
	process            *Process
	parameter          *cwl.Values
	useWorkflowDefault map[string]bool
}

func (r *RegularRunner) MeetConditions(now []Condition) bool {
	// 比较条件集
	for _, need := range r.neededCondition {
		inputCond, ok := need.(InputParamCondition)
		if !ok {
			if !need.Meet(now) {
				return false
			}
		} else {
			meet, useDefault := inputCond.MeetOrDefault(now)
			if !meet {
				return false
			}
			if useDefault {
				r.useWorkflowDefault[inputCond.input.ID] = true
			}
		}
	}
	return true
}

func (r *RegularRunner) Run(conditions chan<- Condition) (err error) {
	//log.Printf("[Step \"%s\"] Start", r.step.ID)
	r.engine.SendMsg(Message{
		Class:     StepMsg,
		Status:    StatusStart,
		TimeStamp: time.Now(),
		ID:        r.step.ID,
	})
	var (
		passBoolean bool
		doScattered bool
		ok          bool
		outs        cwl.Values
	)
	defer func() { // 负责处理运行结束的情况
		if err != nil {
			//log.Printf("[Step \"%s\"] Error:%s", r.step.ID, err)
			r.engine.SendMsg(Message{
				Class:     StepMsg,
				Status:    StatusError,
				TimeStamp: time.Now(),
				ID:        r.step.ID,
				Content:   err,
			})
			conditions <- &StepErrorCondition{
				step: r.step,
				err:  err,
			}
		} else if doScattered {
			// 相关日志由scatter内处理
		} else if !passBoolean {
			//log.Printf("[Step \"%s\"] Skip", r.step.ID)
			r.engine.SendMsg(Message{
				Class:     StepMsg,
				Status:    StatusSkip,
				TimeStamp: time.Now(),
				ID:        r.step.ID,
			})
			for _, output := range r.step.Out {
				conditions <- &OutputParamCondition{
					step:   r.step,
					output: &output,
				}
			}
			conditions <- &StepDoneCondition{
				step:    r.step,
				out:     nil,
				runtime: r.process.runtime,
			}
		} else {
			//log.Printf("[Step \"%s\"] Finish", r.step.ID)
			r.engine.SendMsg(Message{
				Class:     StepMsg,
				Status:    StatusFinish,
				TimeStamp: time.Now(),
				ID:        r.step.ID,
				Content:   outs,
			})
			for _, output := range r.step.Out {
				conditions <- &OutputParamCondition{
					step:   r.step,
					output: &output,
				}
			}
			conditions <- &StepDoneCondition{
				step:    r.step,
				out:     &outs,
				runtime: r.process.runtime,
			}
		}
	}()
	// 1. 如果需要Scatter，任务交由RunScatter
	if r.step.Scatter != nil && len(r.step.Scatter) > 0 {
		if r.step.While != "" {
			return fmt.Errorf("不允许同时使用Scatter和While")
		}
		doScattered = true
		return r.RunScatter(conditions)
	}
	// 2. 处理输入
	for _, in := range r.step.In {
		if in.ValueFrom != "" {
			continue
		}
		if _, ok := r.useWorkflowDefault[in.ID]; ok {
			(*r.process.inputs)[in.ID] = in.Default
		} else if len(in.Source) == 1 {
			tmp, ok := (*r.parameter)[in.Source[0]]
			if !ok || tmp == nil {
				tmp = in.Default
			}
			(*r.process.inputs)[in.ID] = tmp
		} else {
			switch in.LinkMerge {
			case cwl.MERGE_FLATTENED:
				var tmpInput []cwl.Value
				for _, src := range in.Source {
					tmp, ok := (*r.parameter)[src]
					if !ok {
						return errors.New("没有匹配的输入")
					}
					if tmpList, ok := tmp.([]cwl.Value); ok {
						tmpInput = append(tmpInput, tmpList...)
					} else { // 也有可能是单个元素
						tmpInput = append(tmpInput, tmp)
					}
				}
				(*r.process.inputs)[in.ID] = tmpInput
				break
			case cwl.MERGE_NESTED:
				fallthrough
			default:
				(*r.process.inputs)[in.ID] = (*r.parameter)[in.Source[0]]
				break
			}
		}
	}
	// 处理PickValue
	for _, in := range r.step.In {
		if in.PickValue != nil {
			value := (*r.process.inputs)[in.ID]
			value, err = pickValue(value, *in.PickValue)
			if err != nil {
				return fmt.Errorf("PickValue计算失败: %s", err)
			}
			(*r.process.inputs)[in.ID] = value
		}
	}

	// 处理ValueFrom
	err = preprocessInputs(r.process.inputs)
	if err != nil {
		return fmt.Errorf("预处理inputs失败: %v\n", err)
	}
	err = r.process.jsvm.setInputs(*r.process.inputs)
	if err != nil {
		return err
	}
	for _, in := range r.step.In {
		if in.ValueFrom == "" {
			continue
		}
		if len(in.Source) == 1 {
			rawValue, err := r.process.jsvm.Eval(in.ValueFrom, (*r.parameter)[in.Source[0]])
			if err != nil {
				return fmt.Errorf("ValueFrom计算失败: %v", err)
			}
			(*r.process.inputs)[in.ID], err = cwl.ConvertToValue(rawValue)
			if err != nil {
				return fmt.Errorf("转换ValueFrom结果失败: %v", err)
			}
		} else if r.parent.workflow.RequiresMultipleInputFeature() {
			// 这部分可能依旧需要参考LinkMerge方法
			var tmpPlainArr []interface{}
			for _, src := range in.Source {
				tmpValue := (*r.parameter)[src]
				plainValue, err := toJSONMap(tmpValue)
				if err != nil {
					return err
				}
				if plainValue == nil {
					plainValue = otto.NullValue()
				}
				tmpPlainArr = append(tmpPlainArr, plainValue)
			}
			// 计算
			result, err := r.process.jsvm.Eval(in.ValueFrom, tmpPlainArr)
			if err != nil {
				return err
			}
			// 转换输出
			resultValue, err := cwl.ConvertToValue(result)
			if err != nil {
				return err
			}
			// 保存结果
			(*r.process.inputs)[in.ID] = resultValue
		} else { // 要不然就是根据已有值计算
			// 计算
			result, err := r.process.jsvm.Eval(in.ValueFrom, nil)
			if err != nil {
				return err
			}
			// 转换输出
			resultValue, err := cwl.ConvertToValue(result)
			if err != nil {
				return err
			}
			// 保存结果
			(*r.process.inputs)[in.ID] = resultValue
		}
	}
	// 如果需要计算When，计算
	if r.step.When != "" {
		err = r.process.RefreshVMInputs()
		if err != nil {
			return err
		}
		pass, err := r.process.Eval(r.step.When, nil)
		if err != nil {
			return err
		}
		if passBoolean, ok = pass.(bool); !ok {
			return fmt.Errorf("when表达式未输出布尔值")
		} else {
			if !passBoolean {
				// 5. 返回
				return nil
			} // 通过就走正常流程
		}
	} else {
		passBoolean = true
	}
	// 3. 然后使用 Engine.RunProcess()
	if r.step.While != "" {
		outs, err = r.RunLoop()
	} else {
		outs, err = r.engine.RunProcess(r.process)
	}
	if err != nil {
		return err
	}
	// 4. 最后，根据Step的输出，释放Condition
	// 移动到defer处理
	// 5. 返回
	return nil
}

func (r *RegularRunner) RunAtMeetConditions(now []Condition, channel chan<- Condition) (run bool) {
	if r.MeetConditions(now) {
		go func() {
			_ = r.Run(channel)
		}()
		return true
	}
	return false
}

func (r *RegularRunner) SendCtrlSignal(signal Signal) {
	if r.process.signalChannel != nil {
		go func() { r.process.signalChannel <- signal }()
	}
}

func NewStepRunner(e *Engine, parent *WorkflowRunner, step *cwl.WorkflowStep) (StepRunner, error) {
	var (
		err error
	)
	// 目前返回的均为RegularRunner，未来可能考虑返回WorkflowRunner
	ret := RegularRunner{
		neededCondition:    []Condition{},
		engine:             e,
		step:               step,
		process:            nil,
		parameter:          parent.parameter,
		useWorkflowDefault: map[string]bool{},
		parent:             parent,
	}
	// 生成条件集
	for _, input := range step.In {
		ret.neededCondition = append(ret.neededCondition, InputParamCondition{
			step:  ret.step,
			input: input,
		})
	}
	// 创建process
	ret.process, err = ret.engine.GenerateSubProcess(ret.step)
	if err != nil {
		return nil, err
	}
	ret.process.msgTemplate = Message{
		Class: StepMsg,
		ID:    ret.step.ID,
	}
	// 继承父运行时
	ret.process.parentRuntime = e.process.runtime
	// 从engine处继承requirement和hints
	// 准确的说，应该从他的父步骤继承
	// engine < step < Process
	ret.process.root.Process.Base().InheritRequirement(step.Requirements, step.Hints)
	ret.process.root.Process.Base().InheritRequirement(parent.workflow.Requirements, parent.workflow.Hints)
	// 继承ValueFrom
	// 返回
	return &ret, nil
}
