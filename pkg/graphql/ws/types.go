package ws

import ()

type MessageType string

const (
	GQL_CONNECTION_INIT       = MessageType("connection_init")
	GQL_CONNECTION_TERMINATE  = MessageType("connection_terminate")
	GQL_CONNECTION_ERROR      = MessageType("connection_error")
	GQL_CONNECTION_ACK        = MessageType("connection_ack")
	GQL_CONNECTION_KEEP_ALIVE = MessageType("connection_keep_alive")

	GQL_START    = MessageType("start")
	GQL_STOP     = MessageType("stop")
	GQL_DATA     = MessageType("data")
	GQL_ERROR    = MessageType("error")
	GQL_COMPLETE = MessageType("complete")
)

type MessageID string

type OperationMessage struct {
	Type MessageType `json:"type"`
	ID   *MessageID  `json:"id,omitempty"`

	Payload interface{} `json:"payload,omitempty"`
}

type ConnectionParams map[string]interface{}

type OperationParams struct {
	Query         string
	OperationName string
	Variables     map[string]interface{}
}

type OperationResult struct {
	Data   interface{} `json:"data"`
	Errors []string    `json:"errors"`
}

type MessageReader interface {
	ReadMessage() (int, []byte, error)
}

type MessageWriter interface {
	WriteMessage(int, []byte) error
}

type MessageReaderWriter interface {
	MessageReader
	MessageWriter
}
