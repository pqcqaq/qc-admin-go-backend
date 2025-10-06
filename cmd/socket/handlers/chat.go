package handlers

import (
	"context"
	channelhandler "go-backend/pkg/channel_handler"
	"go-backend/pkg/logging"
	"go-backend/pkg/utils"
	"time"
)

func RegisterChat() {
	ctx := context.Background()
	handler := channelhandler.NewChannelHandler(channelhandler.CreateChannelHandlerOptions{
		Topic:       "chat_room/#",
		SendMessage: channelhandler.NewMessageSender(ctx),
		NewChannelReceived: func(channel *channelhandler.IsolateChannel) error {
			// 创建字节队列用于存储消息
			messageQueue := make([]byte, 0)
			ticker := time.NewTicker(5 * time.Second) // 5秒超时检查
			defer ticker.Stop()

			lastMessageTime := time.Now()

			logging.Info("Chat room channel %s established, starting message processing", channel.ID)

			for {
				select {
				case <-ticker.C:

					// 如果消息还没发送完,延续5s
					if len(messageQueue) > 0 {
						lastMessageTime = time.Now()
						continue
					}

					// 检查是否5秒内没有新消息
					if time.Since(lastMessageTime) >= 5*time.Second {
						logging.Info("Chat room channel %s timeout - no new messages for 5 seconds, closing", channel.ID)
						channel.Close()
						return nil
					}

				default:
					// 尝试读取新消息
					msg, err := channel.Read()
					if err != nil {
						// 如果读取出错，检查是否是超时或者channel关闭
						logging.Info("Channel %s read error or closed: %v", channel.ID, err)
						return err
					}

					// 将消息转换为字节并拼接到队列
					if msgBytes, ok := msg.Data.([]byte); ok {
						messageQueue = append(messageQueue, msgBytes...)
						lastMessageTime = time.Now()
					} else if msgStr, ok := msg.Data.(string); ok {
						msgBytes := []byte(msgStr)
						messageQueue = append(messageQueue, msgBytes...)
						lastMessageTime = time.Now()
					} else {
						logging.Warn("Chat room channel %s received unsupported message type", channel.ID)
						continue
					}

					// 按字符边界处理消息队列
					for len(messageQueue) > 0 {
						// 尝试解码UTF-8字符
						validEnd := 0
						for i := 1; i <= len(messageQueue) && i <= 10; i++ { // 最多一次处理10个字节（约3-5个中文字符）
							if utils.IsValidUTF8(messageQueue[:i]) {
								validEnd = i
							}
						}

						if validEnd == 0 {
							// 如果前面的字节都无法组成有效UTF-8，跳过第一个字节
							validEnd = 1
						}

						// 取出有效的字节段
						chunk := messageQueue[:validEnd]
						messageQueue = messageQueue[validEnd:] // 从队列中移除已处理的字节

						// 发送这些字节
						err := channel.Send(utils.ByteToString(chunk))
						if err != nil {
							logging.Error("Chat room channel %s failed to send chunk: %v", channel.ID, err)
							return err
						}

						// 等待100毫秒
						time.Sleep(100 * time.Millisecond)
					}
				}
			}
		},
		CloseChannel: channelhandler.NewCloseChannelHandler(ctx),
	})

	handler.SetLogger(logging.WithName("ChatChannelHandler"))
	handler.RegisterHandler()
}
