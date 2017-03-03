package models

type DeployLog struct {
	Id float64;
	RepoName string;
	TimeStamp string;
}

type Config struct {
	RepoName string `json:"repo_name"`;
	Location string `json:"location"`;
}
