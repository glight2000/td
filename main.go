package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	log "github.com/Sirupsen/logrus"
	"github.com/kataras/iris"
	"net/http"
	"flag"
)

type Config struct {
	Tasks   []*Task `json:"Tasks"`
	LogFile string `json:"LogFile"`
	Listen  string `json:"Listen"`
	Auth    string `json:"Auth"`
}

func main() {
	confFile := flag.String("c", "", "config file")
	flag.Parse()

	conf, err := useConf(*confFile)
	if err != nil {
		panic("Config error :"+ err.Error())
	}

	log.SetLevel(log.DebugLevel)
	log.Debugf("log test - debug")
	log.Infof("log test - info")
	log.Warnf("log test - warn")
	log.Errorf("log test - error")

	logFile, err := os.OpenFile(conf.LogFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Errorf("Open log file(%s) error : %s", conf.LogFile, err.Error())
	} else {
		log.SetOutput(logFile)
	}

	for _, task := range conf.Tasks {
		go task.ConnectEtcd()
		go task.StartDeamon()
	}

	iris.Get("/", func(c *iris.Context) {
		auth := c.URLParam("Auth")
		log.Debug(auth)
		log.Debug(conf.Auth)
		if auth != conf.Auth {
			c.WriteString("Auth failed!")
			return
		}
		c.JSON(http.StatusOK, conf)
	})
	iris.Listen(conf.Listen)
}

func useConf(file string) (*Config, error) {
	conf := new(Config)

	f, err := os.Open(file)
	if err != nil {
		log.Debugf("conf file error : %s \n", err.Error())
		return conf, err
	}

	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Debugf("conf file error : %s \n", err.Error())
		return conf, err
	}

	err = json.Unmarshal(data, conf)
	if err != nil {
		log.Debugf("conf content error : %s \n", err.Error())
		return conf, err
	}
	log.Debugf("conf : %v\n", conf)

	return conf, nil
}