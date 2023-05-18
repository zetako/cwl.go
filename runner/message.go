package runner

import (
	"encoding/json"
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"time"
)

// MessageClass is class of Message source.
// It can be a step, a scattered or looped sub-step or other similar thing
type MessageClass string

const (
	WorkflowMsg = "workflow"
	StepMsg     = "step"
	ScatterMsg  = "scatter"
	IterMsg     = "iter"
)

// MessageStatus represent status of the Message Source.
type MessageStatus string

const (
	StatusInit    = "Init"      // Content is []string representing all steps' ID
	StatusStart   = "Start"     // Content is nil
	StatusFinish  = "Finish"    // Content is cwl.Values representing output
	StatusAssign  = "Assign"    // Content is string representing JobID
	StatusError   = "Error"     // Content is error
	StatusSkip    = "Skip"      // Content is nil
	StatusScatter = "Scattered" // Content is nil
	StatusLoop    = "Looped"    // Content is nil
)

// Message is a struct to transfer status change in workflow process
type Message struct {
	Class     MessageClass  // Class of source
	Status    MessageStatus // Status of source, usually message will be sent if status is changed
	TimeStamp time.Time     // TimeStamp when this message been sent
	ID        string        // ID is to identified source, usually StepID
	Index     int           // Index can locate exact index when source is scatter or loop
	Content   interface{}   // Content is Message's actual info, the type is determined by Status
	//Info      string        // Info is normal message. e.g. in StatusAssign msg, it contains JobID
	//Error     error         // Error is error message, normally used in StatusError msg
	//Values    cwl.Values    // Values is cwl.Values message, normally used in StatusFinish msg
}

// ToString convert a Message to plain string
func (m Message) ToString() string {
	if m.Content == nil {
		return ""
	}
	if tmp, ok := m.Content.(cwl.Values); ok {
		tmpStr, err := json.MarshalIndent(tmp, "", "  ")
		if err != nil {
			return "Unknown Output"
		} else {
			return "Output: " + string(tmpStr)
		}
	}
	return fmt.Sprintf("%v", m.Content)
}

// ToLog convert a Message to log string
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
	if m.ID != "" {
		return fmt.Sprintf("[%s][Step \"%s\"%s][%s] %s", m.TimeStamp.Format(time.DateTime), m.ID, tmp, m.Status, m.ToString())
	} else if m.Class == WorkflowMsg {
		return fmt.Sprintf("[%s][Workflow Root][%s] %s", m.TimeStamp.Format(time.DateTime), m.Status, m.ToString())
	} else {
		return fmt.Sprintf("[%s][Non-Workflow][%s] %s", m.TimeStamp.Format(time.DateTime), m.Status, m.ToString())

	}
}

type MessageReceiver interface {
	SendMsg(message Message)
}

// DefaultMsgReceiver will print all Message it receive to log
type DefaultMsgReceiver struct {
}

// SendMsg impl the MessageReceiver interface
func (DefaultMsgReceiver) SendMsg(message Message) {
	fmt.Println(message.ToLog())
}
