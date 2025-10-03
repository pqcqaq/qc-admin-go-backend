package handlers

import (
	"go-backend/pkg/logging"
	"go-backend/pkg/messaging"
)

func test(message messaging.MessageStruct) error {
	logging.Info("Test handler received message: %v", message)
	return nil
}

func registerTestHandler() {
	messaging.RegisterHandler(messaging.ToUserSocket, test)
}
