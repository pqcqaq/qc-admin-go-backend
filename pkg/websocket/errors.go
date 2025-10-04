package websocket

type ErroeCode string

const (
	ErrInternalServer   ErroeCode = "INTERNAL_SERVER_ERROR"
	ErrInvalidAuthToken ErroeCode = "INVALID_AUTH_TOKEN"
	ErrMissingAuthToken ErroeCode = "MISSING_AUTH_TOKEN"
	ErrInvalidMessageId ErroeCode = "INVALID_MESSAGE_ID"
	ErrTokenExpired     ErroeCode = "TOKEN_EXPIRED"
)
