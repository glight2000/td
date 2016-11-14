package main

import (
	"fmt"
	"syscall"
	"unsafe"
	"strings"
	"encoding/json"
	"io/ioutil"
	"time"
	"os"
)

type ulong int32
type ulong_ptr uintptr

type PROCESSENTRY32 struct {
	dwSize              ulong
	cntUsage            ulong
	th32ProcessID       ulong
	th32DefaultHeapID   ulong_ptr
	th32ModuleID        ulong
	cntThreads          ulong
	th32ParentProcessID ulong
	pcPriClassBase      ulong
	dwFlags             ulong
	szExeFile           [260]byte
}

var end = false

type Config struct {
	ProcName               string `json:"ProcName"`
	ProcLocation           string `json:"ProcLocation"`
	ProcCheckInterval      int `json:"ProcCheckInterval"`
	ProcArgs               []string`json:"ProcArgs"`
	UpdateCheckUrl         string `json:"UpdateCheckUrl"`
	UpdateCheckRequestBody interface{} `json:"UpdateCheckRequestBody"`
	UpdateCheckInterval    int `json:"UpdateCheckInterval"`
}

func main() {
	fmt.Println(checkProcExist("go.exe"))

	conf, err := useConf()

	if err != nil {
		panic("Config error")
	}

	for !end {
		if !checkProcExist(conf.ProcName) {
			fmt.Printf("Process %s not exist, start now.\n", conf.ProcName)
			err := startProcess(conf.ProcName, conf.ProcLocation)
			if err != nil {
				fmt.Errorf("Start %s%s error : %s.\n", conf.ProcLocation, conf.ProcName, err.Error())
			}
		} else {
			fmt.Printf("Process %s exists.\n", conf.ProcName)
		}

		fmt.Printf("Will check %d seconds later.\n", conf.ProcCheckInterval)
		time.Sleep(time.Duration(conf.ProcCheckInterval) * 1e9)
	}
}

func useConf() (*Config, error) {
	conf := new(Config)

	f, err := os.Open("conf.json")
	if err != nil {
		fmt.Printf("conf file error : %s \n", err.Error())
		return conf, err
	}

	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Printf("conf file error : %s \n", err.Error())
		return conf, err
	}

	err = json.Unmarshal(data, conf)
	if err != nil {
		fmt.Printf("conf content error : %s \n", err.Error())
		return conf, err
	}
	fmt.Printf("conf : %v\n", conf)

	return conf, nil
}

func checkProcExist(name string) bool {
	exist := false

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	CreateToolhelp32Snapshot := kernel32.NewProc("CreateToolhelp32Snapshot")
	pHandle, _, _ := CreateToolhelp32Snapshot.Call(uintptr(0x2), uintptr(0x0))
	if int(pHandle) == -1 {
		return false;
	}
	Process32Next := kernel32.NewProc("Process32Next")
	for {
		var proc PROCESSENTRY32;
		proc.dwSize = ulong(unsafe.Sizeof(proc))
		if rt, _, _ := Process32Next.Call(uintptr(pHandle), uintptr(unsafe.Pointer(&proc))); int(rt) == 1 {
			procName := strings.Trim(string(proc.szExeFile[0:]), string([]byte{0}))
			//procId := strconv.Itoa(int(proc.th32ProcessID))
			//fmt.Println("ProcessName : " + procName)
			//fmt.Println("ProcessID : " + procId)

			if procName == name {
				exist = true
			}
		} else {
			break;
		}
	}
	CloseHandle := kernel32.NewProc("CloseHandle")
	_, _, _ = CloseHandle.Call(pHandle)

	return exist
}

func startProcess(procName, procLocation string) error {
	execStr := fmt.Sprintf("%s%s", procLocation, procName)
	attr := &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}, //其他变量如果不清楚可以不设定
	}
	_, err := os.StartProcess(execStr, []string{execStr, }, attr) //vim 打开tmp.txt文件
	if err != nil {
		fmt.Println(err)
	}
	return nil
}