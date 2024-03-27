package xwebsocket

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/core/contextx"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"

	"github.com/sliveryou/go-tool/v2/id-generator/uuid"
)

// Manager websocket 管理器
type Manager struct {
	Upgrader          *websocket.Upgrader // websocket 协议升级器
	ReadTimeout       time.Duration       // 读取超时时间，要大于 HeartbeatInterval
	WriteTimeout      time.Duration       // 写入超时时间
	HeartbeatInterval time.Duration       // 心跳间隔时间
	MaxMessageSize    int64               // 最大消息字节大小

	clients *collection.SafeMap // connID -> *Client
	users   *collection.SafeMap // userID -> []*Client
	mu      sync.RWMutex
}

// NewManager 新建 websocket 管理器
func NewManager() *Manager {
	return &Manager{
		Upgrader: &websocket.Upgrader{
			ReadBufferSize:  2048,
			WriteBufferSize: 2048,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		ReadTimeout:       40 * time.Second,
		WriteTimeout:      10 * time.Second,
		HeartbeatInterval: 30 * time.Second,
		MaxMessageSize:    8192, // 8 MB

		clients: collection.NewSafeMap(),
		users:   collection.NewSafeMap(),
	}
}

// UpgradeClient 将 http 连接升级为 websocket 客户端
func (m *Manager) UpgradeClient(w http.ResponseWriter, r *http.Request, opts ...Option) (*Client, error) {
	conn, err := m.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, errors.WithMessage(err, "upgrade conn to websocket err")
	}

	client := m.register(r.Context(), conn, opts...)
	m.run(client)

	return client, nil
}

// GetClient 获取 websocket 客户端
func (m *Manager) GetClient(connID string) (*Client, bool) {
	val, ok := m.clients.Get(connID)
	if !ok {
		return nil, false
	}

	client, ok := val.(*Client)

	return client, ok
}

// GetUserClient 获取用户 websocket 客户端
func (m *Manager) GetUserClient(userID, appID string) (*Client, bool) {
	clients, ok := m.GetUserClients(userID)
	if !ok {
		return nil, false
	}

	for _, client := range clients {
		// 查找存活且 appID 相同的客户端
		if client.IsAlive() && client.appID == appID {
			return client, true
		}
	}

	return nil, false
}

// GetUserClients 获取用户 websocket 客户端列表
func (m *Manager) GetUserClients(userID string) ([]*Client, bool) {
	val, ok := m.users.Get(userID)
	if !ok {
		return nil, false
	}

	clients, ok := val.([]*Client)
	if !ok {
		return nil, false
	}

	return clients, true
}

// GetClients 获取所有 websocket 客户端
func (m *Manager) GetClients() []*Client {
	clients := make([]*Client, 0, m.clients.Size())
	m.clients.Range(func(key, val any) bool {
		if client, ok := val.(*Client); ok {
			clients = append(clients, client)
		}

		return true
	})

	return clients
}

// SendMsgToClient 发送消息至指定客户端
func (m *Manager) SendMsgToClient(connID string, msgData []byte, msgType ...int) {
	client, ok := m.GetClient(connID)
	if ok {
		client.SendMsg(msgData, msgType...)
	}
}

// SendMsgToUser 发送消息至指定用户
func (m *Manager) SendMsgToUser(userID string, msgData []byte, msgType ...int) {
	clients, ok := m.GetUserClients(userID)
	if ok {
		for _, client := range clients {
			client.SendMsg(msgData, msgType...)
		}
	}
}

// BroadcastMsg 广播消息
func (m *Manager) BroadcastMsg(msgData []byte, msgType ...int) {
	clients := m.GetClients()
	for _, client := range clients {
		client.SendMsg(msgData, msgType...)
	}
}

// DelClient 删除 websocket 客户端
func (m *Manager) DelClient(connID string) {
	client, ok := m.GetClient(connID)
	if ok {
		m.unregister(client)
	}
}

// register 注册 websocket 客户端至 websocket 管理器
func (m *Manager) register(ctx context.Context, conn *websocket.Conn, opts ...Option) *Client {
	ctx, cancel := context.WithCancel(contextx.ValueOnlyFrom(ctx))
	client := &Client{
		ctx:      ctx,
		conn:     conn,
		connID:   uuid.NextV4(),
		connTime: time.Now(),
		cancel:   cancel,
		handlers: make([]Handler, 0),
		msgChan:  make(chan *Msg, 10),
	}
	client.isAlive.Store(true)

	for _, opt := range opts {
		opt(client)
	}

	m.addUserClient(client)
	m.clients.Set(client.connID, client)

	return client
}

// unregister 注销指定 websocket 客户端
func (m *Manager) unregister(client *Client) {
	if client.isAlive.CompareAndSwap(true, false) {
		_ = client.conn.Close() // 关闭连接
		m.clients.Del(client.connID)
		m.delUserClient(client)
		client.cancel()
	}
}

// addUserClient 添加用户客户端
func (m *Manager) addUserClient(client *Client) {
	if userID := client.userID; userID != "" {
		m.mu.Lock()
		defer m.mu.Unlock()

		userClients, _ := m.GetUserClients(userID)
		userClients = append(userClients, client)

		m.users.Set(userID, userClients)
	}
}

// delUserClient 删除用户客户端
func (m *Manager) delUserClient(client *Client) {
	if userID := client.userID; userID != "" {
		m.mu.Lock()
		defer m.mu.Unlock()

		userClients, ok := m.GetUserClients(userID)
		if ok {
			var newClients []*Client
			for _, userClient := range userClients {
				if userClient.connID != client.connID {
					newClients = append(newClients, userClient)
				}
			}

			if len(newClients) > 0 {
				m.users.Set(userID, newClients)
			} else {
				m.users.Del(userID)
			}
		}
	}
}

// run 运行 websocket 客户端异步读写协程
func (m *Manager) run(client *Client) {
	ctx := client.ctx

	// 向客户端写消息
	threading.GoSafeCtx(ctx, func() {
		var err error
		ticker := time.NewTicker(m.HeartbeatInterval)

		defer func() {
			ticker.Stop()
			client.cancel()

			if shouldLog(err) {
				logx.WithContext(ctx).Errorf("xwebsocket: client %s write message err: %v", client.connID, err)
			}
		}()

		for {
			select {
			case <-ctx.Done():
				close(client.msgChan)
				err = client.conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(m.WriteTimeout))
				return
			case msg := <-client.msgChan:
				_ = client.conn.SetWriteDeadline(time.Now().Add(m.WriteTimeout))
				if err = client.conn.WriteMessage(msg.Type, msg.Data); err != nil {
					return
				}
			case <-ticker.C:
				if err = client.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(m.WriteTimeout)); err != nil {
					return
				}
			}
		}
	})

	// 从客户端读消息
	threading.GoSafeCtx(ctx, func() {
		defer func() {
			m.unregister(client)
			client.cancel()
		}()

		client.conn.SetReadLimit(m.MaxMessageSize)
		_ = client.conn.SetReadDeadline(time.Now().Add(m.ReadTimeout))
		client.conn.SetPongHandler(func(string) error {
			_ = client.conn.SetReadDeadline(time.Now().Add(m.ReadTimeout))
			return nil
		})

		for {
			select {
			case <-ctx.Done():
				return
			default:
				msgType, msgData, err := client.conn.ReadMessage()
				if err != nil {
					if shouldLog(err) {
						logx.WithContext(ctx).Errorf("xwebsocket: client %s read message err: %v", client.connID, err)
					}
					return
				}

				if client.IsAlive() {
					client.mu.RLock()
					for _, handler := range client.handlers {
						handler(client, &Msg{Type: msgType, Data: msgData})
					}
					client.mu.RUnlock()
				}
			}
		}
	})
}

// shouldLog 判断是否需要打印日志
func shouldLog(err error) bool {
	if err != nil {
		return websocket.IsUnexpectedCloseError(err,
			websocket.CloseNormalClosure,
			websocket.CloseGoingAway,
			websocket.CloseNoStatusReceived,
			websocket.CloseAbnormalClosure,
		)
	}

	return false
}
