package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

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

	sendSnip    = cslack.Command("snip", "send snippet message")
	apiToken    = sendSnip.Flag("token", "api token").Short('t').String()
	snipChannel = sendSnip.Flag("channel", "channel").Short('c').String()
	title       = sendSnip.Flag("title", "snippet title").Default("cslack message").String()
	comment     = sendSnip.Flag("comment", "message comment").String()

	cslackVer = "v0.0.3"
)

type AlertLevel int

const (
	Danger AlertLevel = iota
	Warn
	Health
	Info
)
const slackBotAPIURL = "https://slack.com/api/files.upload"

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
		sendAsFile()
	case sendSnip.FullCommand():
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

		if err := sendSnippet(*apiToken, *snipChannel, *title, *comment, string(in)); err != nil {
			log.Println("[error]", err)
		}
	}
}

func sendAsFile() error {
	if *verbose {
		log.Println("[info] send message as file.")
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		log.Printf("[error] %s", err)
		return err
	}

	if (stat.Mode() & os.ModeNamedPipe) == 0 {
		log.Println("[error] required stdin")
		return errors.New("required stdin")
	}

	in, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Printf("[error] %s", err)
		return err
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
		return err
	}
	return nil
}

func sendSnippet(token, channel, title, comment, content string) error {
	values := url.Values{}
	values.Set("token", token)
	values.Add("channels", channel)
	values.Add("title", title)
	values.Add("initial_comment", comment)
	values.Add("content", content)

	req, err := http.NewRequest(
		"POST",
		slackBotAPIURL,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		out, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		log.Println("error send message:", string(out))
	}

	return err
}
