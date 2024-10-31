package dbvengine

import (
	df "github.com/nicolasvancan/monvandb/src/dbvengine/dataframe"
)

type ExecutionContext struct {
	OperationsResults map[string]df.Dataframe
}

func (ec *ExecutionContext) AddOperationResult(name string, df df.Dataframe) {
	ec.OperationsResults[name] = df
}

func (ec *ExecutionContext) GetOperationResult(name string) df.Dataframe {
	return ec.OperationsResults[name]
}

type GlobalQueryContext struct {
	HashedQueryResults map[string]interface{}
}
