package worker

import (
	log "github.com/sirupsen/logrus"
	"github.com/box-node-alert-worker/workerpb"
)

//Publish publishes the results
func Publish(resultCh <-chan *workerpb.TaskResult) {

	for {
		select {
		case res := <- resultCh:
			log.Info("Publisher - received ",res)
		}
	}
}