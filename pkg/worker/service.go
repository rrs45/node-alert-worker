package worker 
import (
	"net"
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/box-node-alert-worker/workerpb"
	"github.com/box-node-alert-worker/pkg/cache"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

//Server struct initializes task service
type Server struct {
	WorkCh chan *workerpb.TaskRequest
	StatusCache  *cache.StatusCache
}

//NewServer initializes task service
func NewServer(workCh chan *workerpb.TaskRequest, statusCache *cache.StatusCache) *Server {
	return &Server{
		WorkCh: workCh,
		StatusCache:statusCache,
	}
}

//Task receives new task
func (s *Server) Task(ctx context.Context, req *workerpb.TaskRequest) (*workerpb.TaskAck, error){
log.Infof("GRPC Server - Received task request: %+v", req)
s.WorkCh <- req
return &workerpb.TaskAck {
	Condition: req.GetCondition(),
}, nil
}

//GetTaskStatus sends all running tasks
func (s *Server) GetTaskStatus(ctx context.Context, req  *empty.Empty) (*workerpb.AllTasks, error){
log.Infof("GRPC Server - Received GetTaskStatus request: %+v", req)
buf := make(map[string]*workerpb.TaskStatus)
for key, val := range s.StatusCache.GetAll() {
	buf[key] = &workerpb.TaskStatus{
		Action: val.Action, 
		Worker: "Worker-1",}
}
return &workerpb.AllTasks{Items: buf,}, nil
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