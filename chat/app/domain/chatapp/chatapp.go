package chatapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/zhangpetergo/chat/chat/app/sdk/errs"
	"github.com/zhangpetergo/chat/chat/foundation/logger"
	"time"
)

type app struct {
	WS websocket.Upgrader
}

func NewApp() *app {
	return &app{}
}

func (a *app) connect(c *gin.Context) {
	// client connect websocket
	// 升级http协议为websocket协议
	conn, err := a.WS.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		c.Error(errs.Newf(errs.FailedPrecondition, "websocket upgrade failed: %v", err))
		return
	}
	defer conn.Close()

	usr, err := a.handleShake(conn)
	if err != nil {
		c.Error(errs.Newf(errs.FailedPrecondition, "handshake failed: %v", err))
		return
	}

	logger.Log.Infow("handshake completed", "user", usr)

	//// 创建一个ticker 用于心跳检测
	//ticker := time.NewTicker(time.Second)
	//
	//go func() {
	//	select {
	//	case <-ticker.C:
	//		// 发送心跳检测
	//		if err := conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
	//			c.Error(errs.Newf(errs.FailedPrecondition, "ping failed: %v", err))
	//			return
	//		}
	//	}
	//}()

}

func (a *app) handleShake(conn *websocket.Conn) (user, error) {
	// 服务器向客户端发送握手消息
	if err := conn.WriteMessage(websocket.TextMessage, []byte("HELLO")); err != nil {
		return user{}, err
	}

	// 等待客户端发送 UUID,name
	// 不能一直等待，设置超时时间
	//conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	msg, err := a.readMessage(ctx, conn)
	if err != nil {
		return user{}, fmt.Errorf("read message: %w", err)
	}

	var usr user

	err = json.Unmarshal(msg, &usr)
	if err != nil {
		return user{}, fmt.Errorf("unmarshal: %w", err)
	}
	// 服务器向客户端发送 WELCOME name
	v := fmt.Sprintf("WELCOME %s", usr.Name)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(v)); err != nil {
		return user{}, fmt.Errorf("write message: %w", err)
	}

	return usr, nil
}

func (a *app) readMessage(ctx context.Context, conn *websocket.Conn) ([]byte, error) {

	type response struct {
		message []byte
		err     error
	}

	// 异步等待客户端发送消息
	// 后续处理消息
	// channel 设置缓冲区大小为1的原因是避免发生goroutine泄露
	ch := make(chan response, 1)
	go func() {
		logger.Log.Info("starting handshake read")
		defer logger.Log.Info("completed handshake read")
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

func (a *app) test(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello World",
	})
}

func (a *app) testError(c *gin.Context) {
	c.Error(errs.NewError(errors.New("text error")))
	return
}

func (a *app) testPanic(c *gin.Context) {
	panic("Hello World")
}
