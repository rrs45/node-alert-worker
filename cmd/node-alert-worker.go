package main

import (
	"sync"	
	"context"
	"flag"
	"os"
	"net/http"
	"path/filepath"
	"os/signal"
	"syscall"
	"io"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/box-node-alert-worker/options"
	"github.com/box-node-alert-worker/workerpb"
	"github.com/box-node-alert-worker/pkg/worker"
	"github.com/box-node-alert-worker/pkg/cache"
)

func initClient(apiserver string) (*kubernetes.Clientset, error) {

	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".kube", "config")); err == nil {
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err)
		}
		return kubernetes.NewForConfig(config)
	} 
	kubeConfig, err := restclient.InClusterConfig()
	if err != nil {
		panic(err)
	}
	if apiserver != "" {
		kubeConfig.Host = apiserver
		return kubernetes.NewForConfig(kubeConfig)
	}
	return kubernetes.NewForConfig(kubeConfig)
	
}

func startHTTPServer(addr string, port string) *http.Server {
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
}


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

recvMetricsFile, _ := os.Create(nawo.GetString("general.received_metrics_file"))
defer recvMetricsFile.Close()

resMetricsFile, _ := os.Create(nawo.GetString("general.results_metrics_file"))
defer resMetricsFile.Close()

//Set logrus
log.SetFormatter(&log.JSONFormatter{})
log.SetLevel(log.InfoLevel)
mw := io.MultiWriter(os.Stdout, logFile)
log.SetOutput(mw)

podName := os.Getenv("POD_NAME")

srv := startHTTPServer(nawo.GetString("server.address"), nawo.GetString("server.health_check_port"))

// Create an rest client for kube api
log.Info("Calling initClient for node-alert-responder")
clientset, err := initClient(conf.KubeAPIURL)
if err != nil {
	panic(err)
}

var wg sync.WaitGroup
workCh := make(chan *workerpb.TaskRequest, 3)
resultCh := make(chan *workerpb.TaskResult, 3)
stopCh := make(chan os.Signal)
statusCache := cache.NewStatusCache(nawo.GetString("general.cache_expire_interval")) 
service := worker.NewServer(workCh, statusCache, podName, recvMetricsFile)

signal.Notify(stopCh, syscall.SIGTERM)

wg.Add(3)
//srv := startHTTPServer(nawo.ServerAddress, nawo.ServerPort)
//GRPC server
go func() {
	log.Info("Starting GRPC service for node-alert-worker")
	worker.StartGRPCServer(nawo.GetString("server.address"), nawo.GetString("server.port"), nawo.GetString("certs.cert_file"), nawo.GetString("certs.key_file"), nawo.GetString("certs.ca_cert_file"), service, stopCh)
	wg.Done()
}()

//Worker
go func() {
	log.Info("Starting worker for node-alert-worker")
	worker.Work(statusCache, workCh, resultCh, stopCh, nawo.GetInt("general.max_parallel_tasks"), podName, nawo.GetString("scripts.dir"))
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("Could not stop http server: %s", err)
	}
	wg.Done()
}()

//Publisher
go func() {
	log.Info("Starting publisher for node-alert-worker")
	worker.Publish(clientset, nawo.GetString("responder.namespace"), nawo.GetString("responder.port"), nawo.GetString("certs.cert_file"), nawo.GetString("certs.key_file"), nawo.GetString("certs.ca_cert_file"), resultCh, resMetricsFile, conf.ServerName)
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("Could not stop http server: %s", err)
	}
	wg.Done()
}()

wg.Wait()
}