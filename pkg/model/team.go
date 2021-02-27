package model

type Team struct {
	TeamID string
	Name   string
}

type TeamMemberMap struct {
	TeamID   string
	MemberID string
}

type TeamRepoMap struct {
	TeamID   string
	Registry string
	Repo     string
}
