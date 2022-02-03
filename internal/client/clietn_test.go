package client

import (
	"context"
	"github.com/IDzetI/golang-test-task/pkg/service"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

type testService struct {
	limit          uint64
	duration       time.Duration
	available      int
	availableMutex sync.Mutex
	received       int
}

func newTestService(n uint64, p time.Duration) *testService {
	return &testService{
		limit:     n,
		duration:  p,
		available: int(n),
		received:  0,
	}
}

func (t *testService) GetReceived() int {
	return t.received
}

func (t *testService) GetLimits() (n uint64, p time.Duration) {
	return t.limit, t.duration
}

func (t *testService) updateAvailable(count int) {
	time.Sleep(t.duration)
	t.available += count
}

func (t *testService) Process(_ context.Context, batch service.Batch) error {
	if t.available < 0 {
		return service.ErrBlocked
	}

	t.available -= len(batch)
	go t.updateAvailable(len(batch))

	t.received += len(batch)
	return nil
}

func getTestBatch(n int) service.Batch {
	var b []service.Item
	for i := 0; i < n; i++ {
		b = append(b, service.Item{})
	}
	return b
}

type args struct {
	service *testService
	ctx     context.Context
	data    service.Batch
}

type test struct {
	name string
	args args
}

func getTests() []test {
	return []test{
		{"common",
			args{
				service: newTestService(20, time.Millisecond),
				data:    getTestBatch(1325),
				ctx:     context.Background(),
			}},
		{"less then limit",
			args{
				service: newTestService(20, time.Millisecond),
				data:    getTestBatch(5),
				ctx:     context.Background(),
			}},
		{"equals to limit",
			args{
				service: newTestService(20, time.Millisecond),
				data:    getTestBatch(20),
				ctx:     context.Background(),
			}},
		{"multiple of the limit",
			args{
				service: newTestService(20, time.Millisecond),
				data:    getTestBatch(40),
				ctx:     context.Background(),
			}},
	}
}

func TestSend(t *testing.T) {

	for _, tt := range getTests() {
		t.Run(tt.name, func(t *testing.T) {

			cln := New()
			cln.SetService(tt.args.service)

			sent, err := cln.Send(tt.args.ctx, tt.args.data)

			assert.Equal(t, nil, err)
			assert.Equal(t, len(tt.args.data), sent)
			assert.Equal(t, tt.args.service.GetReceived(), len(tt.args.data))
		})
	}
}

func TestParallelSend(t *testing.T) {

	for _, tt := range getTests() {
		t.Run(tt.name, func(t *testing.T) {

			cln := New()
			cln.SetService(tt.args.service)

			errChan := make(chan error)
			unsentDataChan := make(chan service.Batch)

			cln.ParallelSend(tt.args.ctx, tt.args.data, errChan, unsentDataChan)

			select {
			case err, ok := <-errChan:
				assert.Equal(t, nil, err)
				assert.Equal(t, false, ok)
			case unsentData, ok := <-unsentDataChan:
				assert.Equal(t, nil, unsentData)
				assert.Equal(t, false, ok)
			default:
				assert.Equal(t, tt.args.service.GetReceived(), len(tt.args.data))
			}
		})
	}
}
