package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/samber/lo"

	"jira-helper/services/gitlab"
	"jira-helper/services/html"
)

func main() {
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

	mux := http.NewServeMux()
	mux.HandleFunc("/mrs", newHandler(gitlabClient))

	fs := http.FileServer(http.Dir("./frontend"))
	mux.HandleFunc("/", fs.ServeHTTP)

	handler := Log(mux)
	err := http.ListenAndServe(":4444", handler)
	if err != nil {
		panic(err)
	}
}

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path, r.Method)
		next.ServeHTTP(w, r)
	})
}

func newHandler(gitlabClient *gitlab.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		pageContent, err := waitingForApprove(gitlabClient)
		if err != nil {
			panic(err)
		}
		_, _ = w.Write([]byte(pageContent))
	}
}

func waitingForApprove(gitlabClient *gitlab.Client) (string, error) {
	mrs, err := gitlabClient.WaitingForApprove()
	if err != nil {
		return "", err
	}
	formattedMRs := lo.Map(mrs, func(mr gitlab.MR, _ int) string {
		return fmt.Sprintf("%d %s %s", mr.ID, mr.Title, mr.WebUrl)
	})
	fmt.Printf(strings.Join(formattedMRs, "\n"))

	table := html.PrintTable("Merge requests to review", lo.Map(mrs, func(item gitlab.MR, index int) []html.Cell {
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
	return table, nil
}
