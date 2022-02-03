package client

import (
	"context"
	"github.com/IDzetI/golang-test-task/pkg/service"
)

// Client defines client of external service that can process batches of items.
type Client interface {

	// SetService function for provide external service to Client
	SetService(service service.Service)

	// Send provide method to send data to external service in one thread
	// we use a pointer with the input parameter, since it is more memory
	// efficient, it is dangerous to use it, because if the application
	// as a whole runs in several threads, then the data can be changed
	// in another during the sending, but we assume that this cannot happen
	Send(ctx context.Context, data service.Batch) (sent int, err error)

	// ParallelSend the function involves sending data to the service
	// in multithreading mode, since in this case we assume that the
	// data processing by the service may exceed the required delay,
	// at the same time, all errors and unsent data are sent
	// to special channels for processing
	ParallelSend(ctx context.Context, data service.Batch, errChan chan error, unsentDataChan chan service.Batch)
}

type client struct {
	service service.Service
}

func New() *client {
	return &client{}
}
