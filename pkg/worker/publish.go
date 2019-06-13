package worker

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/box-node-alert-worker/workerpb"
	"google.golang.org/grpc"
)

//Publish publishes the results
func Publish(resultCh <-chan *workerpb.TaskResult) {
	conn, err := grpc.Dial("127.0.0.1:50040", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to connect to worker: %v",err)
		return
	}
	defer conn.Close()
	client := workerpb.NewTaskReceiveServiceClient(conn)
	for {
		select {
		case res := <- resultCh:
			log.Info("Publisher - received ",res)
		
		response, err := client.ResultUpdate(context.Background(), res)
		if err != nil {
			log.Errorf("Unable to send response to scheduler: %v",err)
			continue
		}
		log.Infof("Publisher - sent %s to scheduler", response.Condition)
	}
	}
}