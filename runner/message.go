package runner

import (
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"log"
)

type MessageClass string

const (
	StepMsg    = "step"
	ScatterMsg = "scatter"
	IterMsg    = "iter"
)

type MessageStatus string

const (
	StatusStart   = "Start"
	StatusFinish  = "Finish"
	StatusError   = "Error"
	StatusSkip    = "Skip"
	StatusScatter = "Scattered"
	StatusLoop    = "Looped"
)

type Message struct {
	Class  MessageClass
	Status MessageStatus
	ID     string
	Index  int
	Info   string
	Error  error
	Values cwl.Values
}

func (m Message) ToString() string {
	if m.Status == StatusError {
		return m.Error.Error()
	}
	return ""
}

func (m Message) ToLog() string {
	var tmp string
	switch m.Class {
	case StepMsg:
		// DO NOTHING
	case ScatterMsg:
		tmp = fmt.Sprintf(" :Scatter %d", m.Index)
	case IterMsg:
		tmp = fmt.Sprintf(" :Iteration %d", m.Index)
	default:
		// DO NOTHING
	}

	return fmt.Sprintf("[Step %s%s][%s] %s", m.ID, tmp, m.Status, m.ToString())
}

type MessageReceiver interface {
	SendMsg(message Message)
}

// DefaultMsgReceiver will print all Message it receive to log
type DefaultMsgReceiver struct {
}

// SendMsg impl the MessageReceiver interface
func (DefaultMsgReceiver) SendMsg(message Message) {
	log.Println(message.ToLog())
}
