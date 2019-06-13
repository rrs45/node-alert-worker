package worker 
import (
	"net"
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/box-node-alert-worker/workerpb"
	"google.golang.org/grpc"
)

//Server struct initializes task service
type Server struct {
	WorkCh chan *workerpb.TaskRequest
}

//NewServer initializes task service
func NewServer(workCh chan *workerpb.TaskRequest) *Server {
	return &Server{
		WorkCh: workCh,
	}
}

//Task receives new task
func (s *Server) Task(ctx context.Context, req *workerpb.TaskRequest) (*workerpb.TaskAck, error){
log.Infof("Received: %+v", req)
s.WorkCh <- req
return &workerpb.TaskAck {
	Condition: req.GetCondition(),
}, nil
}

//StartGRPCServer starts GRPC server
func StartGRPCServer(addr string, port string, service *Server){
	srv, err := net.Listen("tcp", "127.0.0.1:50050")
	if err != nil {
		log.Fatalf("Failed to start listener: %v", err)
	}
	
	s := grpc.NewServer()
	workerpb.RegisterTaskServiceServer(s, service)
	
	log.Info("Starting Task service ")
	if err := s.Serve(srv); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}