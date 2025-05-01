package system

type RequestMessage string
type ResponseMessage string

const (
	// Request types
	AUTHENTICATE RequestMessage = "authenticate"
	QUERY        RequestMessage = "query"
	CLOSE_CONN   RequestMessage = "close_conn"
)

const (
	// Response types
	ERROR   ResponseMessage = "error"
	SUCCESS ResponseMessage = "success"
)

type UserMessage struct {
	RequestType string      `json:"type"`
	Message     interface{} `json:"message"`
}

type ErrorMessage struct {
	Message string `json:"message"`
}

type SuccessMessage struct {
	Result map[string]interface{}
}

type AuthenticateMessage struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type QueryMessage string
