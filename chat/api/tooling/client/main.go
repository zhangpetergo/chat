package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
)

func main() {
	err := hack1()
	if err != nil {
		log.Fatal(err)
	}
}

func hack1() error {
	// 客户端访问 websocket 服务端
	const url = "ws://localhost:9000/connect"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	defer conn.Close()

	// -------------------------------------------------------------------------
	// 读取服务端返回的消息
	
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	if string(msg) != "HELLO" {
		return fmt.Errorf("unexpected message: %s", msg)
	}

	// -------------------------------------------------------------------------
	// 向服务端发送消息 {"id":"8ce5af7a-788c-4c83-8e70-4500b775b359","name":"Alice"}

	usr := struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}{
		ID:   uuid.New(),
		Name: "Alice",
	}

	data, err := json.Marshal(usr)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	err = conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}

	// -------------------------------------------------------------------------
	// 读取服务端返回的消息

	_, msg, err = conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	fmt.Println(string(msg))

	return nil
}
