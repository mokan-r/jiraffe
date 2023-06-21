package jira

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/mokan-r/jiraffe/pkg/models"
	"os"
	"time"
)

const (
	LinkedIssueFieldID = "customfield_11600"
)

type priority struct {
	name string
}
type assignee struct {
	name string
}

type transition struct {
	transition string
	priority   priority
	assignee   assignee
}

type Client struct {
	cl         *jira.Client
	CampusList []string
}

func New() (*Client, error) {
	token := os.Getenv("JIRA_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("JIRA_TOKEN environment variable is empty")
	}
	url := os.Getenv("JIRA_URL")
	if url == "" {
		return nil, fmt.Errorf("JIRA_URL environment variable is empty")
	}
	tp := jira.BearerAuthTransport{
		Token: token,
	}

	jiraClient, err := jira.NewClient(tp.Client(), url)
	if err != nil {
		return nil, err
	}
	return &Client{cl: jiraClient}, nil
}

func (c *Client) SearchJiraIssues() []models.Issue {
	location, err := time.LoadLocation("Europe/Moscow")
	issueLinkBase := c.cl.GetBaseURL().Scheme + "://" + c.cl.GetBaseURL().Host + "/browse/"
	if err != nil {
		return nil
	}

	issues := make([]jira.Issue, 0)

	for _, camp := range c.CampusList {
		jql := `project = SUP AND Status = "Open" AND Labels = "` + camp + `"`
		i, _, _ := c.cl.Issue.Search(jql, &jira.SearchOptions{MaxResults: 20})
		issues = append(issues, i...)
	}

	res := make([]models.Issue, 0, len(issues))
	for _, issue := range issues {
		res = append(res, models.Issue{
			Priority:    GetPriorityName(issue.Fields.Priority),
			Key:         issue.Key,
			Link:        issueLinkBase + issue.Key,
			Summary:     issue.Fields.Summary,
			Description: issue.Fields.Description,
			Campus:      c.getCampus(&issue),
			Reporter:    GetUserName(issue.Fields.Reporter),
			Assignee:    GetUserName(issue.Fields.Assignee),
			CreatedAt:   time.Time(issue.Fields.Created).In(location),
		})
	}
	return res
}

func (c *Client) GetIssue(issueID string) (models.Issue, error) {
	issue, _, err := c.cl.Issue.Get(issueID, nil)
	if err != nil {
		return models.Issue{}, err
	}

	location, err := time.LoadLocation("Europe/Moscow")
	issueLinkBase := c.cl.GetBaseURL().Scheme + "://" + c.cl.GetBaseURL().Host + "/browse/"

	return models.Issue{
		Priority:    GetPriorityName(issue.Fields.Priority),
		Key:         issue.Key,
		Link:        issueLinkBase + issue.Key,
		Summary:     issue.Fields.Summary,
		Description: issue.Fields.Description,
		Campus:      c.getCampus(issue),
		Reporter:    GetUserName(issue.Fields.Reporter),
		Assignee:    GetUserName(issue.Fields.Assignee),
		CreatedAt:   time.Time(issue.Fields.Created).In(location),
	}, nil
}

func (c *Client) TransitionIssue(issue *models.Issue) (string, error) {
	transitions, _, err := c.cl.Issue.GetTransitions(issue.Key)
	if err != nil {
		return "", err
	}

	if hasStartProgressTransition(transitions) {
		_, err := c.cl.Issue.DoTransitionWithPayload(issue.Key, transition{
			transition: "11",
			priority:   priority{name: issue.Priority},
			assignee:   assignee{name: issue.Assignee},
		})
		if err != nil {
			return "", err
		}
	} else {
		return fmt.Sprintf("issue %s already in progress", issue.Key), nil
	}

	return `Issue transitioned to "In Progress" successfully`, nil
}

func hasStartProgressTransition(transitions []jira.Transition) bool {
	for _, t := range transitions {
		if t.Name == "Start progress" {
			return true
		}
	}
	return false
}

func (c *Client) getCampus(issue *jira.Issue) (campus string) {
	for _, camp := range c.CampusList {
		for _, l := range issue.Fields.Labels {
			if camp == l {
				campus = camp
			}
		}
	}
	return campus
}
