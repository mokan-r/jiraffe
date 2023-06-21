package jira

import "github.com/andygrunwald/go-jira"

func GetUserName(u *jira.User) string {
	if u == nil {
		return ""
	}
	return u.Name
}

func GetPriorityName(p *jira.Priority) string {
	if p == nil {
		return ""
	}
	return p.Name
}

func GetStatusName(s *jira.Status) string {
	if s == nil {
		return ""
	}
	return s.Name
}

func GetResolutionName(r *jira.Resolution) string {
	if r == nil {
		return ""
	}
	return r.Name
}
