package dbvengine

/* The engine is the entrypoint for queries to the database.
It is responsible for parsing the query, optmize parsed query, generate execution commands end send them to
be executed, returning */

/*
The virtual engine is the entrypoint for queries to the database. It holds context information
*/
type VirtualEngine struct {
	GlobalContext *Context
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
