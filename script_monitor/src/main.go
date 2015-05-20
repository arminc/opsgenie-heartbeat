package main

import (
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/arminc/opsgenie-heartbeat/script_monitor/src/opsgenie"
	"github.com/codegangsta/cli"
)

func main() {
	log.SetLevel(log.WarnLevel)
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Version = "1.0"
	app.Usage = "Send hartbeats to OpsGenie"
	app.Author = "OpsGenie"
	app.Flags = opsgenie.SharedFlags
	app.Commands = opsgenie.Commands
	app.Run(os.Args)
}
