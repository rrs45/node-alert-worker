package worker 
import (
	"net"
	"os"
	"fmt"
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/box-node-alert-worker/workerpb"
	"github.com/box-node-alert-worker/pkg/cache"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

//Server struct initializes task service
type Server struct {
	WorkCh chan *workerpb.TaskRequest
	StatusCache  *cache.StatusCache
	PodName string
}

//NewServer initializes task service
func NewServer(workCh chan *workerpb.TaskRequest, statusCache *cache.StatusCache, podName string) *Server {
	return &Server{
		WorkCh: workCh,
		StatusCache:statusCache,
		PodName: podName,
	}
}

//Task receives new task
func (s *Server) Task(ctx context.Context, req *workerpb.TaskRequest) (*workerpb.TaskAck, error){
p, ok := peer.FromContext(ctx)
if !ok {
    log.Error("GRPC Server - Cannot get peer info")
}
log.Infof("GRPC Server - Received task from %+v, request: %+v", p.Addr,req)
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
		Worker: s.PodName,}
}
return &workerpb.AllTasks{Items: buf,}, nil
}

//StartGRPCServer starts GRPC server
func StartGRPCServer(addr string, port string, service *Server, stopCh chan os.Signal){
	srv, err := net.Listen("tcp", fmt.Sprintf("%s:%s",addr,port) )
	if err != nil {
		log.Fatalf("GRPC Server - Failed to start listener: %v", err)
	}
	
	s := grpc.NewServer()
	workerpb.RegisterTaskServiceServer(s, service)
	workerpb.RegisterTaskStatusServiceServer(s, service)
	
	log.Info("GRPC Server - Starting routine to listen for SIGTERM")
	go func() {
		<- stopCh
		log.Infof("GRPC Server routine- Caught SIGTERM, shutting down GRPC listener")
		close(service.WorkCh)
		s.GracefulStop()
	}()

	log.Info("GRPC Server - Starting Task and TaskStatus service ")
	if err := s.Serve(srv); err != nil {
		log.Fatalf("GRPC Server - Failed to serve: %v", err)
	}
}