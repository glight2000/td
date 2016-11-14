# td
Task daemon

Watching tasks status. Restart when the processes are gone.

Keeping get update notification from etcd. And restart the latest version.

Http interface to get the status of tasks. Oh it's ugly now :).

## Installation
`go get -u github.com/glight2000/td`

## Config
```JSON
{
  "Tasks": [
    {
      "TaskName": "testa",
      "TaskSuffix": "exe",
      "TaskPath": "d:/",
      "TaskArgs": [
        "-name","Json",
        "-department","WhiteHouse"
      ],
      "EtcdUrl": "http://127.0.0.1:2379",
      "UpdateCheckUrl": "http://localhost:8111/update"
    },
    {
      "TaskName": "testb",
      "TaskSuffix": "exe",
      "TaskPath": "d:/",
      "TaskArgs": [
        "-name","Json",
        "-department","WhiteHouse"
      ],
      "EtcdUrl": "http://127.0.0.1:2379",
      "UpdateCheckUrl": "http://localhost:8111/update"
    }
  ],
  "LogFile": "",
  "Listen": ":9010",
  "Auth": "abc"
}
```

Mutiple task with same name? Don't.

Set `Listen` with "127.0.0.1:port" to avoid being touch by other machines.

Keep `Auth` blank to request directly otherwise use `http://ip:port/?Auth=[Auth]` to invoke the tasks status.

## Start
`td -c /usr/local/tdconf.json`

## Known issues

This is use for windows and haven't tested on other os.

When there is no TaskSuffix.Oops.
