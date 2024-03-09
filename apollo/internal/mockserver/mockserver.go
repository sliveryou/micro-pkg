package mockserver

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"sync"
)

type notification struct {
	NamespaceName  string `json:"namespaceName,omitempty"`
	NotificationID int    `json:"notificationId,omitempty"`
}

type result struct {
	// AppID          string            `json:"appId"`
	// Cluster        string            `json:"cluster"`
	NamespaceName  string            `json:"namespaceName"`
	Configurations map[string]string `json:"configurations"`
	ReleaseKey     string            `json:"releaseKey"`
}

type mockServer struct {
	server http.Server

	mu            sync.Mutex
	notifications map[string]int
	config        map[string]map[string]string
}

func (s *mockServer) NotificationHandler(rw http.ResponseWriter, req *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = req.ParseForm()
	var notifications []notification
	if err := json.Unmarshal([]byte(req.FormValue("notifications")), &notifications); err != nil {
		panic(err)
	}

	var changes []notification
	for _, noti := range notifications {
		if currentID := s.notifications[noti.NamespaceName]; currentID != noti.NotificationID {
			changes = append(changes, notification{NamespaceName: noti.NamespaceName, NotificationID: currentID})
		}
	}

	if len(changes) == 0 {
		rw.WriteHeader(http.StatusNotModified)
		return
	}
	bts, _ := json.Marshal(&changes)
	_, _ = rw.Write(bts)
}

func (s *mockServer) ConfigHandler(rw http.ResponseWriter, req *http.Request) {
	_ = req.ParseForm()

	namespace, releaseKey := filepath.Base(req.URL.Path), req.FormValue("releaseKey")
	config := s.Get(namespace)

	result := result{NamespaceName: namespace, Configurations: config, ReleaseKey: releaseKey}
	bts, _ := json.Marshal(&result)
	_, _ = rw.Write(bts)
}

var server *mockServer

func (s *mockServer) Set(namespace, key, value string) {
	server.mu.Lock()
	defer server.mu.Unlock()

	notificationID := s.notifications[namespace]
	notificationID++
	s.notifications[namespace] = notificationID

	if kv, ok := s.config[namespace]; ok {
		kv[key] = value
		return
	}
	kv := map[string]string{key: value}
	s.config[namespace] = kv
}

func (s *mockServer) Get(namespace string) map[string]string {
	server.mu.Lock()
	defer server.mu.Unlock()

	return s.config[namespace]
}

func (s *mockServer) Delete(namespace, key string) {
	server.mu.Lock()
	defer server.mu.Unlock()

	if kv, ok := s.config[namespace]; ok {
		delete(kv, key)
	}

	notificationID := s.notifications[namespace]
	notificationID++
	s.notifications[namespace] = notificationID
}

// Set namespace's key value
func Set(namespace, key, value string) {
	server.Set(namespace, key, value)
}

// Delete namespace's key
func Delete(namespace, key string) {
	server.Delete(namespace, key)
}

// Run mock server
func Run(addr ...string) error {
	initServer(addr...)
	return server.server.ListenAndServe()
}

func initServer(addr ...string) {
	server = &mockServer{
		notifications: map[string]int{},
		config:        map[string]map[string]string{},
	}
	mux := http.NewServeMux()
	mux.Handle("/notifications/", http.HandlerFunc(server.NotificationHandler))
	mux.Handle("/configs/", http.HandlerFunc(server.ConfigHandler))
	server.server.Handler = mux
	server.server.Addr = ":8080"
	if len(addr) > 0 {
		server.server.Addr = addr[0]
	}
}

// Close mock server
func Close() error {
	return server.server.Shutdown(context.TODO())
}
