package main

type cloudWatchEvent struct {
	Account   string         `json:"account"`
	Detail    ecrEventDetail `json:"detail"`
	ID        string         `json:"id"`
	Region    string         `json:"region"`
	Resources []interface{}  `json:"resources"`
	Source    string         `json:"source"`
	Time      string         `json:"time"`
	Version   string         `json:"version"`
}

type ecrEventDetail struct {
	AwsRegion         string                 `json:"awsRegion"`
	EventCategory     string                 `json:"eventCategory"`
	EventID           string                 `json:"eventID"`
	EventName         string                 `json:"eventName"`
	EventSource       string                 `json:"eventSource"`
	EventTime         string                 `json:"eventTime"`
	EventType         string                 `json:"eventType"`
	EventVersion      string                 `json:"eventVersion"`
	ManagementEvent   bool                   `json:"managementEvent"`
	ReadOnly          bool                   `json:"readOnly"`
	RequestID         string                 `json:"requestID"`
	RequestParameters ecrPushImageRequest    `json:"requestParameters"`
	Resources         []ecrPushImageResource `json:"resources"`
	ResponseElements  ecrPushImageResponse   `json:"responseElements"`
	SourceIPAddress   string                 `json:"sourceIPAddress"`
	UserAgent         string                 `json:"userAgent"`
	// ignore
	// UserIdentity      Foo_sub10  `json:"userIdentity"`
}

type ecrPushImageRequest struct {
	ImageManifest          string `json:"imageManifest"`
	ImageManifestMediaType string `json:"imageManifestMediaType"`
	ImageTag               string `json:"imageTag"`
	RegistryID             string `json:"registryId"`
	RepositoryName         string `json:"repositoryName"`
}

type ecrPushImageResource struct {
	Arn       string `json:"ARN"`
	AccountID string `json:"accountId"`
}

type ecrPushImageResponse struct {
	Image ecrPushImage `json:"image"`
}

type ecrPushImage struct {
	ImageID                ecrPushImageID `json:"imageId"`
	ImageManifest          string         `json:"imageManifest"`
	ImageManifestMediaType string         `json:"imageManifestMediaType"`
	RegistryID             string         `json:"registryId"`
	RepositoryName         string         `json:"repositoryName"`
}

type ecrPushImageID struct {
	ImageDigest string `json:"imageDigest"`
	ImageTag    string `json:"imageTag"`
}
