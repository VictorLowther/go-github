package main

import (
	"github.com/VictorLowther/go-github/client"
	"fmt"
)

func main() {
	c,err := client.New("",client.AUTH_NETRC,"","")
	if err != nil {
		panic("Could not create new client!")
	} else if c.Ping() {
		fmt.Println("Able to contact Github")
		fmt.Printf("%d/%d API calls left this hour.\n",c.CallsRemaining,c.CallsLimit)
	} else {
		panic("Unable to contact Github")
	}
	user,err := c.CurrentUser()
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("Current user name: %s (%d), created at %s\n",user.Name,user.ID,user.CreatedAt.String())
	}
	repo,err := c.GetRepo("dellcloudedge","crowbar")
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("Repo name: %s (%s)\n",repo.Name, repo.Owner.Login)
	}
	pull,err := c.GetPullRequest("dellcloudedge","barclamp-crowbar",300)
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("PRQ Title: %s target: %s\n",pull.Title,pull.Base.Ref)
	}
}