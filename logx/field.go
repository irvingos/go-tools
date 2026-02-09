package logx

// alias
type Field = string

const (
	// base
	FieldTraceID  Field = "trace_id"
	FieldUsername Field = "username"
	FieldCaller   Field = "caller"

	// http
	FieldLatency  Field = "latency"
	FieldStatus   Field = "status"
	FieldRemoteIP Field = "remote_ip"
	FieldMethod   Field = "method"
	FieldPath     Field = "path"
	FieldQuery    Field = "query"
	FieldUA       Field = "ua"
	FieldHeaders  Field = "headers"
	FieldCode     Field = "code"

	// recovery
	FieldEvent    Field = "event"
	FieldFile     Field = "file"
	FieldFunction Field = "function"
	FieldRecover  Field = "recover"
	FieldStack    Field = "stack"
)
