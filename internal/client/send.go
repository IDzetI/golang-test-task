package client

import (
	"context"
	"github.com/IDzetI/golang-test-task/pkg/service"
	"time"
)

// Send function implements sending data to the service in one thread
func (c *client) Send(ctx context.Context, data service.Batch) (sent int, err error) {

	// check input data for empty
	if data == nil || len(data) == 0 {
		return
	}

	//set initial value
	dataLength, isLastIteration := len(data), false
	limit, duration := c.service.GetLimits()

	for true {
		// calculate left border on next bach in data
		leftBachBorder := sent + int(limit)
		if leftBachBorder > dataLength {
			leftBachBorder = dataLength
			isLastIteration = true
		}

		// save processing start time
		t := time.Now()

		// send bach to service
		// we do not use the context since the documentation
		// does not contain any description of how it will be
		// used on the service side
		err = c.service.Process(ctx, (data)[sent:leftBachBorder])
		if err != nil {
			// return an error, because in addition to the
			// fact that the condition could be violated,
			// the service may simply be unavailable or there
			// may be an error in the data that should be processed
			return
		}

		// check if all data has been sent,
		// otherwise update the value of successfully sent
		sent = leftBachBorder
		if isLastIteration {
			break
		}

		// calculate required waiting time
		sleepTime := duration - time.Now().Sub(t)

		// check the waiting time, since the work takes place in one thread
		// and the service could work enough time when processing
		// the sent data so that waiting would not be required
		if sleepTime > 0 {
			time.Sleep(sleepTime)
		}
	}
	return
}
