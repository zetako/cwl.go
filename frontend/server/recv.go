package server

import (
	"github.com/lijiang2014/cwl.go"
	"github.com/lijiang2014/cwl.go/message"
	"log"
)

type serverMsgReceiver struct {
}

func (r serverMsgReceiver) SendMsg(msg message.Message) {
	// 0. workflow is special
	if msg.Class == message.WorkflowMsg {
		if msg.Status == message.StatusInit {
			parent, err := globalCwlServer.status.GetByID(msg.ID)
			if err != nil {
				log.Println("Invalid Path ID", msg.ID.Path())
				r.unknownMsg(msg)
			}
			childArr, ok := msg.Content.([]string)
			if !ok {
				log.Println("Message Content Type Error")
				r.unknownMsg(msg)
			}
			for _, child := range childArr {
				globalCwlServer.status.Append(parent.GenChild(child))
			}
			return
		}
	}
	// 1. get target
	var (
		parent, target *message.StepStatus
		err            error
	)
	parent, err = globalCwlServer.status.GetByID(msg.ID)
	if err != nil {
		log.Println("Invalid Path ID", msg.ID.Path())
		r.unknownMsg(msg)
	}
	if msg.Class == message.StepMsg {
		target = parent
	} else {
		target, err = parent.Child.Get(msg.Index)
		if err != nil {
			log.Println("Invalid Index", msg.Index)
			r.unknownMsg(msg)
		}
	}
	// 2. set status
	switch msg.Status {
	case message.StatusAssign:
		job, ok := msg.Content.(string)
		if !ok {
			log.Println("Invalid Index", msg.Index)
			r.unknownMsg(msg)
		}
		target.JobID = job
	case message.StatusStart, message.StatusSkip:
		target.Status = msg.Status
	case message.StatusFinish, message.StatusScatter:
		out, ok := msg.Content.(cwl.Values)
		if !ok {
			log.Println("Message Content Type Error")
			r.unknownMsg(msg)
		}
		target.Status = msg.Status
		target.Output = out
	case message.StatusError:
		err, ok := msg.Content.(error)
		if !ok {
			log.Println("Message Content Type Error")
			r.unknownMsg(msg)
		}
		target.Status = msg.Status
		target.Error = err
	default:
		log.Println("Unknown Status", msg.Status)
		r.unknownMsg(msg)
	}
}

func (r serverMsgReceiver) unknownMsg(msg message.Message) {
	log.Println(msg.ToLog())
}

func init() {
	globalCwlServer.status.Append(&message.StepStatus{
		ID:     message.PathID{"root"},
		Status: message.StatusStart,
	})
}
