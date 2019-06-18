package worker

import (
		"time"
		"bytes"
		"os/exec"
	
		log "github.com/sirupsen/logrus"
		"github.com/box-node-alert-worker/workerpb"
		"github.com/box-node-alert-worker/pkg/cache"
		"github.com/box-node-alert-worker/pkg/types"
		"github.com/golang/protobuf/ptypes/timestamp"
	
)

//Work kicksoff Ansible play based on requeste received
func Work(statusCache *cache.StatusCache, workCh <-chan *workerpb.TaskRequest, resultCh chan<- *workerpb.TaskResult, maxTasks int ) {
	limit := make(chan struct{}, maxTasks)
	for {
		select {
		case task := <-workCh:
			limit <- struct{}{}	
			go func() {
			cond := task.Node +"_" + task.Condition
			log.Infof("Worker routine - setting %s in status cache", cond)
			statusCache.Set(cond, types.Status{
				Action: task.Node,
				Params: task.Params,
				Timestamp: time.Now(),
			})
			res := execCmd(task.Node, task.Action, task.Condition)
			resultCh <- res
			log.Infof("Worker routine - deleting %s from status cache", cond)
			statusCache.DelItem(cond)
			<-limit
		}()
		}
	}
}

func execCmd(node string, play string, condition string) (*workerpb.TaskResult){
	log.Infof("Worker - Running: %s %s", node, play)
	//args := []string{"-i", node, "plays/"+play, "-e", "@/home/rajsingh/.local/bin/ansible-playbook/raj_pwd.yml", "--vault-password-file", "/home/rajsingh/.local/bin/ansible-playbook/vault_pwd.txt"}
	
	//cmd := exec.Command("ansible-playbook", "/ansible/plays/"+play, "-e", "@/ansible/raj_pwd.yml", "--vault-password-file", "/ansible/vault_pwd.txt")
	cmd := exec.Command("ansible-playbook", "/ansible/plays/"+play, "-e","node="+node, "-e","play="+play)
	//log.Infof("Worker - running ansible with: %s", args)
	var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
	/*cmdReader, err := cmd.StdoutPipe()
	if err != nil {
	  return err
	} 
	
	scanner := bufio.NewScanner(cmdReader)
	go func() {
	  for scanner.Scan() {
		log.Infof("Worker - %s", scanner.Text())
	  }
	}() */
	
	err := cmd.Run()
	ts := timestamp.Timestamp{
		Seconds: time.Now().Unix(),
	}
	if err != nil {
	  log.Infof("Worker - Err: %s", string(stderr.Bytes()) )
	  log.Infof("Worker - Out: %s", string(stdout.Bytes()) )
	  log.Errorf("Worker - Cannot run command: %v", err)
	  
	  return &workerpb.TaskResult{
		Node: node,
		Condition: condition,
		Action: play,
		Worker: "Worker-1",
		Success: false,
		Timestamp: &ts,
	  }
	} 

	log.Infof("Worker - Out: %s", string(stdout.Bytes()) )
		
	return &workerpb.TaskResult{
		Node: node,
		Condition: condition,
		Action: play,
		Worker: "Worker-1",
		Success: true,
		Timestamp: &ts,
	}
	
	/*err = cmd.Wait()
	if err != nil {
		log.Errorf("Worker - Failed to execute command: %v", err)
	}*/
	
	}
