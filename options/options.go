package options

import (
	"flag"
	"time"
	log "github.com/sirupsen/logrus"
)

//AlertWorkerOptions is struct to gather options for the worker
type AlertWorkerOptions struct {
	ServerAddress string
	ServerPort    string
	APIServerHost string
	LogFile       string

	CacheExpireInterval string
	MaxParallel  int
	Namespace string
}

//NewAlertWorkerOptions returns a flagset
func NewAlertWorkerOptions() *AlertWorkerOptions {
	return &AlertWorkerOptions{}
}

//AddFlags adds options to the flagset
func (awo *AlertWorkerOptions) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&awo.ServerAddress, "address", "127.0.0.1", "Address to bind the alert worker server.")
	fs.StringVar(&awo.ServerPort, "port", "9090", "Port to bind the alert worker server for /healthz endpoint")
	fs.StringVar(&awo.APIServerHost, "apiserver-host", "", "Custom hostname used to connect to Kubernetes ApiServer")
	fs.StringVar(&awo.LogFile, "log-file", "/var/log/service/node-alert-worker.log", "Log file to store all logs")

	fs.IntVar(&awo.MaxParallel, "-max-parallel",3, "Maximum number of remediations that can work in parallel")
	fs.StringVar(&awo.Namespace, "namespace", "node-alert-worker", "Namespace where worker will be deployed")

	fs.StringVar(&awo.CacheExpireInterval, "cache-expire-interval", "10h", "Time period after which cache entries will expire")
}

//ValidOrDie checks some of the options are valid
func (awo *AlertWorkerOptions) ValidOrDie() {
	_, err := time.ParseDuration(awo.CacheExpireInterval)
	if err != nil {
		log.Error("Options - Incorrect cache-expire-interval, sample format: 10s or 1m or 1h; ", err)
		log.Panic("Incorrect options")
	}

}