package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/zhangpetergo/chat/chat/app/sdk/errs"
	"github.com/zhangpetergo/chat/chat/foundation/logger"
	"github.com/zhangpetergo/chat/chat/foundation/web"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

var ErrUserExists = fmt.Errorf("user already exists")
var ErrUserNotExists = fmt.Errorf("user not exists")

type Chat struct {
	users map[uuid.UUID]User
	mu    sync.RWMutex
}

func NewChat() *Chat {
	c := Chat{
		users: make(map[uuid.UUID]User),
	}
	c.Ping()
	return &c
}

// HandleShake 如果 func 需要 struct 的成员变量，那么 func 必须是 struct 的方法
// 比如说使用 logger.Log，那么 handleShake 必须是 app 的方法
// 只不过这里的 logger 是全局变量，如果使用依赖注入，那么需要app
func (c *Chat) HandleShake(ctx context.Context, w http.ResponseWriter, r *http.Request) (User, error) {

	var ws websocket.Upgrader
	// client connect websocket
	// 升级http协议为websocket协议
	conn, err := ws.Upgrade(w, r, nil)
	if err != nil {
		return User{}, errs.Newf(errs.FailedPrecondition, "websocket upgrade failed: %v", err)
	}

	// 服务器向客户端发送握手消息
	if err := conn.WriteMessage(websocket.TextMessage, []byte("HELLO")); err != nil {
		return User{}, err
	}

	// 等待客户端发送 UUID,name
	// 不能一直等待，设置超时时间
	//conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*100)
	defer cancel()

	usr := User{
		Conn: conn,
	}

	msg, err := c.readMessage(ctx, usr)
	if err != nil {
		return User{}, fmt.Errorf("read message: %w", err)
	}

	err = json.Unmarshal(msg, &usr)
	if err != nil {
		return User{}, fmt.Errorf("unmarshal: %w", err)
	}

	// 添加用户
	if err := c.addUser(ctx, usr); err != nil {
		defer conn.Close()
		// 用户已经存在
		if err := conn.WriteMessage(websocket.TextMessage, []byte("Already connected")); err != nil {
			return User{}, fmt.Errorf("write message: %w", err)
		}
		return User{}, fmt.Errorf("add User: %w", err)
	}

	// 服务器向客户端发送 WELCOME name
	v := fmt.Sprintf("WELCOME %s", usr.Name)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(v)); err != nil {
		return User{}, fmt.Errorf("write message: %w", err)
	}

	logger.Log.With(zap.String("uuid", web.GetTraceID(ctx).String())).Infow("handshake completed", "User", usr)

	return usr, nil
}

// =============================================================================

func (c *Chat) Listen(ctx context.Context, usr User) {
	for {

		msg, err := c.readMessage(ctx, usr)
		if err != nil {
			// 如果这里客户端主动断开连接，我们返回，结束连接
			// 或者是context取消，结束连接
			if c.isCriticalError(ctx, err) {
				return
			}
			// 其他错误，继续等待
			continue
		}

		var inMsg inMessage
		err = json.Unmarshal(msg, &inMsg)
		if err != nil {
			logger.Log.Infow("chat-listen-unmarshal", "uuid", web.GetTraceID(ctx).String(), "err", err)
			continue
		}
		// 发送信息到对应的用户
		err = c.sendMessage(inMsg)
		if err != nil {
			logger.Log.Infow("chat-listen-send", "uuid", web.GetTraceID(ctx).String(), "err", err)
		}

	}
}

func (c *Chat) isCriticalError(ctx context.Context, err error) bool {
	switch err.(type) {
	case *websocket.CloseError:
		logger.Log.Infow("chat-isCriticalError", "uuid", web.GetTraceID(ctx).String(), "status", "client disconnected")
		return true

	default:
		if errors.Is(err, context.Canceled) {
			logger.Log.Infow("chat-isCriticalError", "uuid", web.GetTraceID(ctx).String(), "status", "client canceled")
			return true
		}

		logger.Log.Infow("chat-isCriticalError", "uuid", web.GetTraceID(ctx).String(), "err", err)
		return false
	}
}

func (c *Chat) readMessage(ctx context.Context, usr User) ([]byte, error) {

	type response struct {
		message []byte
		err     error
	}

	// 异步等待客户端发送消息
	// 后续处理消息
	// channel 设置缓冲区大小为1的原因是避免发生goroutine泄露
	ch := make(chan response, 1)
	go func() {

		logger.Log.Infow("chat-readMessage", "uuid", web.GetTraceID(ctx).String(), "status", "started")
		defer logger.Log.Infow("chat-readMessage", "uuid", web.GetTraceID(ctx).String(), "status", "completed")
		_, msg, err := usr.Conn.ReadMessage()

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
		c.removeUser(ctx, usr.ID)
		return nil, ctx.Err()
	case resp = <-ch:
		if resp.err != nil {
			c.removeUser(ctx, usr.ID)
			return nil, resp.err
		}
	}
	return resp.message, nil
}

func (c *Chat) sendMessage(msg inMessage) error {
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
		From: User{
			ID:   from.ID,
			Name: from.Name,
		},
		To: User{
			ID:   to.ID,
			Name: to.Name,
		},
		Msg: msg.Msg,
	}

	if err := to.Conn.WriteJSON(m); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

// 创建所有连接的副本
func (c *Chat) connections() map[uuid.UUID]*websocket.Conn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// 创建所有连接的副本
	m := make(map[uuid.UUID]*websocket.Conn, len(c.users))
	for k, v := range c.users {
		m[k] = v.Conn
	}
	return m
}

func (c *Chat) Ping() {
	ticker := time.NewTicker(time.Second * 10)
	go func() {

		ctx := context.Background()
		for {

			<-ticker.C

			logger.Log.Infow("ping", "uuid", web.GetTraceID(ctx).String())

			// 如何取消每次ping的时候的加锁操作
			m := c.connections()
			for k, conn := range m {
				if err := conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
					logger.Log.Error("ping failed", zap.Error(err))
					c.removeUser(ctx, k)
				}
			}

		}
	}()

}

// -------------------------------------------------------------------------

// addUser 添加用户，如果用户已经存在，返回错误
func (c *Chat) addUser(ctx context.Context, usr User) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 如果用户已经存在，返回错误
	if _, exists := c.users[usr.ID]; exists {
		return fmt.Errorf("user already exists")
	}
	// 添加用户
	c.users[usr.ID] = usr
	logger.Log.Infow("add user", "uuid", web.GetTraceID(ctx), "user", usr)
	return nil
}

// removeUser 移除用户，如果用户不存在，返回错误
func (c *Chat) removeUser(ctx context.Context, userID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果用户不存在，返回错误
	conn, exists := c.users[userID]
	if !exists {
		return
	}
	delete(c.users, userID)
	logger.Log.Infow("remove user", "uuid", web.GetTraceID(ctx).String(), "user", userID)
	// 关闭连接
	conn.Conn.Close()
}
