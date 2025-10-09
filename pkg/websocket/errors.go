package websocket

type ErroeCode string

const (
	ErrInternalServer       ErroeCode = "INTERNAL_SERVER_ERROR"
	ErrChannelCreateTimeout ErroeCode = "CHANNEL_CREATE_TIMEOUT"
	ErrEmptyTopic           ErroeCode = "EMPTY_TOPIC"
	ErrInvalidAuthToken     ErroeCode = "INVALID_AUTH_TOKEN"
	ErrMissingAuthToken     ErroeCode = "MISSING_AUTH_TOKEN"
	ErrInvalidMessageId     ErroeCode = "INVALID_MESSAGE_ID"
	ErrTokenExpired         ErroeCode = "TOKEN_EXPIRED"
	ErrChannelExists        ErroeCode = "CHANNEL_ALREADY_EXISTS"
)
