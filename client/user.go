package client

import (
	"encoding/json"
	"fmt"
	"time"
	"errors"
)

type User struct {
	ID           int       `json:"id"`
	Login        string    `json:"login"`
	Url          string    `json:"url"`
	Name         string    `json:"name"`
	Company      string    `json:"company"`
	HtmlURL      string    `json:"html_url"`
	Blog         string    `json:"blog"`
	Email        string    `json:"email"`
	Location     string    `json:"location"`
	Bio          string    `json:"bio"`
	Type         string    `json:"type"`
	PublicRepos  int       `json:"public_repos"`
	Hireable     bool      `json:"hireable"`
	CreatedAt    time.Time `json:"created_at"`
}

type Users []User

func (c *Client) CurrentUser() (user *User, err error) {
	if c.authtype == AUTH_NONE {
		return nil,errors.New("Cannot get current user when using AUTH_NONE")
	}
	res,err := c.Get("user")
	if err != nil { return }
	defer res.rawResponse.Body.Close()
	if res.rawResponse.StatusCode != 200 {
		return nil,errors.New("Error getting the current user!")
	}
	dec := json.NewDecoder(res.rawResponse.Body)
	err = dec.Decode(&user)
	return
}

func (c *Client) GetUser(login string) (user *User, err error) {
	res,err := c.Get(fmt.Sprintf("users/%s",login))
	if err != nil { return }
	defer res.rawResponse.Body.Close()
	dec := json.NewDecoder(res.rawResponse.Body)
	err = dec.Decode(&user)
	return
}
