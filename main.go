package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/samber/lo"

	"jira-helper/services/gitlab"
	"jira-helper/services/html"
)

func main() {
	command := "none"

	argsWithProg := os.Args
	for i, arg := range argsWithProg {
		if arg == "-command" && i != len(argsWithProg)-1 {
			command = argsWithProg[i+1]
		}
	}
	if command == "mrs" {
		waitingForApprove()
		return
	}
	fmt.Printf("Unknown command: %s", command)
}

func waitingForApprove() {
	authToken := os.Getenv("GITLAB_AUTH_TOKEN")
	gitlabHost := os.Getenv("GITLAB_HOST")

	if authToken == "" {
		fmt.Println("Need to set GITLAB_AUTH_TOKEN in .env")
		return
	}

	gitlabClient := gitlab.NewClient(gitlabHost, authToken)
	if err := gitlabClient.CheckAuth(); err != nil {
		fmt.Printf("Check gitlab auth error: %s", err.Error())
		return
	}

	mrs, err := gitlabClient.WaitingForApprove()
	if err != nil {
		fmt.Printf("Can't get issues: %s", err.Error())
		return
	}
	formattedMRs := lo.Map(mrs, func(mr gitlab.MR, _ int) string {
		return fmt.Sprintf("%d %s %s", mr.ID, mr.Title, mr.WebUrl)
	})
	fmt.Printf(strings.Join(formattedMRs, "\n"))

	f, err := os.Create("./frontend/tmp.html")
	if err != nil {
		fmt.Printf("Can't create html file: %s", err.Error())
		return
	}
	defer f.Close()

	table := html.PrintTable("Waiting for approve", lo.Map(mrs, func(item gitlab.MR, index int) []html.Cell {
		return []html.Cell{
			{Key: "ID", Value: html.Value{Value: item.ID, Link: item.WebUrl}},
			{Key: "Title", Value: html.Value{Value: item.Title}},
			{Key: "Project", Value: html.Value{Value: strings.TrimRight(item.References.Full, item.References.Short)}},
			{Key: "Author", Value: html.Value{Value: item.Author.Name}},
			{Key: "Approved by", Value: html.Value{Value: lo.Map(item.ApprovedBy, func(user gitlab.User, _ int) string {
				return user.Name
			})}},
			{Key: "Approved", Value: html.Value{Value: item.ApprovedByMe, IsCheckbox: true}},
		}
	}))
	_, err = f.WriteString(table)
	if err != nil {
		fmt.Printf("Can't write html file: %s", err.Error())
		return
	}
}
