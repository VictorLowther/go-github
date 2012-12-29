package client

import (
	"encoding/json"
	"fmt"
	"time"
	"errors"
)

type PullRequestBranch struct {
	Label     string `json:"label"`
	Ref       string `json:"ref"`
	SHA       string `json:"sha"`
	Repo      Repo   `json:"repo"`
}

type PullRequest struct {
	URL string `json:"url"`
	HtmlURL string `json:"html_url"`
	Number int `json:"number"`
	State string `json:"state"`
	Title string `json:"title"`
	Body string `json:"body"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	//ClosedAt time.Time `json:"closed_at,omitempty"`
	//MergedAt time.Time `json:"merged_at,omitempty"`
	Head PullRequestBranch `json:"head"`
	Base PullRequestBranch `json:"base"`
}

type PullRequests []PullRequest

func (c *Client) GetPullRequests(login, repo, state string) (pulls PullRequests, err error) {
	var prq_state string
	switch state {
	case "":
		prq_state = "open"
	case "open","closed":
		prq_state = state
	default:
		return nil,errors.New("state must be either open or closed.")
	}
	res,err := c.Get(fmt.Sprintf("repos/%s/%s/pulls?state=%s",login,repo,prq_state))
	if err != nil { return }
	var p PullRequests
	var dec *json.Decoder
	for res != nil {
		dec = json.NewDecoder(res.Response.Body)
		err = dec.Decode(&p)
		res.Response.Body.Close()
		if err != nil { return nil,err }
		pulls = append(pulls,p...)
		res = res.NextPage()
	}
	return
}

func (c *Client) GetPullRequest(login, repo string, id int) (pull *PullRequest, err error) {
	res,err := c.Get(fmt.Sprintf("repos/%s/%s/pulls/%d",login,repo,id))
	if err != nil { return }
	defer res.Response.Body.Close()
	dec := json.NewDecoder(res.Response.Body)
	err = dec.Decode(&pull)
	return
}