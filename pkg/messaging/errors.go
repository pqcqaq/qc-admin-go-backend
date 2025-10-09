package messaging

import "fmt"

type MessagingError struct {
	ID  string
	Msg string
}

func (e *MessagingError) Error() string {
	return fmt.Sprintf("Messaging error [%s]: %s", e.ID, e.Msg)
}
func NewError(id, msg string) error {
	return &MessagingError{
		ID:  id,
		Msg: msg,
	}
}

var (
	ErrNotSupported = NewError("ERR_NOT_SUPPORTED", "The requested operation is not supported")
)

func IsNotSupportedError(err error) bool {
	if err == nil {
		return false
	}
	if me, ok := err.(*MessagingError); ok {
		return me.ID == "ERR_NOT_SUPPORTED"
	}
	return false
}
