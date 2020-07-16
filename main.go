package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/tkyshm/cslack/slack"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	cslack  = kingpin.New("cslack", "slack cli tool")
	verbose = cslack.Flag("verbose", "Verbose mode").Short('v').Bool()

	version = cslack.Command("version", "display version")

	sendFile   = cslack.Command("file", "send message as file to slack")
	alertLevel = sendFile.Flag("level", "set alert level ('info' 'danger' 'health' 'warn')").Short('l').String()
	username   = sendFile.Flag("username", "username").Short('u').String()
	webhookURL = sendFile.Flag("webhook-url", "webhook url").Required().Short('w').String()
	channel    = sendFile.Flag("channel", "channel name (e.g. '#general')").Required().Short('c').String()

	cslackVer = "v0.0.1"
)

type AlertLevel int

const (
	Danger AlertLevel = iota
	Warn
	Health
	Info
)

var colors = map[AlertLevel]string{
	Danger: "#fc2f2f",
	Warn:   "#ffcc14",
	Health: "#27d871",
	Info:   "#2bb9f2",
}

func main() {
	switch kingpin.MustParse(cslack.Parse(os.Args[1:])) {
	case version.FullCommand():
		fmt.Printf("version: %s\n", cslackVer)
	case sendFile.FullCommand():
		if *verbose {
			log.Println("[info] send message as file.")
		}

		stat, err := os.Stdin.Stat()
		if err != nil {
			log.Printf("[error] %s", err)
			return
		}

		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			log.Println("[error] required stdin")
			return
		}

		in, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Printf("[error] %s", err)
			return
		}

		var user = "cslack"
		if *username != "" {
			user = *username
		}

		var color, text string
		switch *alertLevel {
		case "info":
			color = colors[Info]
		case "warn":
			color = colors[Warn]
			text = "<!here>"
		case "danger":
			color = colors[Danger]
			text = "<!channel>"
		case "health":
			color = colors[Health]
			text = "<!here>"
		default:
			color = colors[Info]
		}

		param := slack.FileParam{
			Text:     text,
			Username: user,
			Channel:  *channel,
			Attachments: []slack.Attachment{
				{
					Color: color,
					Fields: []slack.Field{
						{
							Title: "cslack message",
							Value: string(in),
						},
					},
				},
			},
		}

		if ret, err := slack.PostAsFile(param, *webhookURL); err != nil {
			log.Println("[error]", err)
			log.Println("[error] resp:", string(ret))
		}
	}
}
