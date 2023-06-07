package status

import (
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"github.com/lijiang2014/cwl.go/runner/message"
	"sort"
	"sync"
)

var (
	IdxExceedError   error = fmt.Errorf("index exceed array's length")
	KeyNotFoundError error = fmt.Errorf("cannot found target with key")
)

type StepStatus struct {
	ID     message.PathID
	Status message.MessageStatus
	JobID  string
	Output cwl.Values
	Error  error
	Child  StepStatusArray
}

func (s *StepStatus) GenChild(child string) *StepStatus {
	tmp := StepStatus{
		ID:     s.ID.ChildPathID(child),
		Status: message.StatusInit,
	}
	s.Child.Append(&tmp)
	return &tmp
}

// StepStatusArray is a thread-safe array of StepStatus
type StepStatusArray struct {
	array []*StepStatus
	sync.RWMutex
}

// Append will add an element to array
func (a *StepStatusArray) Append(status *StepStatus) {
	a.Lock()
	defer a.Unlock()
	if a.array == nil {
		a.array = []*StepStatus{}
	}
	a.array = append(a.array, status)
}

// Get returns an element of specified index
func (a *StepStatusArray) Get(idx int) (*StepStatus, error) {
	a.RLock()
	defer a.RUnlock()
	if idx >= len(a.array) {
		return nil, IdxExceedError
	}
	return a.array[idx], nil
}

// GetByID returns an element of specified runner.PathID; it's O(n) method
func (a *StepStatusArray) GetByID(ID message.PathID) (*StepStatus, error) {
	a.RLock()
	defer a.RUnlock()
	for _, s := range a.array {
		if s.ID.Equal(ID) {
			return s, nil
		}
	}
	return nil, KeyNotFoundError
}

// Foreach just apply a function to each element
func (a *StepStatusArray) Foreach(f func(status *StepStatus)) {
	a.RLock()
	defer a.RUnlock()
	for _, s := range a.array {
		f(s)
	}
}

// ForeachE just apply a function with an error return to each element, stop when err != nil
func (a *StepStatusArray) ForeachE(f func(status *StepStatus) error) error {
	a.RLock()
	defer a.RUnlock()
	for _, s := range a.array {
		err := f(s)
		if err != nil {
			return err
		}
	}
	return nil
}

// SortByPath will sort all elements by their tiers
func (a *StepStatusArray) SortByPath() {
	a.Lock()
	defer a.Unlock()
	sort.Slice(a.array, func(i, j int) bool {
		return len(a.array[i].ID) < len(a.array[j].ID)
	})
}

// GenerateTree can implement tree struct to plain array
func (a *StepStatusArray) GenerateTree() {
	// 1. Sort
	a.SortByPath()
	// 2. Lock
	a.Lock()
	defer a.Unlock()
	// 3. In loop, generate
	for _, ss := range a.array {
		if len(ss.ID) <= 1 { // No parent
			continue
		}
		parentID := ss.ID[:len(ss.ID)-1]

		for _, parent := range a.array {
			if parent.ID.Equal(parentID) { // Find parent
				parent.Child.Append(ss)
				break
			}
		}
	}
}

// GetTier return a new sub-array only contains son of specified base path
func (a *StepStatusArray) GetTier(base message.PathID) *StepStatusArray {
	ret := StepStatusArray{}
	a.Foreach(func(s *StepStatus) {
		if s.ID.IsSonOf(base) {
			ret.Append(s)
		}
	})
	return &ret
}
