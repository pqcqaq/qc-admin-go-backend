package client

import (
	"fmt"
	"log"
	"time"
)

// Example 展示如何使用WebSocket客户端
func Example() {
	// 创建客户端配置
	options := SocketOptions{
		URL:               "ws://localhost:8080/ws",
		Token:             "your-auth-token",
		HeartbeatInterval: 30 * time.Second,
		Debug:             true,
		RefreshToken: func() (string, error) {
			// 实现token刷新逻辑
			return "new-token", nil
		},
		ErrorHandler: func(msg ErrorMsgData) {
			log.Printf("WebSocket error: %s - %s", msg.Code, msg.Detail)
		},
	}

	// 创建客户端实例
	client := NewSocketClient(options)

	// 监听连接状态变化
	stateUnsub := client.OnStateChange(func(state WebSocketState) {
		log.Printf("Connection state changed to: %s", state.String())
	})
	defer stateUnsub()

	// 连接到服务器
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// 订阅消息
	unsub1 := client.Subscribe("user/+/message", func(data interface{}, topic string) {
		log.Printf("Received message on topic %s: %+v", topic, data)
	})
	defer unsub1()

	// 订阅系统通知
	unsub2 := client.Subscribe("system/#", func(data interface{}, topic string) {
		log.Printf("System notification on topic %s: %+v", topic, data)
	})
	defer unsub2()

	// 发送消息
	if err := client.SendMessage("user/123/message", map[string]interface{}{
		"text": "Hello, World!",
		"type": "text",
	}); err != nil {
		log.Printf("Failed to send message: %v", err)
	}

	// 创建频道进行双向通信
	channel, err := client.CreateChannel("chat/room1",
		func(data interface{}) {
			log.Printf("Channel message received: %+v", data)
		},
		func(reason ErrorMsgData) {
			log.Printf("Channel closed: %s - %s", reason.Code, reason.Detail)
		},
	)
	if err != nil {
		log.Printf("Failed to create channel: %v", err)
	} else {
		// 通过频道发送消息
		if err := channel.Send(map[string]interface{}{
			"message": "Hello from channel!",
		}); err != nil {
			log.Printf("Failed to send channel message: %v", err)
		}

		// 设置频道关闭处理器
		channel.OnClose(func(reason ErrorMsgData) {
			log.Printf("Channel was closed by server: %s - %s", reason.Code, reason.Detail)
		})

		// 在需要时关闭频道
		defer channel.Close()
	}

	// 保持连接一段时间
	time.Sleep(10 * time.Second)

	// 断开连接
	client.Disconnect()
}

// ExampleBasicUsage 基本使用示例
func ExampleBasicUsage() {
	// 最简单的使用方式
	client := NewSocketClient(SocketOptions{
		URL:   "ws://localhost:8080/ws",
		Token: "your-token",
		Debug: true,
	})

	// 连接
	if err := client.Connect(); err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer client.Disconnect()

	// 订阅并处理消息
	unsub := client.Subscribe("notifications", func(data interface{}, topic string) {
		fmt.Printf("Notification: %+v\n", data)
	})
	defer unsub()

	// 发送消息
	client.SendMessage("ping", "pong")

	// 等待
	time.Sleep(5 * time.Second)
}

// ExampleChannelUsage 频道使用示例
func ExampleChannelUsage() {
	client := NewSocketClient(SocketOptions{
		URL:   "ws://localhost:8080/ws",
		Token: "your-token",
		Debug: true,
	})

	if err := client.Connect(); err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer client.Disconnect()

	// 创建文件上传频道
	channel, err := client.CreateChannel("upload/file",
		func(data interface{}) {
			// 处理上传进度或结果
			if dataMap, ok := data.(map[string]interface{}); ok {
				if progress, exists := dataMap["progress"]; exists {
					fmt.Printf("Upload progress: %v%%\n", progress)
				}
				if status, exists := dataMap["status"]; exists && status == "completed" {
					fmt.Println("Upload completed!")
				}
			}
		},
		func(reason ErrorMsgData) {
			fmt.Printf("Upload failed: %s\n", reason.Detail)
		},
	)

	if err != nil {
		log.Fatalf("Failed to create upload channel: %v", err)
	}

	// 发送文件数据
	channel.Send(map[string]interface{}{
		"filename": "document.pdf",
		"size":     1024000,
		"chunk":    1,
	})

	// 等待上传完成
	time.Sleep(30 * time.Second)
	channel.Close()
}
