package internal

import (
	"os"
	"os/signal"
)

func WaitInterruption() {

	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, os.Interrupt)

	<-sigCh

}
