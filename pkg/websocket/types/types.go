package types

import "go-backend/pkg/messaging"

type MessageSender func(message messaging.SocketMessagePayload) error
