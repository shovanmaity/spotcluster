package main

import (
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	instancecontroller "github.com/shovanmaity/spotcluster/controller/instance"
	poolcontroller "github.com/shovanmaity/spotcluster/controller/pool"
)

// Set logging property
func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		PadLevelText:    true,
		TimestampFormat: time.RFC3339,
	})
}

func main() {
	// Create pool controller
	poolcontroller, err := poolcontroller.New()
	if err != nil {
		logrus.Panic(err)
	}

	// Create instance controller
	instancecontroller, err := instancecontroller.New()
	if err != nil {
		logrus.Panic(err)
	}

	stopChannel := make(chan struct{}, 0)
	var waitGroup sync.WaitGroup

	// Start pool controller
	go func() {
		waitGroup.Add(1)
		poolcontroller.Run(stopChannel)
		waitGroup.Done()
	}()

	// Start instance controller
	go func() {
		waitGroup.Add(1)
		instancecontroller.Run(stopChannel)
		waitGroup.Done()
	}()

	sigCh := make(chan os.Signal, 0)
	signal.Notify(sigCh, os.Kill, os.Interrupt)

	// Wait for sig kill or sig int and send message to stop
	// pool controller and instance controller.
	<-sigCh
	stopChannel <- *new(struct{})
	stopChannel <- *new(struct{})

	// Wait for contoller to stop
	waitGroup.Wait()
}
