package opsgenie

import (
	"log"
	"time"

	"github.com/codegangsta/cli"
)

const mandatoryFlags = "[apiKey] and [name] are mandatory"
const intervalWrong = "[intervalUnit] can only be one of the following: mintes, hours or days"

//SharedFlags are used to show the main flags for the application
var SharedFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "apiKey, k",
		Value: "",
		Usage: "API key",
	},
	cli.StringFlag{
		Name:  "name, n",
		Value: "",
		Usage: "heartbeat name",
	},
}

var loopFlags = []cli.Flag{
	cli.DurationFlag{
		Name:  "loopInterval, l",
		Value: time.Duration(60 * time.Second),
		Usage: "Loop interval as a duration",
	},
}

var startFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "description, d",
		Value: "",
		Usage: "Heartbeat description",
	},
	cli.IntFlag{
		Name:  "interval, i",
		Value: 10,
		Usage: "Amount of time OpsGenie waits for a send request before creating alert",
	},
	cli.StringFlag{
		Value: "minutes",
		Name:  "intervalUnit, u",
		Usage: "[minutes, hours or days]",
	},
}

//Commands are used to show the commands for the application
var Commands = []cli.Command{
	{
		Name:        "start",
		Usage:       "Adds a new heartbeat and then sends a hartbeat",
		Description: "Adds a new heartbeat to OpsGenie with the configuration from the given flags. If the heartbeat with the name specified in -name exists, updates the heartbeat accordingly and enables it. It also sends a heartbeat message to activate the heartbeat.",
		Flags:       startFlags,
		Action: func(c *cli.Context) {
			startHeartbeatAndSend(extractArgs(c))
		},
	},
	{
		Name:        "startLoop",
		Usage:       "Same as start and sendLoop",
		Description: "Combines start and sendLoop",
		Flags:       append(startFlags, loopFlags...),
		Action: func(c *cli.Context) {
			StartHeartbeatLoop(extractArgs(c))
		},
	},
	{
		Name:        "stop",
		Usage:       "Disables the heartbeat",
		Description: "Disables the heartbeat specified with -name, or deletes it if -delete is true. This can be used to end the heartbeat monitoring that was previously started.",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "delete",
				Usage: "Delete the heartbeat",
			},
		},
		Action: func(c *cli.Context) {
			stopHeartbeat(extractArgs(c))
		},
	},
	{
		Name:        "send",
		Usage:       "Sends a heartbeat",
		Description: "Sends a heartbeat message to reactivate the heartbeat specified with -name.",
		Action: func(c *cli.Context) {
			sendHeartbeat(extractArgs(c))
		},
	},
	{
		Name:        "sendLoop",
		Usage:       "Keep sending",
		Description: "Sends a continouse heartbeat message to reactivate the heartbeat specified with -name.",
		Flags:       loopFlags,
		Action: func(c *cli.Context) {
			sendHeartbeatLoop(extractArgs(c))
		},
	},
}

//OpsArgs contain the application arguments
type OpsArgs struct {
	ApiKey       string
	Name         string
	Description  string
	Interval     int
	IntervalUnit string
	LoopInterval time.Duration
	Delete       bool
}

func extractArgs(c *cli.Context) OpsArgs {
	if c.GlobalString("apiKey") == "" || c.GlobalString("name") == "" {
		logAndExit(mandatoryFlags)
	}
	if c.String("intervalUnit") != "" && (c.String("intervalUnit") == "minutes" || c.String("intervalUnit") == "hours" || c.String("intervalUnit") == "days") != true {
		logAndExit(intervalWrong)
	}
	return OpsArgs{c.GlobalString("apiKey"), c.GlobalString("name"), c.String("description"), c.Int("interval"), c.String("intervalUnit"), c.Duration("loopInterval"), c.Bool("delete")}
}

var logAndExit = func(msg string) {
	log.Fatal(msg)
}
