package worker

import (
	"fmt"
	"math/rand"
	"time"
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/box-node-alert-worker/workerpb"
	"google.golang.org/grpc"
)

//Publish publishes the results
func Publish(addr string, port string, resultCh <-chan *workerpb.TaskResult) {
PUBLISHLOOP:
	for {
		select {
		case res, ok := <- resultCh:
			if !ok {
				log.Infof("Publisher - Results channel closed")
				break PUBLISHLOOP
			}
			log.Info("Publisher - received ",res)
			conn, err := connect(addr,port)
			client := workerpb.NewTaskReceiveServiceClient(conn)
			response, err := client.ResultUpdate(context.Background(), res)
			if err != nil {
				log.Errorf("Publisher - Unable to send response to scheduler: %v",err)
				continue
			}
			conn.Close()
			log.Infof("Publisher - sent %s to scheduler", response.Condition)
		}
	}
	log.Info("Publisher - Stopping")
}

func connect(addr string, port string) (*grpc.ClientConn, error){
	for {		
			conn, err := grpc.Dial(fmt.Sprintf("%s:%s",addr,port), grpc.WithInsecure())
			if err != nil {
				n := rand.Intn(10)
				time.Sleep(time.Duration(n)*time.Second)
				log.Errorf("Publisher - Unable to connect to worker: %v",err)
				log.Info("Publisher - retrying connection to responder")
				continue
			} else {
				return conn, nil
			}
		}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}