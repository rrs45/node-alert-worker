package main
import (
	"sync"	
	"flag"
	"os"
	"os/signal"
	"syscall"
	"io"

	log "github.com/sirupsen/logrus"
	
	"github.com/box-node-alert-worker/options"
	"github.com/box-node-alert-worker/workerpb"
	"github.com/box-node-alert-worker/pkg/worker"
	"github.com/box-node-alert-worker/pkg/cache"
)

/*func startHTTPServer(addr string, port string) *http.Server {
	mux := http.NewServeMux()
	srv := &http.Server{Addr: addr + ":" + port, Handler: mux}
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	go func() {
		log.Info("Starting HTTP server for alert-responder")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Could not start http server: %s", err)
		}
	}()
	return srv
}*/


func main() {
//Parse command line options
conf := options.GetConfig()
conf.AddFlags(flag.CommandLine)
flag.Parse()
nawo, err := options.NewConfigFromFile(conf.File)
if err!= nil {
	log.Fatalf("Cannot parse config file: %v", err)
}
options.ValidOrDie(nawo)
logFile, _ := os.OpenFile(nawo.GetString("general.log_file"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
defer logFile.Close()
//Set logrus
log.SetFormatter(&log.JSONFormatter{})
log.SetLevel(log.InfoLevel)
mw := io.MultiWriter(os.Stdout, logFile)
log.SetOutput(mw)

podName := os.Getenv("POD_NAME")

var wg sync.WaitGroup
workCh := make(chan *workerpb.TaskRequest, 3)
resultCh := make(chan *workerpb.TaskResult, 3)
stopCh := make(chan os.Signal)
statusCache := cache.NewStatusCache(nawo.GetString("general.cache_expire_interval")) 
service := worker.NewServer(workCh, statusCache, podName)

signal.Notify(stopCh, syscall.SIGTERM)

wg.Add(3)
//srv := startHTTPServer(nawo.ServerAddress, nawo.ServerPort)
//GRPC server
go func() {
	log.Info("Starting GRPC service for node-alert-worker")
	worker.StartGRPCServer(nawo.GetString("server.address"), nawo.GetString("server.port"), service, stopCh)
	wg.Done()
}()

//Worker
go func() {
	log.Info("Starting worker for node-alert-worker")
	worker.Work(statusCache, workCh, resultCh, stopCh, nawo.GetInt("general.max_parallel_tasks"), podName)
	wg.Done()
}()

//Publisher
go func() {
	log.Info("Starting publisher for node-alert-worker")
	worker.Publish(nawo.GetString("responder.address"), nawo.GetString("responder.port"), resultCh)
	wg.Done()
}()

wg.Wait()
}