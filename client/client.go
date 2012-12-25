package client

import (
	"errors"
	"os"
	"fmt"
	"bytes"
	"io"
	"net/url"
	"net/http"
	"regexp"
	"strings"
	"strconv"
	"code.google.com/p/go-netrc/netrc"
)

type authTypes int

const (
	AUTH_PASSWORD = iota
	AUTH_NETRC
	AUTH_OAUTH2
	AUTH_NONE
)

var (
	UnknownAuthMethod = errors.New("Unknown authentication method")
	InvalidAuthData = errors.New("Invalid authentication data passed to client.New")
	InvalidEndpoint = errors.New("Invalid Github API endpoint")
)

var clientMaps = make(map[string]http.Client)

const GITHUB_API = "https://api.github.com"

func (t authTypes) AuthOK() (valid bool) {
	if t >= AUTH_PASSWORD && t <= AUTH_OAUTH2 {
		return true
	}
	return false
}

type Client struct {
	endpoint *url.URL
	authtype authTypes
	username string
	token string
	CallsLimit, CallsRemaining int
	httpClient *http.Client
}

func New(endpoint string, authtype authTypes, username string, token string) (client *Client,err error) {
	if endpoint == "" {
		endpoint = GITHUB_API
	}
	endpoint_uri,err := url.Parse(endpoint)
	if err != nil {
		return nil,InvalidEndpoint
	}
	switch authtype {
	case AUTH_NONE:
		if username != "" || token != "" {
			return nil, errors.New("username and token must be empty when using AUTH_NONE")
		}
	case AUTH_OAUTH2:
		fallthrough
	case AUTH_PASSWORD:
		if username == "" || token == "" {
			return nil,InvalidAuthData
		}
	case AUTH_NETRC:

		if machine,err := netrc.FindMachine(os.ExpandEnv("${HOME}/.netrc"),endpoint_uri.Host);  err != nil || machine.Name == "default" {
			return nil,InvalidAuthData
		} else {
			username = machine.Login
			token = machine.Password
		}
	default:
		return nil, UnknownAuthMethod
	}
	client = &Client{endpoint: endpoint_uri, username: username, token: token, authtype: authtype, httpClient: &http.Client{}}
	return client, nil
}

func (c *Client) doRequest(method, path string,body io.Reader) (req *http.Request, err error) {
	if req, err = http.NewRequest(method,path,body); err != nil {
		return
	}
	switch c.authtype {
	case AUTH_NETRC:
		fallthrough
	case AUTH_PASSWORD:
		req.SetBasicAuth(c.username,c.token)
	case AUTH_OAUTH2:
		req.Header.Add("Authorization", "token "+c.token)
	case AUTH_NONE:
		// nothing
	default:
		panic("Cannot happen in makeRequest!")
	}
	return
}

func (c *Client) makeRequest(method, api_path string,body io.Reader) (req *http.Request, err error) {
	req, err = c.doRequest(method,c.endpoint.String()+"/"+api_path,body)
	return
}

type Response struct {
	client *Client
	rawResponse *http.Response
	CallsLimit, CallsRemaining int
	FirstPage, LastPage, NextPage, PrevPage func() *Response
}


func followLink(l string, c *Client) (res *Response) {
	req,err := c.doRequest("GET",l,nil)
	if err != nil {
		return nil
	}
	res,err = c.Do(req)
	if err != nil {
		return nil
	}
	return
}

func newResponse(c *Client, r *http.Response) (res *Response, err error) {
	res = new(Response)
	linkRE, _ := regexp.Compile("<(.*)>; rel=\"(.*)\"")
	res.client = c
	res.rawResponse = r
	if res.CallsLimit,err = strconv.Atoi(r.Header.Get("X-Ratelimit-Limit")); err != nil {
		return nil,errors.New("Cannot find current API call limit!")
	}
	if res.CallsRemaining,err = strconv.Atoi(r.Header.Get("X-Ratelimit-Remaining")); err != nil {
		return nil,errors.New("Cannot find current API calls remaining!")
	}
	c.CallsLimit, c.CallsRemaining = res.CallsLimit, res.CallsRemaining
	res.PrevPage = func() *Response { return nil }
	res.NextPage = func() *Response { return nil }
	res.FirstPage = func() *Response { return nil }
	res.LastPage = func() *Response { return nil }
	links := strings.Split(r.Header.Get("Link"),",")
	for i := range links {
		match := linkRE.FindStringSubmatch(links[i])
		if len(match) == 3 {
			name := match[2]
			link := match[1]
			switch name {
			case "prev":  res.PrevPage = func() *Response {
					return followLink(link,c)
				}
			case "next":  res.NextPage = func() *Response {
					return followLink(link,c)
				}
			case "first": res.FirstPage = func() *Response {
					return followLink(link,c)
				}
			case "last":  res.LastPage = func() *Response {
					return followLink(link,c)
				}
			}
		}
	}
	if r.StatusCode >= 400 {
		defer r.Body.Close()
		body := new(bytes.Buffer)
		if _,err := body.ReadFrom(r.Body); err != nil {
			return nil,err
		}
		return nil,errors.New(body.String())
	}
	return
}

func (c *Client) Do(req *http.Request) (res *Response, err error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return
	}
	res,err = newResponse(c,resp)
	return
}

func (c *Client) Get(apiPath string) (res *Response, err error) {
	req,err := c.makeRequest("GET",apiPath,nil)
	if err != nil {
		return
	}
	res,err = c.Do(req)
	return
}

func (c *Client) Ping() (ok bool) {
	res,err := c.Get("rate_limit")
	if err != nil {
		fmt.Print(err)
		return false
	}
	defer res.rawResponse.Body.Close()
	ok = res.rawResponse.StatusCode == 200
	return
}