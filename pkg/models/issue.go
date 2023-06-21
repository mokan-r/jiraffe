package models

import "time"

type Issue struct {
	ID          int64
	TgMessageID int
	Key         string
	Link        string
	Priority    string
	Summary     string
	Description string
	Campus      string
	Reporter    string
	Assignee    string
	CreatedAt   time.Time
}

func (i *Issue) String() string {
	return "[" + i.Key + "]" + "(" + i.Link + ")" +
		"\n\n🦒*" + i.Summary +
		"*\n\n🦒 `Description`:\n\n" + i.Description +
		"\n\n🦒 `Assignee:     `" + i.Assignee +
		"\n🦒	`Priority:     `" + i.Priority +
		"\n🦒	`Created at:   `" + i.CreatedAt.Format("02.01.2006 15:04 GMT-07")
}
