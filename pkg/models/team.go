package models

type Team struct {
	DBBaseRecord
	TeamID string `dynamo:"team_id" json:"team_id"`
	Name   string `dynamo:"name" json:"name"`
}

type TeamMemberMap struct {
	DBBaseRecord
	TeamID   string `dynamo:"team_id" json:"team_id"`
	MemberID string `dynamo:"member_id" json:"member_id"`
}

type TeamRepoMap struct {
	DBBaseRecord
	TeamID   string `dynamo:"team_id" json:"team_id"`
	Registry string `dynamo:"registry" json:"registry"`
	Repo     string `dynamo:"repo" json:"repo"`
}
