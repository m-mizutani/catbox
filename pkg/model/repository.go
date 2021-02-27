package model

type RepoConfig struct {
	DBBaseRecord
	MainTags     []string `dynamo:"main_tags" json:"main_tags"`
	SlackChannel *string  `dynamo:"slack_channel,omitempty" json:"slack_channel"`
}
