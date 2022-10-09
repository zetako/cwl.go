package brunner

import "strings"

// runtimeToVM loads tool.InputsVM with runtime context
func (tool *Tool) runtimeToVM() (err error) {
	tool.Task.infof("begin load runtime to js vm")
	prefix := tool.Task.Root.ID + "/" // need to trim this from all the input.ID's
	tool.InputsVM = tool.JSVM.Copy()
	context := make(map[string]interface{})
	var f interface{}
	for _, input := range tool.Task.Root.Inputs {
		inputID := strings.TrimPrefix(input.ID, prefix)

		// fixme: handle array of files
		// note: this code block is extraordinarily janky and needs to be refactored
		// error here.
		switch {
		case input.Provided == nil:
			context[inputID] = nil
		case input.Types[0].Type == CWLFileType:
			if input.Provided.Entry != nil {
				// no valueFrom specified in inputBinding
				if input.Provided.Entry.Location != "" {
					f = fileObject(input.Provided.Entry.Location)
				}
			} else {
				// valueFrom specified in inputBinding - resulting value stored in input.Provided.Raw
				tool.Task.infof("input: %v; input provided raw: %v", input.ID, input.Provided.Raw)
				switch input.Provided.Raw.(type) {
				case string:
					f = fileObject(input.Provided.Raw.(string))
				case *File, []*File:
					f = input.Provided.Raw
				default:
					return tool.Task.errorf("unexpected datatype representing file object in input.Provided.Raw")
				}
			}
			fileContext, err := preProcessContext(f)
			if err != nil {
				return tool.Task.errorf("failed to preprocess file context: %v; error: %v", f, err)
			}
			context[inputID] = fileContext
		default:
			context[inputID] = input.Provided.Raw // not sure if this will work in general - so far, so good though - need to test further
		}
	}
	if err = tool.JSVM.Set("runtime", context); err != nil {
		return tool.Task.errorf("failed to set inputs context in js vm: %v", err)
	}
	tool.Task.infof("end load inputs to js vm")
	return nil
}
