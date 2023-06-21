package main

import (
	"github.com/mokan-r/jiraffe/internal/jira"
	"github.com/mokan-r/jiraffe/internal/telegram"
	"github.com/mokan-r/jiraffe/pkg/db"
	"log"
)

type application struct {
	jiraClient     *jira.Client
	telegramClient *telegram.Client
	DB             db.DB
}

func (app *application) Serve() {
	for {
		jiraIssues := app.jiraClient.SearchJiraIssues()
		for _, issue := range jiraIssues {
			exists, err := app.DB.IsIssueExists(issue.Key)
			if err != nil {
				log.Println(err)
			}
			if !exists {
				err := app.DB.InsertIssue(issue)
				if err != nil {
					log.Println("Error inserting jira issue")
					continue
				}
				id, err := app.DB.GetCampusID(issue.Campus)
				if err != nil {
					log.Println("Error fetching campus id")
					continue
				}
				err = app.telegramClient.NotifyAboutNewIssue(issue, id)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}
