package worker 
import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"time"
	"os"
	"fmt"
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/box-node-alert-worker/workerpb"
	"github.com/box-node-alert-worker/pkg/cache"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/credentials"
)

//Server struct initializes task service
type Server struct {
	WorkCh chan *workerpb.TaskRequest
	StatusCache  *cache.StatusCache
	PodName string
	MetricsPath string
}

//NewServer initializes task service
func NewServer(workCh chan *workerpb.TaskRequest, statusCache *cache.StatusCache, podName string, metricsPath string) *Server {
	return &Server{
		WorkCh: workCh,
		StatusCache:statusCache,
		PodName: podName,
		MetricsPath: metricsPath,
	}
}

//Task receives new task
func (s *Server) Task(ctx context.Context, req *workerpb.TaskRequest) (*workerpb.TaskAck, error){
p, ok := peer.FromContext(ctx)
if !ok {
    log.Error("GRPC Server - Cannot get peer info")
}
curTime := time.Now().Unix()
metricsData := make([][]byte, 4)

log.Infof("GRPC Server - Received task from %+v, request: %+v", p.Addr,req)
metricsData = append(metricsData, []byte(fmt.Sprintf("put skynet_node_autoremediation.task.received.node %d %s pod=%s", curTime, req.Node, s.PodName)) )
metricsData = append(metricsData, []byte(fmt.Sprintf("put skynet_node_autoremediation.task.received.condition %d %s pod=%s",curTime, req.Condition, s.PodName)) )
metricsData = append(metricsData, []byte(fmt.Sprintf("put skynet_node_autoremediation.task.received.action %d %s pod=%s", curTime, req.Action, s.PodName)) )
metricsData = append(metricsData, []byte(fmt.Sprintf("put skynet_node_autoremediation.task.received.params %d %s pod=%s", curTime, req.Params, s.PodName)) )
for _, metric := range metricsData {
	err := ioutil.WriteFile(s.MetricsPath, metric, 0644) 
	if err!= nil {
		log.Errorf("GRPC Server - Could not write to metrics file: %v", err)
	}
}
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
func StartGRPCServer(addr string, port string, certFile string, keyFile string, caCertFile string, service *Server, stopCh chan os.Signal){
	// Load the certificates from disk
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("GRPC Server - Could not load certificates: %v", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		log.Fatalf("GRPC Server - Could read CA certificates: %v", err)
	}

	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("GRPC Server - Could not append CA certs to pool: %v", err)
	}
	
	srv, err := net.Listen("tcp", fmt.Sprintf("%s:%s",addr,port) )
	if err != nil {
		log.Fatalf("GRPC Server - Failed to start listener: %v", err)
	}
	
	tlsConfig := tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	}
	tlsConfig.BuildNameToCertificate()
	// Create the TLS configuration to pass to the GRPC server
	creds := credentials.NewTLS(&tlsConfig)

	s := grpc.NewServer(grpc.Creds(creds))
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