package worker

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"
	"fmt"
	"math/rand"
	"time"
	"context"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"github.com/box-node-alert-worker/workerpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

//Publish publishes the results
func Publish(client *kubernetes.Clientset, namespace string, port string, certFile string, keyFile string, caCertFile string, resultCh <-chan *workerpb.TaskResult, metricsFile *os.File, tlsName string) {
// Load the certificates from disk
certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
if err != nil {
	log.Fatalf("Publisher - Could not load certificates: %v", err)
}

// Create a certificate pool from the certificate authority
certPool := x509.NewCertPool()
ca, err := ioutil.ReadFile(caCertFile)
if err != nil {
	log.Fatalf("Publisher - Could read CA certificates: %v", err)
}

// Append the client certificates from the CA
if ok := certPool.AppendCertsFromPEM(ca); !ok {
	log.Fatalf("Publisher - Could not append CA certs to pool: %v", err)
}

// Create the TLS credentials for transport
creds := credentials.NewTLS(&tls.Config{
	ServerName: tlsName,
	Certificates: []tls.Certificate{certificate},
	RootCAs:      certPool,
})
	
PUBLISHLOOP:
	for {
		select {
		case res, ok := <- resultCh:
			if !ok {
				log.Infof("Publisher - Results channel closed")
				break PUBLISHLOOP
			}
			log.Info("Publisher - received ",res)
			curTime := time.Now().Unix()
			successInt := int(0)
			if  res.Success {
				successInt = 1
			}
		
			metricsData := []byte(fmt.Sprintf("put skynet_node_autoremediation.task.result %d %d node=%s condition=%s action=%s", curTime, successInt, res.Node, res.Condition, res.Action))
			//err := ioutil.WriteFile(metricsFile, metricsData, 0644) 
			_, err := metricsFile.Write(metricsData)
			if err!= nil {
				log.Errorf("GRPC Server - Could not write to metrics file: %v", err)
			}
			
			conn, err := connect(client, namespace, port, creds)
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

func connect(client *kubernetes.Clientset, namespace string, port string, creds credentials.TransportCredentials) (*grpc.ClientConn, error){
	
	for {	
			podList, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
			if err != nil {
				log.Errorf("Publisher - Could not list responder pod: %v", err)
				n := rand.Intn(10)
				log.Infof("Publisher - retrying after %v seconds", n)
				time.Sleep(time.Duration(n)*time.Second)
				continue
			}	
			addr := podList.Items[0].Status.PodIP
			conn, err := grpc.Dial(fmt.Sprintf("%s:%s",addr,port),  grpc.WithTransportCredentials(creds))
			if err != nil {
				log.Errorf("Publisher - Unable to connect to worker: %v",err)
				n := rand.Intn(10)
				log.Infof("Publisher - retrying after %v seconds", n)
				time.Sleep(time.Duration(n)*time.Second)
				continue
			} else {
				return conn, nil
			}
		}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}