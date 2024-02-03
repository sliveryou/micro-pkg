package consistenthash

import (
	"context"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

func init() {
	RegisterBuilder(DefaultKey)
}

func TestChPicker_PickErr(t *testing.T) {
	builder := &chPickerBuilder{chKey: DefaultKey}
	picker := builder.Build(base.PickerBuildInfo{})
	_, err := picker.Pick(balancer.PickInfo{})
	require.EqualError(t, err, balancer.ErrNoSubConnAvailable.Error())
}

func TestChPicker_Pick(t *testing.T) {
	cases := []struct {
		name       string
		candidates int
	}{
		{
			name:       "single",
			candidates: 1,
		},
		{
			name:       "two",
			candidates: 2,
		},
		{
			name:       "multiple",
			candidates: 10,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			const total = 10000
			builder := &chPickerBuilder{chKey: DefaultKey}
			buildInfo := base.PickerBuildInfo{ReadySCs: make(map[balancer.SubConn]base.SubConnInfo)}
			for i := 0; i < c.candidates; i++ {
				sc := &mockClientConn{address: strconv.Itoa(i)}
				buildInfo.ReadySCs[sc] = base.SubConnInfo{
					Address: resolver.Address{Addr: strconv.Itoa(i)},
				}
			}

			picker := builder.Build(buildInfo)
			var wg sync.WaitGroup
			wg.Add(total)
			for i := 0; i < total; i++ {
				go func() {
					_, _ = picker.Pick(balancer.PickInfo{
						FullMethodName: "/",
						Ctx:            context.Background(),
					})
					wg.Done()
				}()
			}

			wg.Wait()
			pk, ok := picker.(*chPicker)
			assert.True(t, ok)
			t.Logf("%+v", pk.ch)
		})
	}
}

type mockClientConn struct {
	address string
}

func (m *mockClientConn) UpdateAddresses(addresses []resolver.Address) {
}

func (m *mockClientConn) Connect() {
}

func (m *mockClientConn) GetOrBuildProducer(builder balancer.ProducerBuilder) (p balancer.Producer, closeFunc func()) {
	return
}

func (m *mockClientConn) Shutdown() {
}
