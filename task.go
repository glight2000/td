package main

import (
	"fmt"
	"sort"
	"os"
	"strings"
	"strconv"
	"time"
	"net/http"
	"io"
	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"errors"
	"encoding/json"
	"os/exec"
)

type Task struct {
	TaskName       string `json:"TaskName"`
	TaskSuffix     string `json:"TaskSuffix"`
	TaskPath       string `json:"TaskPath"`
	TaskArgs       []string`json:"TaskArgs"`
	EtcdUrl        string `json:"EtcdUrl"`
	UpdateCheckUrl string `json:"UpdateCheckUrl"`

	CurrentVersion int `json:"-"`
	CurrentProcess *os.Process `json:"-"`
}

func (this *Task) StartDeamon() {
	// find exist process
	pId, pName, err := FindProcessIdByName(this.TaskName)
	if err != nil {
		log.Errorf("Find Process start with %s error : %s", this.TaskName, err.Error())
	}

	if strings.Contains(pName, "-") {
		this.CurrentVersion, err = strconv.Atoi(strings.TrimSuffix(strings.Split(pName, "-")[1], fmt.Sprintf(".%s", this.TaskSuffix)))
		if err != nil {
			log.Errorf("Get version from process %s error : %s, will treat as an old one.", pName, err.Error())
		}
		log.Debugf("Found process %d %s(ver %d) with", pId, pName, this.CurrentVersion)
	}

	this.CurrentProcess, err = os.FindProcess(pId)

	if err != nil {
		log.Debugf("Find process by pid %d error : %s", pId, err.Error())

		// start first time
		err := this.Exec()
		if err != nil {
			log.Errorf("Exec error : %s", err.Error())
			this.CurrentProcess = nil
			return
		}
	}

	for true {
		state, err := this.CurrentProcess.Wait()
		if err != nil {
			log.Errorf("Get Process state error : %s", err.Error())
		}
		if state.Exited() {
			//process terminated
			log.Debugf("Process %s(ver %d) terminated. try to start...", this.TaskName, this.CurrentVersion)
			err = this.Exec()
			if err != nil {
				log.Errorf("Exec error : %s", err.Error())
				this.CurrentProcess = nil
				return
			}
		}
	}
}

func (this *Task) ConnectEtcd() error {
	//
	log.Debugf("CheckUpdate for %s", this.TaskName)
	cfg := client.Config{
		Endpoints:               []string{this.EtcdUrl},
		Transport:               client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Errorf(err.Error())
	}

	kapi := client.NewKeysAPI(c)
	watcher := kapi.Watcher("update", &client.WatcherOptions{AfterIndex:0, Recursive:false})
	log.Debugf("Watching etcd(%s) endpoint", this.EtcdUrl)

	for true {
		response, err := watcher.Next(context.Background())
		if err != nil {
			log.Debugf("Etcd watcher error : %s", err.Error())
		} else {
			log.Debugf("Etcd got update with value : %s", response.Node.Value)
			var data struct {
				AppName   string `json:"appname"`
				Version   int `json:"version"`
				Md5Sum    string `json:"md5sum"`
				UpdateUrl string `json:"updateurl"`
			}
			err := json.Unmarshal([]byte(response.Node.Value), &data)
			if err != nil {
				log.Errorf("Update data error : %s", err.Error())
			}
			log.Debugf("Update data : %+v", data)

			if data.AppName != this.TaskName {
				log.Errorf("App name not correct local:%s update:%s)", this.TaskName, data.AppName)
			} else {
				//check version
				log.Debugf("Current version : %d , latest version : %d", this.CurrentVersion, data.Version)
				if data.Version <= this.CurrentVersion {
					log.Debugf("Is updated.")
					continue
				}

				//make new file
				filePath := fmt.Sprintf("%s/%s-%d.%s", this.TaskPath, this.TaskName, data.Version, this.TaskSuffix)
				log.Debugf("Create file (%s)", filePath)
				file, err := os.Create(filePath)
				if err != nil {
					log.Errorf("Create file (%s) error : %s", filePath, err.Error())
					file.Close()
					continue
				}

				//download the latest binary file
				log.Debugf("Start download from %s", data.UpdateUrl)
				resp, err := http.Get(data.UpdateUrl)
				log.Debugf("Status code : %d ", resp.StatusCode)

				io.Copy(file, resp.Body)
				if err != nil {
					log.Errorf("Write file (%s) error : %s", filePath, err.Error())
					file.Close()
					continue
				}
				log.Debugf("Download succed!")

				//check md5 sum
				md5sum, err := FileHashMd5(filePath)
				if err != nil {
					log.Errorf("Get local md5 sum error :%s", err.Error())
					file.Close()
					continue
				}
				if md5sum != data.Md5Sum {
					log.Errorf("Check md5 sum error local:%s  expected:%s", md5sum, data.Md5Sum)
					file.Close()
					continue
				}
				file.Close()

				//kill current process and the daemon will start latest version
				log.Debugf("Stopping the current process...")
				if this.CurrentProcess == nil {
					this.StartDeamon()
				} else {
					this.CurrentProcess.Kill()
				}
			}
		}
	}
	log.Debugf("Deamon(%s) update end. Must something wrong when checking update from etcd server.", this.TaskName)
	return nil
}

func (this *Task) Exec() error {
	//find latest one
	list, err := ListVersionsInPath(this.TaskPath, this.TaskName, this.TaskSuffix)
	if err != nil {
		log.Errorf("Open task path error : %s", err.Error())
		return errors.New(fmt.Sprintf("Find excute(%s) failed", this.TaskName))
	}
	if len(list) == 0 {
		log.Errorf("Task %s not found in directory : %s", this.TaskName, this.TaskPath)
		return errors.New(fmt.Sprintf("Find excute(%s) failed", this.TaskName))
	}

	sort.Ints(list)
	err = errors.New("Not started!")
	retryTimes := 0
	for err != nil {
		tarVer := list[len(list) - retryTimes - 1]
		//exec
		execStr := fmt.Sprintf("%s%s-%d.%s", this.TaskPath, this.TaskName, tarVer, this.TaskSuffix)
		log.Debugf("Will start process : %s", execStr)
		attr := &os.ProcAttr{
			Files: []*os.File{os.Stdin, nil, nil},
		}

		argv := append([]string{execStr}, this.TaskArgs...)
		this.CurrentProcess, err = os.StartProcess(execStr, argv, attr)
		this.CurrentVersion = tarVer

		if err != nil {
			retryTimes = retryTimes + 1
			if retryTimes == len(list) - 1 {
				return errors.New(fmt.Sprintf("All excute(%s) failed", this.TaskName))
			}
		}
	}
	//suc
	return nil
}

/**
A high level excute with exec package
 */
func (this *Task) Exec1() error {
	//find latest one
	list, err := ListVersionsInPath(this.TaskPath, this.TaskName, this.TaskSuffix)
	if err != nil {
		return errors.New(fmt.Sprintf("Open task path error : %s", err.Error()))
	}
	if len(list) == 0 {
		return errors.New(fmt.Sprintf("Task %s not found in directory : %s", this.TaskName, this.TaskPath))
	}

	sort.Ints(list)
	latestVer := list[len(list) - 1]

	//exec
	execStr := fmt.Sprintf("%s%s-%d.%s", this.TaskPath, this.TaskName, latestVer, this.TaskSuffix)
	cmd := exec.Command(execStr, this.TaskArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}
	this.CurrentVersion = latestVer
	this.CurrentProcess = cmd.Process
	cmd.Wait()
	return err
}