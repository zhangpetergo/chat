package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/zhangpetergo/chat/chat/foundation/logger"
	"go.uber.org/zap"
	"sync"
	"time"
)

var ErrUserExists = fmt.Errorf("user already exists")
var ErrUserNotExists = fmt.Errorf("user not exists")

type Chat struct {
	users map[uuid.UUID]connection
	mu    sync.RWMutex
}

func NewChat() *Chat {
	return &Chat{
		users: make(map[uuid.UUID]connection),
	}
}

// HandleShake 如果 func 需要 struct 的成员变量，那么 func 必须是 struct 的方法
// 比如说使用 logger.Log，那么 handleShake 必须是 app 的方法
// 只不过这里的 logger 是全局变量，如果使用依赖注入，那么需要app
func (c *Chat) HandleShake(conn *websocket.Conn, traceID string) error {
	// 服务器向客户端发送握手消息
	if err := conn.WriteMessage(websocket.TextMessage, []byte("HELLO")); err != nil {
		return err
	}

	// 等待客户端发送 UUID,name
	// 不能一直等待，设置超时时间
	//conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	msg, err := c.readMessage(ctx, conn, traceID)
	if err != nil {
		return fmt.Errorf("read message: %w", err)
	}

	var usr user

	err = json.Unmarshal(msg, &usr)
	if err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	// 添加用户
	if err := c.addUser(usr, conn); err != nil {
		defer conn.Close()
		// 用户已经存在
		if err := conn.WriteMessage(websocket.TextMessage, []byte("Already connected")); err != nil {
			return fmt.Errorf("write message: %w", err)
		}
		return fmt.Errorf("add user: %w", err)
	}

	// 服务器向客户端发送 WELCOME name
	v := fmt.Sprintf("WELCOME %s", usr.Name)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(v)); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	logger.Log.With(zap.String("uuid", traceID)).Infow("handshake completed", "user", usr)

	return nil
}

func (c *Chat) SendMessage(msg inMessage) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// 如果用户不存在，返回错误
	from, exists := c.users[msg.FromID]
	if !exists {
		return ErrUserNotExists
	}
	to, exists := c.users[msg.ToID]
	if !exists {
		return ErrUserNotExists
	}

	// 构建消息
	m := outMessage{
		From: user{
			ID:   from.id,
			Name: from.name,
		},
		To: user{
			ID:   to.id,
			Name: to.name,
		},
		Msg: msg.Msg,
	}

	if err := to.conn.WriteJSON(m); err != nil {
		// 如果发送消息失败，移除用户
		// 这里不用手动移除用户，因为 ping goroutine 会检测到连接断开，自动移除用户
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

func (c *Chat) Listen(ctx context.Context, conn *websocket.Conn, traceID string) {
	for {
		msg, err := c.readMessage(ctx, conn, traceID)
		if err != nil {
			logger.Log.Error("read message failed", zap.Error(err))
			return
		}

		var inMsg inMessage
		err = json.Unmarshal(msg, &inMsg)
		if err != nil {
			logger.Log.Error("unmarshal message failed", zap.Error(err))
			return
		}
		// 发送信息到对应的用户
		err = c.SendMessage(inMsg)
		if err != nil {
			logger.Log.Error("send message failed", zap.Error(err))
		}

	}
}

// 创建所有连接的副本
func (c *Chat) connections() map[uuid.UUID]connection {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// 创建所有连接的副本
	m := make(map[uuid.UUID]connection, len(c.users))
	for k, v := range c.users {
		m[k] = v
	}
	return m
}

func (c *Chat) Ping() {
	ticker := time.NewTicker(time.Second * 10)
	for {

		<-ticker.C
		// 如何取消每次ping的时候的加锁操作
		m := c.connections()
		for k, conn := range m {
			if err := conn.conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				logger.Log.Error("ping failed", zap.Error(err))
				c.removeUser(k)
			}
		}

	}
}

// -------------------------------------------------------------------------

// addUser 添加用户，如果用户已经存在，返回错误
func (c *Chat) addUser(usr user, conn *websocket.Conn) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 如果用户已经存在，返回错误
	if _, exists := c.users[usr.ID]; exists {
		return fmt.Errorf("user already exists")
	}
	// 添加用户
	c.users[usr.ID] = connection{
		id:   usr.ID,
		name: usr.Name,
		conn: conn,
	}
	return nil
}

// removeUser 移除用户，如果用户不存在，返回错误
func (c *Chat) removeUser(userID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果用户不存在，返回错误
	conn, exists := c.users[userID]
	if !exists {
		return
	}
	delete(c.users, userID)
	// 关闭连接
	conn.conn.Close()
}

func (c *Chat) readMessage(ctx context.Context, conn *websocket.Conn, traceID string) ([]byte, error) {

	type response struct {
		message []byte
		err     error
	}

	// 异步等待客户端发送消息
	// 后续处理消息
	// channel 设置缓冲区大小为1的原因是避免发生goroutine泄露
	ch := make(chan response, 1)
	go func() {
		logger.Log.With(zap.String("uuid", traceID)).Info("starting handshake read")
		defer logger.Log.With(zap.String("uuid", traceID)).Info("completed handshake read")
		_, msg, err := conn.ReadMessage()

		if err != nil {
			ch <- response{message: nil, err: err}
			return
		}
		// 将消息返回给主协程
		ch <- response{message: msg, err: nil}
	}()

	// 主协程处理消息

	var resp response
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp = <-ch:
		if resp.err != nil {
			return nil, resp.err
		}
	}
	return resp.message, nil
}
