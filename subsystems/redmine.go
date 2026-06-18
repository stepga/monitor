package subsystems

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/stepga/monitor/bus"
	"github.com/stepga/monitor/config"
	"github.com/stepga/monitor/store"
)

type Redmine struct {
	token string
}

func (r *Redmine) updateIssue() {
	var body []byte
	var err error
	critical := store.FetchCritical()
	if len(critical) == 0 {
		body, err = json.Marshal(map[string]any{
			"issue": map[string]any{
				"status_id":      "3", // Resolved
				"description":    "Nothing to see here",
				"assigned_to_id": "", // Delete assignee
			},
		})
	} else {
		var description []string

		for _, msg := range critical {
			description = append(description, msg.Summary())
		}

		body, err = json.Marshal(map[string]any{
			"issue": map[string]any{
				"status_id":   "1", // New
				"description": strings.Join(description, "\n"),
			},
		})
	}

	if err != nil {
		fmt.Printf("Error: redmine: %s\n", err)
		return
	}

	url := fmt.Sprintf("%s/issues/%s.json?key=%s",
		config.Cfg.Redmine.Url,
		config.Cfg.Redmine.IssueId,
		r.token)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: redmine: %s\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: redmine: %s\n", err)
		return
	}
	defer resp.Body.Close()

	if (resp.StatusCode != http.StatusOK) && (resp.StatusCode != http.StatusNoContent) {
		fmt.Printf("Error: redmine: put returned status %s\n", resp.Status)
		return
	}
}

func (r *Redmine) Init() error {
	r.token = os.Getenv("REDMINE_TOKEN")
	if r.token == "" {
		return fmt.Errorf("REDMINE_TOKEN is required")
	}
	go func() {
		ch := bus.Subscribe()
		defer bus.Unsubscribe(ch)
		for m := range ch {
			switch m.(type) {
			case bus.CriticalListChanged:
				go r.updateIssue()
			}
		}
	}()

	return nil
}
