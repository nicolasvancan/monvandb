package dbvengine

import (
	"github.com/nicolasvancan/monvandb/src/database"
)

const (
	// NodeStateNotExecuted represents a node that has not been executed
	NodeStateIdle = iota
	// NodeStateExecuting represents a node that is currently
	NodeStateExecuting
	// NodeStateExecuted represents a node that has been executed
	NodeStateExecuted
	// Node state error
	NodeStateError
)

type ExecutionPlan struct {
	Database         *database.Database        // Pointer to database access
	ExecutionNodes   map[string]*ExecutionNode // List of nodes to execute
	ExecutionContext *ExecutionContext
}

// CreatePlan creates an execution plan from a list of query commands
// Comming from query parser
func (ep *ExecutionPlan) CreatePlan(queryCommands []any) {

}

type ExecutionNode struct {
	DependsOn []string // Depends on hashed OperationResults
	NodeState int
}

// Interface for execution
type Execution interface {
}
