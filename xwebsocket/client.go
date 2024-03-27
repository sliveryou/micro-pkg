package xwebsocket

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

// Client websocket 客户端
type Client struct {
	ctx      context.Context    // 上下文
	conn     *websocket.Conn    // 连接
	connID   string             // 连接ID
	connTime time.Time          // 连接时间
	appID    string             // 应用ID
	userID   string             // 用户ID
	isAlive  atomic.Bool        // 是否存活
	cancel   context.CancelFunc // 取消函数
	handlers []Handler          // 消息处理器
	msgChan  chan *Msg          // 消息通道
	mu       sync.RWMutex       // 读写锁
}

// Msg websocket 消息
type Msg struct {
	Type int    // 类型
	Data []byte // 内容
}

// Handler 消息处理器
type Handler func(client *Client, msg *Msg)

// Ctx 上下文
func (c *Client) Ctx() context.Context {
	return c.ctx
}

// Conn 连接
func (c *Client) Conn() *websocket.Conn {
	return c.conn
}

// ConnID 连接ID
func (c *Client) ConnID() string {
	return c.connID
}

// ConnTime 连接时间
func (c *Client) ConnTime() time.Time {
	return c.connTime
}

// AppID 应用ID
func (c *Client) AppID() string {
	return c.appID
}

// UserID 用户ID
func (c *Client) UserID() string {
	return c.userID
}

// IsAlive 是否存活
func (c *Client) IsAlive() bool {
	return c.isAlive.Load()
}

// Cancel 取消函数
func (c *Client) Cancel() context.CancelFunc {
	return c.cancel
}

// AddHandlers 添加消息处理器
func (c *Client) AddHandlers(handlers ...Handler) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.handlers = append(c.handlers, handlers...)
}

// SetHandlers 设置消息处理器
func (c *Client) SetHandlers(handlers ...Handler) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.handlers = handlers
}

// SendMsg 发送消息
func (c *Client) SendMsg(msgData []byte, msgType ...int) {
	if c.IsAlive() {
		c.msgChan <- &Msg{
			Type: getMsgType(msgType...),
			Data: msgData,
		}
	}
}

// SendJSON 发送 json 消息
func (c *Client) SendJSON(v any) error {
	if c.IsAlive() {
		data, err := json.Marshal(v)
		if err != nil {
			return errors.WithMessage(err, "json marshal err")
		}

		c.msgChan <- &Msg{
			Type: websocket.TextMessage,
			Data: data,
		}
	}

	return nil
}

// getMsgType 获取消息类型
func getMsgType(msgType ...int) int {
	mt := websocket.TextMessage
	if len(msgType) > 0 {
		mt = msgType[0]
	}

	return mt
}
