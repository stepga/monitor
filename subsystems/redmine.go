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

type GetStatus struct {
	Id int `json:"id"`
}

type GetIssue struct {
	Status GetStatus `json:"status"`
}

type GetResponse struct {
	Issue GetIssue `json:"issue"`
}

type PutIssue struct {
	StatusId     int     `json:"status_id"`
	Description  string  `json:"description"`
	AssignedToId *string `json:"assigned_to_id,omitempty"`
}

type PutResponse struct {
	Issue PutIssue `json:"issue"`
}

func (r *Redmine) getIssueStatus(url, issue_id string) (int, error) {
	resp, err := http.Get(fmt.Sprintf("%s/issues/%s.json?key=%s", url, issue_id, r.token))
	if err != nil {
		return 0, fmt.Errorf("failed to get issue: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("get returned status %s", resp.Status)
	}

	reply := &GetResponse{}
	if err := json.NewDecoder(resp.Body).Decode(reply); err != nil {
		return 0, fmt.Errorf("could not parse reply: %s", err)
	}

	return reply.Issue.Status.Id, nil
}

func (r *Redmine) updateIssue(url, issue_id string) {
	critical := store.FetchCritical()
	var reply *PutResponse
	if len(critical) == 0 {
		delete_assignee := ""
		reply = &PutResponse{
			Issue: PutIssue{
				StatusId:     3,
				Description:  "Nothing to see here",
				AssignedToId: &delete_assignee,
			},
		}
	} else {
		// We want to set the issue status to "New" (1), but
		// only if its not "In progress" (2)
		status, err := r.getIssueStatus(url, issue_id)
		if err != nil {
			fmt.Printf("redmine: failed to get issue: %s\n", err)
			return
		}

		if status != 2 {
			// If status isn't In progress (2) update it
			// to New
			status = 1
		}

		var description []string
		for _, msg := range critical {
			description = append(description, msg.Summary())
		}

		reply = &PutResponse{
			Issue: PutIssue{
				StatusId:    status,
				Description: strings.Join(description, "\n"),
			},
		}
	}

	body, err := json.Marshal(reply)
	if err != nil {
		fmt.Printf("redmine: json marshall failed: %s\n", err)
		return
	}

	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/issues/%s.json?key=%s", url, issue_id, r.token),
		bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("redmine: creating request failed: %s\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("redmine: request failed: %s\n", err)
		return
	}
	defer resp.Body.Close()

	if (resp.StatusCode != http.StatusOK) && (resp.StatusCode != http.StatusNoContent) {
		fmt.Printf("redmine: put returned status %s\n", resp.Status)
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
				go r.updateIssue(config.Cfg.Redmine.Url, config.Cfg.Redmine.IssueId)
			}
		}
	}()

	return nil
}
