package worker

import (
		"time"
		"bytes"
			
		"os/exec"
	
		log "github.com/sirupsen/logrus"
		"github.com/box-node-alert-worker/workerpb"
		"github.com/golang/protobuf/ptypes/timestamp"
	
)

//Work kicksoff Ansible play based on requeste received
func Work(workCh <-chan *workerpb.TaskRequest, resultCh chan<- *workerpb.TaskResult, maxTasks int ) {
	for {
		select {
		case task := <-workCh:
			//args := 
			res, pass := execCmd(task.Node, task.Action, task.Condition)
			if pass {
				resultCh <- res
			}
		}
	}
}

func execCmd(node string, play string, condition string) (*workerpb.TaskResult, bool){
	log.Infof("Worker - Running: %s %s", node, play)
	//args := []string{"-i", node, "plays/"+play, "-e", "@/home/rajsingh/.local/bin/ansible-playbook/raj_pwd.yml", "--vault-password-file", "/home/rajsingh/.local/bin/ansible-playbook/vault_pwd.txt"}
	
	cmd := exec.Command("/home/rajsingh/.local/bin/ansible-playbook", "-i", node+",", "/home/rajsingh/ansible-skynet/plays/"+play, "-e", "@/home/rajsingh/ansible-skynet/raj_pwd.yml", "--vault-password-file", "/home/rajsingh/ansible-skynet/vault_pwd.txt", "-vvv")
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
	if err != nil {
	  log.Infof("Worker - Err: %s", string(stderr.Bytes()) )
	  log.Infof("Worker - Out: %s", string(stdout.Bytes()) )
	  log.Errorf("Worker - Cannot run command: %v", err)
	  
	  return &workerpb.TaskResult{}, false
	} 

	log.Infof("Worker - Out: %s", string(stdout.Bytes()) )
	ts := timestamp.Timestamp{
		Seconds: time.Now().Unix(),
	}	
	return &workerpb.TaskResult{
		Node: node,
		Condition: condition,
		Action: play,
		Worker: "Worker-1",
		Success: true,
		Timestamp: &ts,
	}, true
	
	/*err = cmd.Wait()
	if err != nil {
		log.Errorf("Worker - Failed to execute command: %v", err)
	}*/
	
	}
