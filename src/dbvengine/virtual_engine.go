package dbvengine

import (
	"time"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
	"github.com/nicolasvancan/monvandb/src/dbvengine/contexts"
	df "github.com/nicolasvancan/monvandb/src/dbvengine/dataframe"
	executor "github.com/nicolasvancan/monvandb/src/dbvengine/executor"
	parser "github.com/nicolasvancan/monvandb/src/dbvengine/parser"
	monvan_parser "github.com/nicolasvancan/monvandb/src/dbvengine/parser/custom_parser"
	builder "github.com/nicolasvancan/monvandb/src/dbvengine/planner"
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

	return ve
}

func (ve *VirtualEngine) Execute(databaseName string, query string) ExecutionResults {
	results := ExecutionResults{ExecutionFailed, df.Dataframe{}, 0, ""}
	// Parse And analyze Query the query
	var parsedQuery interface{} = nil

	parsedQuery, shouldPass, err := monvan_parser.Parse(query)

	if shouldPass {
		parsedQuery, err = sqlparser.Parse(query)
	}

	if err != nil {
		results.ErrorMessage = err.Error()
		return results
	}
	// Analyze the query
	analyzedQuery := parser.AnalyzeQuery(databaseName, parsedQuery)
	// Execute the query
	if analyzedQuery.Error() != nil {
		results.ErrorMessage = analyzedQuery.Error().Error()
		return results
	}

	// Build the plan
	builder.BuildPlan(analyzedQuery, ve.ExecutionLayer)
	// Get time now
	before := time.Now()
	// Execute the plan
	ve.ExecutionLayer.Start()
	// Get time after
	after := time.Now()
	// Calculate the duration in seconds
	results.Duration = int(after.Sub(before).Seconds())
	// Return the results
	results.Status = ExecutionSuccess
	results.Result = ve.ExecutionLayer.Context.Result
	return results
}

type Context struct {
	// The current database
	Ttl              int
	ExecutionResults map[string]df.Dataframe
}

type ExecutionStatus int

const (
	ExecutionFailed ExecutionStatus = iota
	ExecutionSuccess
)

type ExecutionResults struct {
	Status       ExecutionStatus
	Result       df.Dataframe
	Duration     int
	ErrorMessage string
}
