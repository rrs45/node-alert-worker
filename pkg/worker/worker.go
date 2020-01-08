package worker

import (
		"time"
		"os/exec"
		"bufio"
		"strings"
		"strconv"
		"os"
		"math/rand"
	
		log "github.com/sirupsen/logrus"
		"github.com/box-node-alert-worker/workerpb"
		"github.com/box-node-alert-worker/pkg/cache"
		"github.com/box-node-alert-worker/pkg/types"
		"github.com/golang/protobuf/ptypes/timestamp"
	
)

//Work kicksoff Ansible play based on requeste received
func Work(statusCache *cache.StatusCache, workCh <-chan *workerpb.TaskRequest, resultCh chan<- *workerpb.TaskResult, stopCh <-chan os.Signal, maxTasks int, podName string, scriptsDir string) {
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
			randomID := 100 + rand.Intn(999-100)
			log.Infof("Worker - Starting routine %d", randomID)
			go func(routineID int) {
				cond := task.Node +"_" + task.Condition
				log.Infof("Worker Routine%d - setting %s in status cache", routineID, cond)
				statusCache.Set(cond, types.Status{
					Action: task.Node,
					Params: task.Params,
					Timestamp: time.Now(),
				})
				pass := execCmd(routineID, task, scriptsDir)
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

				log.Infof("Worker Routine%d - deleting %s from status cache", routineID, cond)
				statusCache.DelItem(cond)
				<-limit
			}(randomID)
				
		}
		log.Info("Worker - Coming out of worker loop")
	}
//Checking if all tasks are completed	
for i := 0; i < maxTasks; i++ {
	log.Info("Worker - Checking if no more tasks")
	limit <- struct{}{}	
}
log.Info("Worker - All tasks completed, stopping worker")
close(resultCh)

}

func execCmd(routineID int, task *workerpb.TaskRequest, scriptsDir string) (bool){
	os.Chdir(scriptsDir)
	cmdName := "ansible-playbook"
	cmdArgs := []string{"-i", task.Node+",", task.Action, "-e", "play_name="+task.Action, "-e", "source="+task.Source, "-e", "routineID="+strconv.Itoa(routineID)}
	for _, p:= range strings.Fields(task.Params) {
		cmdArgs = append(cmdArgs, "-e")
		cmdArgs = append(cmdArgs, p)
	}
	cmd := exec.Command(cmdName, cmdArgs...)
	log.Infof("Worker Routine%d - Running: %s %v", routineID, cmdName, cmdArgs)

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Errorf("Worker Routine%d - Error creating StdoutPipe: %v",routineID, err)
		return false
	}
	done := make(chan struct{})
	scanner := bufio.NewScanner(cmdReader)

	go func() {
		for scanner.Scan() {
			log.Infof("Worker Routine%d - %s\n", routineID, scanner.Text())
		}
		done  <- struct{}{}
	}()
	/*var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr*/
	
	err1 := cmd.Start()
	
	if err1 != nil {
	  //log.Infof("Worker Routine%d - Err: %s", string(stderr.Bytes()) )
	  //log.Infof("Worker Routine%d - Out: %s", string(stdout.Bytes()) )
	  log.Errorf("Worker Routine%d - Cannot run command: %v", routineID, err1) 
	  return false
	}

	<- done

	err = cmd.Wait()
	if err != nil {
		log.Errorf("Worker Routine%d - Error waiting for Cmd: %v",routineID, err)
		return false
	}

	return true
}
