package subscription

import (
	"go-backend/pkg/logging"
	"go-backend/pkg/utils"
	"sync"
)

type SubscribeSuccessPayload struct {
	Topic     string
	UserID    uint64
	SessionId string
	ClientId  uint64
}

type MessageToClient = map[string]interface{}

type SubscribeSuccessListener func(payload SubscribeSuccessPayload) (MessageToClient, error)

var (
	handlers = make(map[string][]SubscribeSuccessListener)
	rwLock   sync.RWMutex
)

func RegisterSubscribeSuccessListener(topic string, listener SubscribeSuccessListener) {
	rwLock.Lock()
	defer rwLock.Unlock()
	handlers[topic] = append(handlers[topic], listener)
}

// 发布消息
func PublishSubscribeSuccessMessage(topic string, payload SubscribeSuccessPayload) ([]MessageToClient, error) {
	rwLock.RLock()
	defer rwLock.RUnlock()

	logging.Info("Publishing subscribe success message: topic=%s, payload=%v", topic, payload)

	var wg sync.WaitGroup
	results := make([]MessageToClient, 0)

	for subTopic, listeners := range handlers {
		if utils.MatchTopic(subTopic, topic) {
			wg.Add(1)
			for _, listener := range listeners {
				go func(listener SubscribeSuccessListener) {
					defer wg.Done()
					result, err := listener(payload)
					if err == nil {
						results = append(results, result)
					} else {
						logging.Warn("Failed to process message for topic %s: %v", subTopic, err)
					}
				}(listener)
			}
		}
	}

	wg.Wait()
	return results, nil
}
