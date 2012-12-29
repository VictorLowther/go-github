package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Repo struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	FullName     string    `json:"full_name"`
	URL          string    `json:"url"`
	HtmlURL      string    `json:"html_url"`
	CloneURL     string    `json:"clone_url"`
	Owner        User      `json:"owner"`
	MasterBranch string    `json:"master_branch"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	PushedAt     time.Time `json:"pushed_at"`
}

type Repos []Repo

func (c *Client) CurrentUserRepos() (repos Repos, err error) {
	if c.authtype == AUTH_NONE {
		err = errors.New("Cannot get current user repos when using AUTH_NONE")
		return
	}
	res,err := c.Get("user/repos")
	defer res.Response.Body.Close()
	if err != nil { return }
	var r Repos
	var dec *json.Decoder
	for res != nil {
		dec = json.NewDecoder(res.Response.Body)
		err = dec.Decode(&r)
		if err != nil { return nil,err }
		repos = append(repos,r...)
		res = res.NextPage()
	}
	return
}

func (c *Client) UserRepos(login string) (repos Repos, err error) {
	res,err := c.Get(fmt.Sprintf("orgs/%s",login))
	if err != nil {
		res,err = c.Get(fmt.Sprintf("users/%s/repos",login))
	} else {
		res.Response.Body.Close()
		res,err = c.Get(fmt.Sprintf("orgs/%s/repos",login))
	}
	defer res.Response.Body.Close()
	if err != nil { return }
	var r Repos
	var dec *json.Decoder
	for res != nil {
		dec = json.NewDecoder(res.Response.Body)
		err = dec.Decode(&r)
		if err != nil { return nil,err }
		repos = append(repos,r...)
		res = res.NextPage()
	}
	return
}

func (c *Client) GetRepo(login string, reponame string) (repo *Repo, err error) {
	res,err := c.Get(fmt.Sprintf("repos/%s/%s",login,reponame))
	if err != nil { return }
	defer res.Response.Body.Close()
	dec := json.NewDecoder(res.Response.Body)
	err = dec.Decode(&repo)
	return
}