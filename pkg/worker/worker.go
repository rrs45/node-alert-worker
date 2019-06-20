package worker

import (
		"time"
		"bytes"
		"os/exec"
		"os"
	
		log "github.com/sirupsen/logrus"
		"github.com/box-node-alert-worker/workerpb"
		"github.com/box-node-alert-worker/pkg/cache"
		"github.com/box-node-alert-worker/pkg/types"
		"github.com/golang/protobuf/ptypes/timestamp"
	
)

//Work kicksoff Ansible play based on requeste received
func Work(statusCache *cache.StatusCache, workCh <-chan *workerpb.TaskRequest, resultCh chan<- *workerpb.TaskResult, stopCh <-chan os.Signal, maxTasks int, podName string ) {
	limit := make(chan struct{}, maxTasks)

WORKERLOOP:
	for {
		select {
		case task, ok := <-workCh:
			if !ok {
				log.Infof("Worker - Work channel closed")
				break WORKERLOOP
			}
			limit <- struct{}{}	
			go func() {
				cond := task.Node +"_" + task.Condition
				log.Infof("Worker routine - setting %s in status cache", cond)
				statusCache.Set(cond, types.Status{
					Action: task.Node,
					Params: task.Params,
					Timestamp: time.Now(),
				})
				pass := execCmd(task.Node, task.Action, task.Condition)
				ts := timestamp.Timestamp{
					Seconds: time.Now().Unix(),
				}
				if pass {
					resultCh <- &workerpb.TaskResult{
						Node: task.Node,
						Condition: task.Condition,
						Action: task.Action,
						Worker: podName,
						Success: true,
						Timestamp: &ts,
					  }
				} else {
					resultCh <- &workerpb.TaskResult{
						Node: task.Node,
						Condition: task.Condition,
						Action: task.Action,
						Worker: podName,
						Success: false,
						Timestamp: &ts,
					  }
				}

				log.Infof("Worker routine - deleting %s from status cache", cond)
				statusCache.DelItem(cond)
				<-limit
			}()
				
		}
	}
//Checking if all tasks are completed	
for i := 0; i < maxTasks; i++ {
	limit <- struct{}{}	
}
log.Info("Worker - All tasks completed, stopping worker")
close(resultCh)

}

func execCmd(node string, play string, condition string) (bool){
	log.Infof("Worker - Running: %s %s", node, play)
	//args := []string{"-i", node, "plays/"+play, "-e", "@/home/rajsingh/.local/bin/ansible-playbook/raj_pwd.yml", "--vault-password-file", "/home/rajsingh/.local/bin/ansible-playbook/vault_pwd.txt"}
	
	//cmd := exec.Command("ansible-playbook", "/ansible/plays/"+play, "-e", "@/ansible/raj_pwd.yml", "--vault-password-file", "/ansible/vault_pwd.txt")
	cmd := exec.Command("ansible-playbook", "/ansible/plays/"+play, "-e","node="+node, "-e","play="+play)
	//log.Infof("Worker - running ansible with: %s", args)
	var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
	
	err := cmd.Run()
	
	if err != nil {
	  log.Infof("Worker - Err: %s", string(stderr.Bytes()) )
	  log.Infof("Worker - Out: %s", string(stdout.Bytes()) )
	  log.Errorf("Worker - Cannot run command: %v", err) 
	  return false
	}

	log.Infof("Worker - Out: %s", string(stdout.Bytes()) )
	return true
}
