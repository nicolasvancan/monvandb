package executor

import (
	"fmt"
	"sync"
)

type ExecutionNodeState string

const (
	NodeIdle     ExecutionNodeState = "IDLE"
	NodeRunning  ExecutionNodeState = "RUNNING"
	NodeFinished ExecutionNodeState = "FINISHED"
	NodeError    ExecutionNodeState = "ERROR"
)

type ExecutionNode struct {
	Id        string
	Final     bool
	Operation Operation
	Layer     *ExecutionLayer
	DependsOn []string
	Notifies  []string
	State     ExecutionNodeState
}

type OpNotification struct {
	NodeId string
	Err    error
}

func NewExecutionNode(id string, executionLayer *ExecutionLayer) *ExecutionNode {
	return &ExecutionNode{
		Id:        id,
		Final:     false,
		Layer:     executionLayer,
		DependsOn: make([]string, 0),
		Notifies:  make([]string, 0),
		State:     NodeIdle,
	}
}

func (en ExecutionNode) String() string {
	return fmt.Sprintf("Node: %s, DependsOn: %v, Notifies: %v, State: %s, Final: %v", en.Id, en.DependsOn, en.Notifies, en.State, en.Final)
}

func (en *ExecutionNode) AddDependency(nodeId string) {
	en.DependsOn = append(en.DependsOn, nodeId)
}

func (en *ExecutionNode) AddNotify(nodeId string) {
	en.Notifies = append(en.Notifies, nodeId)
}

func (en *ExecutionNode) OnNotified(nodeId string) {
	for i, n := range en.DependsOn {
		if n == nodeId {
			en.DependsOn = removeNodeFromSlice(en.DependsOn, i)
			en.Notifies = append(en.Notifies, nodeId)
		}
	}
}

func (en *ExecutionNode) Run(wg *sync.WaitGroup, notify chan<- OpNotification) {

	en.State = NodeRunning
	fmt.Printf("Node %s is Running\n", en.Id)

	// Get results from dependencies
	dependencies := make([]interface{}, len(en.DependsOn))

	// Get operation Results
	for _, dep := range en.DependsOn {
		dependencies = append(dependencies, en.Layer.Context.GetOperationResult(dep))
	}

	// Run the operation Inserting both dataframe dependencies first
	df, err := Operations[en.Operation.Name](append(dependencies, en.Operation.Args...)...)

	if err != nil {
		en.State = NodeError
		notify <- OpNotification{NodeId: en.Id, Err: err}
		return
	}

	// Add result to context
	en.Layer.Context.AddOperationResult(en.Id, df)

	// if it is final, add it to the final result of context
	if en.Final {
		en.Layer.Context.Result = df
	}

	// Remove its id from dependencies of context
	en.Layer.Context.RemoveDependency(en.Id)
	en.State = NodeFinished
	notify <- OpNotification{NodeId: en.Id, Err: nil}
}

// removeNodeFromSlice removes the element at index i from the slice.
func removeNodeFromSlice(slice []string, i int) []string {
	if i < 0 || i >= len(slice) {
		return slice // Return the original slice if the index is out of range
	}
	return append(slice[:i], slice[i+1:]...)
}
