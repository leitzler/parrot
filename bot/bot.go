package bot

import (
	"encoding/json"
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"strings"
)

const configFile string = "config.json"

// Bot representation
type Bot struct {
	slackAPI *slack.Client
	config   *Config
}

// New creates a new Bot
func New(api *slack.Client) *Bot {
	newBot := &Bot{
		slackAPI: api,
		config:   NewConfig(),
	}
	if err := newBot.config.fromJSON(configFile); err != nil {
		log.Fatal("Failed to load configuration: ", err)
	}
	newBot.slackAPI.SetDebug(newBot.config.Debug)
	return newBot
}

// HandleMessage routes all incoming messages
func (b *Bot) HandleMessage(event *slack.MessageEvent) {
	if b.config.Debug == true {
		log.Printf("%#v\n", event)
	}

	if event.BotID != "" || event.User == "" || event.SubType != "" {
		return
	}

	// Direct message always starts with "D" and channels with "C" according to Slack support
	if strings.HasPrefix(event.Channel, "D") {
		b.handlePrivateMessage(event)
	} else if strings.HasPrefix(event.Channel, "C") {
		b.handleChannelMessage(event)
	}
}

func (b *Bot) handleChannelMessage(event *slack.MessageEvent) {
	// Require triggers to start with @
	if !strings.HasPrefix(event.Text, "@") {
		return
	}

	messageTrigger := strings.Fields(event.Text[1:])[0] // Strip @ and take first word

	failed := []string{}
	notified := []string{}

	for _, receiver := range b.config.Notifiers[messageTrigger] {
		if err := b.shareTo(event, receiver, event.Text); err != nil {
			failed = append(failed, receiver)
			log.Printf("Failed to send notification to: %v (%v)", receiver, err)
		} else {
			notified = append(notified, receiver)
		}
	}

	if len(failed) > 0 {
		b.replyInPrivate(event, fmt.Sprintf("Failed to send notification to %v users!", len(failed)))
		b.react(event, "warning")
	}
	if len(notified) > 0 {
		statusReply := fmt.Sprint("Notified: ", strings.Join(uidsAsLinks(notified), ", "))
		b.replyInPrivate(event, statusReply)
		b.react(event, "mega")
	}
}

func (b *Bot) handlePrivateMessage(event *slack.MessageEvent) {
	// Private messages allowed by all
	if event.Text == "list" {
		for triggerWord, receivers := range b.config.Notifiers {
			b.replyInPrivate(event, fmt.Sprintf("@%v => %v", triggerWord, strings.Join(uidsAsLinks(receivers), ", ")))
		}
	} else if event.User != b.config.Admin {
		b.replyInPrivate(event, "I currently only understand `list` that lists all notification groups..")
		return
	}

	// Private messages that require elevated access
	if strings.HasPrefix(event.Text, "set") {
		fields := strings.Fields(event.Text)
		if len(fields) < 3 {
			return
		}
		triggerWord := fields[1]
		receivers := parseReceivers(fields[2:])

		b.config.Notifiers[triggerWord] = receivers
		b.replyInPrivate(event, fmt.Sprintf("New trigger word: %v\nReceivers: %v\n", triggerWord, uidsAsLinks(receivers)))
	} else if strings.HasPrefix(event.Text, "del") {
		fields := strings.Fields(event.Text)
		if len(fields) != 2 {
			return
		}
		delete(b.config.Notifiers, fields[1])
		b.react(event, "white_check_mark")
	} else if event.Text == "debug" {
		b.config.Debug = !b.config.Debug
		b.slackAPI.SetDebug(b.config.Debug)
		b.react(event, "white_check_mark")
	} else if event.Text == "save" {
		if err := b.config.toJSON(configFile); err != nil {
			b.replyInPrivate(event, err.Error())
		} else {
			b.react(event, "white_check_mark")
		}
	}
}

func (b *Bot) shareTo(event *slack.MessageEvent, receiver string, text string) error {
	posterInfo, err := b.slackAPI.GetUserInfo(event.User)
	if err != nil {
		log.Println("Failed to fetch user info: ", err)
		return err
	}
	params := slack.PostMessageParameters{AsUser: true}

	attachment := slack.Attachment{
		AuthorIcon:    posterInfo.Profile.Image48,
		AuthorName:    posterInfo.Name,
		AuthorSubname: posterInfo.Profile.RealName,
		Fallback:      fmt.Sprintf("%v: %v", posterInfo.Name, event.Text),
		Footer:        fmt.Sprintf("Posted in <#%v>", event.Channel),
		MarkdownIn:    []string{"text"},
		Text:          text,
		Ts:            json.Number(event.Timestamp),
	}
	params.Attachments = []slack.Attachment{attachment}

	_, _, err = b.slackAPI.PostMessage(receiver, "", params)
	return err
}

func (b *Bot) reply(event *slack.MessageEvent, message string) error {
	_, _, err := b.slackAPI.PostMessage(event.Channel, message, slack.PostMessageParameters{AsUser: true})
	return err
}
func (b *Bot) replyInPrivate(event *slack.MessageEvent, message string) error {
	_, _, err := b.slackAPI.PostMessage(event.User, message, slack.PostMessageParameters{AsUser: true})
	return err
}

func (b *Bot) react(event *slack.MessageEvent, reaction string) error {
	err := b.slackAPI.AddReaction(reaction, slack.ItemRef{
		Channel:   event.Channel,
		Timestamp: event.Timestamp,
	})
	if err != nil {
		log.Println(err)
	}
	return err
}

func uidsAsLinks(uids []string) []string {
	links := make([]string, len(uids))
	for i := range uids {
		links[i] = fmt.Sprintf("<@%v>", uids[i])
	}
	return links
}

func parseReceivers(users []string) []string {
	validReceivers := []string{}
	for _, user := range users {
		if strings.HasPrefix(user, "<@U") && strings.HasSuffix(user, ">") {
			validReceivers = append(validReceivers, user[2:len(user)-1])
		}
	}
	return validReceivers
}
