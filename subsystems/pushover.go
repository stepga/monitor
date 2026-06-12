package subsystems

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/stepga/monitor/bus"
	"github.com/stepga/monitor/node"
)

type Pushover struct {
	token string
	user  string
}

func (p *Pushover) sendMessage(title, message string) error {
	data := url.Values{
		"token":   {p.token},
		"user":    {p.user},
		"title":   {title},
		"message": {message},
	}

	resp, err := http.PostForm(
		"https://api.pushover.net/1/messages.json",
		data,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pushover returned status %s", resp.Status)
	}

	return nil
}

func (p *Pushover) Init() error {
	p.token = os.Getenv("PUSHOVER_TOKEN")
	if p.token == "" {
		return fmt.Errorf("PUSOVER_TOKEN is required")
	}
	p.user = os.Getenv("PUSHOVER_USER")
	if p.token == "" {
		return fmt.Errorf("PUSOVER_USER is required")
	}
	go func() {
		ch := bus.Subscribe()
		defer bus.Unsubscribe(ch)
		for m := range ch {
			switch msg := m.(type) {
			// Just Placeholder code for now
			case node.NodeInfo:
				go p.sendMessage("NodeInfo", fmt.Sprintf("Node %s reported!", msg.HostName))
			}
		}
	}()

	return nil
}
