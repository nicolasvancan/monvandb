package contexts

import (
	df "github.com/nicolasvancan/monvandb/src/dbvengine/dataframe"
)

type ExecutionContext struct {
	OperationsResults map[string]df.Dataframe
	Dependencies      map[string][]string
	Result            df.Dataframe
}

func NewExecutionContext() *ExecutionContext {
	return &ExecutionContext{
		OperationsResults: make(map[string]df.Dataframe),
		Dependencies:      make(map[string][]string),
	}
}

func (ec *ExecutionContext) RemoveDependency(name string) {
	if _, ok := ec.Dependencies[name]; !ok {
		return
	}

	if len(ec.Dependencies[name]) > 0 {
		for _, dep := range ec.Dependencies[name] {
			if dep == name {
				ec.RemoveDependency(dep)
			}
		}
	} else {
		delete(ec.Dependencies, name)
		// Clean also the data in memory
		delete(ec.OperationsResults, name)
	}
}

func (ec *ExecutionContext) AddDependency(dependsOn string, nodeId string) {
	if _, ok := ec.Dependencies[dependsOn]; !ok {
		ec.Dependencies[dependsOn] = []string{}
	}

	ec.Dependencies[dependsOn] = append(ec.Dependencies[dependsOn], nodeId)
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
