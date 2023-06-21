package db

import "github.com/mokan-r/jiraffe/pkg/models"

type DB interface {
	InsertIssue(issue models.Issue) error
	InsertUser(userName string, campusID int64) error
	InsertCampus(name string, topicID int64) error
	GetIssuesKeys() ([]models.Issue, error)
	GetUsers() ([]models.User, error)
	GetUsersInCampus(campusID int64) ([]models.User, error)
	GetCampuses() ([]models.Campus, error)
	DeleteUser(userName string) error
	IsIssueExists(issueKey string) (bool, error)
	GetCampusID(campusName string) (int64, error)
}
