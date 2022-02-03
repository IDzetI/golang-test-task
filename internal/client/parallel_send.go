package client

import (
	"context"
	"github.com/IDzetI/golang-test-task/pkg/service"
	"math"
	"time"
)

// ParallelSend implements sending data in multithreading mode
func (c *client) ParallelSend(ctx context.Context, data service.Batch, errChan chan error, unsentDataChan chan service.Batch) {

	// check input data for empty
	if data == nil || len(data) == 0 {
		return
	}

	//set initial value
	sent, dataMaxIndex, isLastIteration := 0, len(data)-1, false
	limit, duration := c.service.GetLimits()

	// create a channel to make sure that all parallel processing has been done
	numberOfIteration := int(math.Ceil(float64(dataMaxIndex) / float64(limit)))
	done := make(chan bool, numberOfIteration)

	for true {
		// calculate left border on next bach in data
		leftBachBorder := sent + int(limit)
		if leftBachBorder > dataMaxIndex {
			leftBachBorder = dataMaxIndex
			isLastIteration = true
		}
		// save processing start time
		t := time.Now()

		// we start data processing in a separate thread, since we assume that
		// the processing time will be greater than the required waiting time,
		// we pass parameters without using pointers, since they may change during the sending
		go func(rightBachBorder, leftBachBorder int, isLastIteration bool, ctx context.Context,
			batch service.Batch, errChan chan error, unsentDataChan chan service.Batch,
			done chan bool) {

			// send bach to service
			// we do not use the context since the documentation
			// does not contain any description of how it will be
			// used on the service side
			err := c.service.Process(ctx, data[rightBachBorder:leftBachBorder])
			if err != nil {
				errChan <- err
				unsentDataChan <- data[rightBachBorder:leftBachBorder]
			}

			// sending information about the completion of processing
			done <- true

		}(sent, leftBachBorder, isLastIteration, ctx, data, errChan, unsentDataChan, done)

		// check if all data has been sent,
		// otherwise update the value of successfully sent
		if isLastIteration {
			break
		} else {
			sent = leftBachBorder
		}

		// calculate required waiting time
		sleepTime := time.Now().Sub(t) - duration

		// check the waiting time, since the work takes place in one thread
		// and the service could work enough time when processing
		// the sent data so that waiting would not be required
		if sleepTime > 0 {
			time.Sleep(sleepTime)
		}
	}

	// waiting for the completion of processes
	for i := 0; i < numberOfIteration; i++ {
		<-done
	}
}
