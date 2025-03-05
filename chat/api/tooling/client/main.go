package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/zhangpetergo/chat/chat/foundation/logger"
	"go.uber.org/zap"
	"log"
	"os"
)

func main() {
	err := hack1()
	if err != nil {
		log.Fatal(err)
	}
}

func hack1() error {
	users := []uuid.UUID{
		uuid.MustParse("8ce5af7a-788c-4c83-8e70-4500b775b359"),
		uuid.MustParse("d92d3e84-a08d-4d55-b211-8199299495a2"),
	}

	var ID uuid.UUID
	switch os.Args[1] {
	case "0":
		ID = users[0]
	case "1":
		ID = users[1]
	}

	fmt.Println("ID:", ID.String())

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
		ID:   ID,
		Name: "Peter",
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

	go func() {
		for {
			_, msg, err = conn.ReadMessage()
			if err != nil {
				fmt.Println("read: %w", err)
				return
			}

			var outMsg outMessage
			err = json.Unmarshal(msg, &outMsg)
			if err != nil {
				logger.Log.Error("unmarshal message failed", zap.Error(err))
				return
			}

			fmt.Println(outMsg.Msg)
		}
	}()

	// -------------------------------------------------------------------------
	// handshake 完成后发送消息

	// 生成一个 uuid
	// 8ce5af7a-788c-4c83-8e70-4500b775b359
	// d92d3e84-a08d-4d55-b211-8199299495a2

	for {
		fmt.Printf("\n\n")
		fmt.Print("message>")
		reader := bufio.NewReader(os.Stdin)

		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("read string: %w", err)
		}

		var from, to uuid.UUID

		switch os.Args[1] {
		case "0":
			from = users[0]
			to = users[1]
		case "1":
			from = users[1]
			to = users[0]
		}

		inMsg := inMessage{
			FromID: from,
			ToID:   to,
			Msg:    input,
		}

		data, err = json.Marshal(inMsg)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}
		err = conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			return fmt.Errorf("write: %w", err)
		}
	}
}

type inMessage struct {
	FromID uuid.UUID `json:"fromID"`
	ToID   uuid.UUID `json:"toID"`
	Msg    string    `json:"msg"`
}

type outMessage struct {
	From user   `json:"from"`
	To   user   `json:"to"`
	Msg  string `json:"msg"`
}

type user struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
