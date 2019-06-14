package main
import (
	"sync"	
	"flag"

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

//Set logrus
log.SetFormatter(&log.JSONFormatter{})
log.SetLevel(log.InfoLevel)

//Parse command line options
nawo := options.NewAlertWorkerOptions()
nawo.AddFlags(flag.CommandLine)
flag.Parse()
nawo.ValidOrDie()

var wg sync.WaitGroup
workCh := make(chan *workerpb.TaskRequest, 3)
resultCh := make(chan *workerpb.TaskResult, 3)
statusCache := cache.NewStatusCache(nawo.CacheExpireInterval) 

wg.Add(3)
service := worker.NewServer(workCh, statusCache)
//srv := startHTTPServer(nawo.ServerAddress, nawo.ServerPort)
//GRPC server
go func() {
	log.Info("Starting GRPC service for node-alert-worker")
	worker.StartGRPCServer(nawo.ServerAddress, nawo.ServerPort, service)
	wg.Done()
}()

//Worker
go func() {
	log.Info("Starting worker for node-alert-worker")
	worker.Work(statusCache, workCh, resultCh, nawo.MaxParallel)
	wg.Done()
}()

//Publisher
go func() {
	log.Info("Starting publisher for node-alert-worker")
	worker.Publish(resultCh)
	wg.Done()
}()

wg.Wait()
}