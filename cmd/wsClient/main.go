package main

import (
	"fmt"
	pkgClient "go-backend/pkg/websocket/client"
	"log"
	"time"
)

func main() {
	// 创建客户端配置
	options := pkgClient.SocketOptions{
		URL:               "ws://localhost:8088/ws",
		Token:             "456465465",
		HeartbeatInterval: 30 * time.Second,
		Debug:             true,
		ErrorHandler: func(msg pkgClient.ErrorMsgData) {
			log.Printf("WebSocket error: %s - %s", msg.Code, msg.Detail)
		},
	}

	// 创建客户端实例
	client := pkgClient.NewSocketClient(options)

	client.OnRefreshToken(func() (string, error) {
		// 实现token刷新逻辑
		return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJnby1iYWNrZW5kIiwic3ViIjoiNTg0MTE4MTg2NDEzMTI5NzI5QDU4NTYzMjkxMzM0OTkzNjE1MCIsImV4cCI6MTc2MDM3NDkzOSwibmJmIjoxNzU5NTEwOTM5LCJpYXQiOjE3NTk1MTA5MzksInVzZXJfaWQiOjU4NDExODE4NjQxMzEyOTcyOSwiY2xpZW50RGV2aWNlSWQiOjU4NTYzMjkxMzM0OTkzNjE1MCwiaXNSZWZyZXNoIjp0cnVlLCJleHBpdHkiOjE3NjAzNzQ5MzkxOTEsInJlbWVtYmVyTWUiOmZhbHNlfQ.lD0VyOImNp5ZJYVbJ0gDsf9vg_EouTFczbfM7fgkcGI", nil
	})

	// 监听连接状态变化
	stateUnsub := client.OnStateChange(func(state pkgClient.WebSocketState) {
		log.Printf("Connection state changed to: %s", state.String())
	})

	// 连接到服务器
	conn, err := client.Connect()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	<-conn

	// 订阅消息
	unsub1 := client.Subscribe("user/+/message", func(data interface{}, topic string) {
		log.Printf("Received message on topic %s: %+v", topic, data)
	})

	// 订阅系统通知
	unsub2 := client.Subscribe("system/#", func(data interface{}, topic string) {
		log.Printf("System notification on topic %s: %+v", topic, data)
	})

	// 发送消息
	if err := client.SendMessage("user/123/message", "Hello, World!"); err != nil {
		log.Printf("Failed to send message: %v", err)
	}

	hookChan := make(chan struct{})
	// 注册频道开放钩子 - 监听服务器主动创建的频道
	channelHookUnsub := client.RegisterChannelOpen("test_handler/+", func(channel pkgClient.Channel) {
		log.Printf("新频道打开: topic: %s", channel.Topic())

		count := 0
		ticker := time.NewTicker(10 * time.Millisecond)

		// 启动一个 goroutine 定期发送消息
		go func() {
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					// 这里我们无法直接检查频道是否关闭，所以使用 channel.Wait() 来判断
					select {
					case <-channel.Wait():
						log.Printf("频道钩子检测到频道已关闭，停止发送消息")
						hookChan <- struct{}{}
						return
					default:
						// 频道仍然开放，发送消息
						count++
						message := fmt.Sprintf("hello from hook %d", count)
						if err := channel.Send(message); err != nil {
							log.Printf("频道钩子发送消息失败: %v", err)
							return
						}
					}
				case <-channel.Wait():
					log.Printf("频道钩子检测到频道关闭")
					return
				}
			}
		}()

		// 设置频道关闭处理器
		channel.OnClose(func(reason pkgClient.ErrorMsgData) {
			log.Printf("频道钩子: 频道已关闭: %s - %s", reason.Code, reason.Detail)
		})
	})

	// 创建频道进行双向通信
	channel, err := client.CreateChannel("test_panic/1",
		func(data interface{}) {
			log.Printf("频道消息: %+v", data)
		},
		func(reason pkgClient.ErrorMsgData) {
			log.Printf("err: Channel closed: %s - %s", reason.Code, reason.Detail)
		},
	)
	if err != nil {
		log.Printf("频道创建失败: %v", err)
	} else {
		log.Printf("频道创建成功")

		// 创建一个计数器和定时器
		count := 0
		ticker := time.NewTicker(100 * time.Millisecond)

		// 启动一个 goroutine 定期发送消息
		go func() {
			for range ticker.C {
				count++
				message := fmt.Sprintf("hello from client %d", count)
				if err := channel.Send(message); err != nil {
					log.Printf("Failed to send channel message: %v", err)
					ticker.Stop()
					return
				}
			}
		}()

		// 设置频道关闭处理器
		channel.OnClose(func(reason pkgClient.ErrorMsgData) {
			log.Printf("频道已关闭: %s - %s", reason.Code, reason.Detail)
			ticker.Stop()
		})
	}

	// 等待频道结束或超时
	if channel != nil {
		log.Printf("等待频道结束...")
		<-channel.Wait()
	}

	// 等待hook结束
	log.Printf("等待频道钩子结束...")
	<-hookChan
	// 取消订阅
	unsub1()
	unsub2()
	stateUnsub()
	channelHookUnsub() // 取消频道开放钩子

	// 断开连接
	log.Printf("断开连接...")
	disconnectDone := client.Disconnect()
	<-disconnectDone // 等待断开连接完成
}
