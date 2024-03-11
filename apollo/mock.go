package apollo

import agollo "github.com/philchia/agollo/v4"

// MockClient ...
type MockClient struct{}

// Start ...
func (c *MockClient) Start() error {
	return nil
}

// Stop ...
func (c *MockClient) Stop() error {
	return nil
}

// OnUpdate ...
func (c *MockClient) OnUpdate(f func(*agollo.ChangeEvent)) {
}

// GetString ...
func (c *MockClient) GetString(key string, opts ...agollo.OpOption) string {
	return ""
}

// GetContent ...
func (c *MockClient) GetContent(opts ...agollo.OpOption) string {
	return ""
}

// GetPropertiesContent ...
func (c *MockClient) GetPropertiesContent(opts ...agollo.OpOption) string {
	return ""
}

// GetAllKeys ...
func (c *MockClient) GetAllKeys(opts ...agollo.OpOption) []string {
	return nil
}

// GetReleaseKey ...
func (c *MockClient) GetReleaseKey(opts ...agollo.OpOption) string {
	return ""
}

// SubscribeToNamespaces ...
func (c *MockClient) SubscribeToNamespaces(namespaces ...string) error {
	return nil
}
