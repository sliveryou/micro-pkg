package xwebsocket

import "time"

// Option 可选配置
type Option func(c *Client)

// WithConnTime 配置连接时间
func WithConnTime(connTime time.Time) Option {
	return func(c *Client) {
		c.connTime = connTime
	}
}

// WithAppID 配置应用ID，可用于区分该 websocket 连接来自哪个应用，常见如："web", "android" 和 "ios"
func WithAppID(appID string) Option {
	return func(c *Client) {
		c.appID = appID
	}
}

// WithUserID 配置用户ID，可用于区分该 websocket 连接属于哪个用户
func WithUserID(userID string) Option {
	return func(c *Client) {
		c.userID = userID
	}
}

// WithHandlers 配置消息处理器，可用于集中处理从客户端读取到的消息
func WithHandlers(handlers ...Handler) Option {
	return func(c *Client) {
		c.handlers = handlers
	}
}

// WithChanSize 配置通道大小
func WithChanSize(chanSize int) Option {
	return func(c *Client) {
		c.msgChan = make(chan *Msg, chanSize)
	}
}
