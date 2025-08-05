package bridge

// ExecutionContext holds all the context for a workflow execution
type ExecutionContext struct {
	Workspace    *Workspace
	ConfigData   []byte
	Metadata     ExecutionMetadata
	Secrets      map[string]string
	Environment  map[string]string
	EventPayload map[string]interface{}
	DryRun       bool
}

// ExecutionMetadata contains metadata about the execution
type ExecutionMetadata struct {
	Space    string
	Unit     string
	Revision int
	Actor    string
}
