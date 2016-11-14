package main

import (
	"strings"
	"fmt"
	"strconv"
	"os"
	"crypto/md5"
	"io"
	"encoding/hex"
	"os/exec"
	"syscall"
	"errors"
	"io/ioutil"
)

func FindProcessIdByName(name string) (int, string, error) {
	cmd := exec.Command("tasklist.exe", "/fo", "csv", "/nh")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return -1, "", errors.New(fmt.Sprintf("Exec tasklist for process list error : %s", err.Error()))
	}

	for _, line := range strings.Split(string(out), "\n") {
		infs := strings.Split(line, ",")
		if len(infs) >= 2 {
			pName := strings.Trim(infs[0], "\"")
			pId, _ := strconv.Atoi(strings.Trim(infs[1], "\""))

			if strings.HasPrefix(pName, name) {
				return pId, pName, nil
			}
		}
	}

	return -1, "", errors.New("Process not found.")
}

func ListVersionsInPath(dirPth string, prefix, suffix string) (list []int, err error) {
	list = make([]int, 0)
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}
	for _, fi := range dir {
		if fi.IsDir() {
			continue
		}
		if strings.HasPrefix(fi.Name(), prefix) && strings.HasSuffix(fi.Name(), suffix) {
			verStr := strings.TrimPrefix(fi.Name(), fmt.Sprintf("%s-", prefix))
			verStr = strings.TrimSuffix(verStr, fmt.Sprintf(".%s", suffix))
			ver, err := strconv.Atoi(verStr)
			if err != nil {
				//log.Warnf("File name (%s) is not in style : %s", fi.Name(), err.Error())
			} else {
				list = append(list, ver)
			}
		}
	}
	return list, nil
}

func FileHashMd5(filePath string) (string, error) {
	var md5Str string
	file, err := os.Open(filePath)
	if err != nil {
		return md5Str, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return md5Str, err
	}
	fmt.Println(len(hash.Sum(nil)))
	hashInBytes := hash.Sum(nil)[:16]
	md5Str = hex.EncodeToString(hashInBytes)
	return strings.ToUpper(md5Str), nil
}
