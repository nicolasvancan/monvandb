package dbvengine

import (
	"github.com/nicolasvancan/monvandb/src/dbvengine/contexts"
	executor "github.com/nicolasvancan/monvandb/src/dbvengine/executor"
)

/* The engine is the entrypoint for queries to the database.
It is responsible for parsing the query, optmize parsed query, generate execution commands end send them to
be executed, returning */

/*
The virtual engine is the entrypoint for queries to the database. It holds context information
*/
type VirtualEngine struct {
	GlobalContext  *Context
	ExecutionLayer *executor.ExecutionLayer
}

func NewVirtualEngine(globalContext *Context) *VirtualEngine {
	ve := new(VirtualEngine)
	ve.GlobalContext = globalContext
	// Create execution Layer
	ve.ExecutionLayer = executor.NewExecutionLayer(new(contexts.ExecutionContext))
	// Create the execution layer

	return ve
}

type Context struct {
	// The current database
	Ttl              int
	ExecutionResults map[string]interface{}
}

type ExecutionResults struct {
	Status       int
	Result       interface{}
	ErrorMessage string
}
