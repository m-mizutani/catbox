package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/golambda"
)

type ImageManifestResult struct {
	Config        imageLayer      `json:"config"`
	Layers        []imageLayer    `json:"layers"`
	MediaType     string          `json:"mediaType"`
	Manifests     []imageManifest `json:"manifests"`
	SchemaVersion int             `json:"schemaVersion"`
}

type imageLayer struct {
	Digest    string `json:"digest"`
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
}

type imageManifest struct {
	Digest    string        `json:"digest"`
	MediaType string        `json:"mediaType"`
	Platform  imagePlatform `json:"platform"`
	Size      int64         `json:"size"`
}

type imagePlatform struct {
	Architecture string `json:"architecture"`
	Os           string `json:"os"`
}

type imageLayerDigests struct {
	digests []string
}

type imageConfigData struct {
	Architecture    string          `json:"architecture"`
	Config          imageConfig     `json:"config"`
	Container       string          `json:"container"`
	ContainerConfig containerConfig `json:"container_config"`
	Created         string          `json:"created"`
	DockerVersion   string          `json:"docker_version"`
	History         []imageHistory  `json:"history"`
	Os              string          `json:"os"`
	Rootfs          imageRootFS     `json:"rootfs"`
}

type containerConfig struct {
	AttachStderr bool        `json:"AttachStderr"`
	AttachStdin  bool        `json:"AttachStdin"`
	AttachStdout bool        `json:"AttachStdout"`
	Cmd          []string    `json:"Cmd"`
	Domainname   string      `json:"Domainname"`
	Entrypoint   interface{} `json:"Entrypoint"`
	Env          []string    `json:"Env"`
	Hostname     string      `json:"Hostname"`
	Image        string      `json:"Image"`
	OnBuild      interface{} `json:"OnBuild"`
	OpenStdin    bool        `json:"OpenStdin"`
	StdinOnce    bool        `json:"StdinOnce"`
	Tty          bool        `json:"Tty"`
	User         string      `json:"User"`
	Volumes      interface{} `json:"Volumes"`
	WorkingDir   string      `json:"WorkingDir"`
}

type imageConfig struct {
	AttachStderr bool        `json:"AttachStderr"`
	AttachStdin  bool        `json:"AttachStdin"`
	AttachStdout bool        `json:"AttachStdout"`
	Cmd          []string    `json:"Cmd"`
	Domainname   string      `json:"Domainname"`
	Entrypoint   interface{} `json:"Entrypoint"`
	Env          []string    `json:"Env"`
	Hostname     string      `json:"Hostname"`
	Image        string      `json:"Image"`
	Labels       interface{} `json:"Labels"`
	OnBuild      interface{} `json:"OnBuild"`
	OpenStdin    bool        `json:"OpenStdin"`
	StdinOnce    bool        `json:"StdinOnce"`
	Tty          bool        `json:"Tty"`
	User         string      `json:"User"`
	Volumes      interface{} `json:"Volumes"`
	WorkingDir   string      `json:"WorkingDir"`
}

type imageHistory struct {
	Created    string `json:"created"`
	CreatedBy  string `json:"created_by"`
	EmptyLayer bool   `json:"empty_layer"`
}

type imageRootFS struct {
	DiffIds []string `json:"diff_ids"`
	Type    string   `json:"type"`
}

// extractRegionFromRegistry parses ECR path and extract region
func extractRegionFromRegistry(registry string) (string, error) {
	// Example: 1111111111.dkr.ecr.ap-northeast-1.amazonaws.com
	ptn := regexp.MustCompile(`^\d+\.dkr\.ecr\.([a-z0-9-]+)\.amazonaws\.com$`)

	matched := ptn.FindStringSubmatch(registry)
	if len(matched) != 2 {
		return "", fmt.Errorf("Invalid registry format: %s", registry)
	}

	return matched[1], nil
}

// extractAccountFromRegistry parses ECR path and extract region
func extractAccountFromRegistry(registry string) (string, error) {
	// Example: 1111111111.dkr.ecr.ap-northeast-1.amazonaws.com
	ptn := regexp.MustCompile(`^(\d+)\.dkr\.ecr\.[a-z0-9-]+\.amazonaws\.com$`)

	matched := ptn.FindStringSubmatch(registry)
	if len(matched) != 2 {
		return "", fmt.Errorf("Invalid registry format: %s", registry)
	}

	return matched[1], nil
}

// GetRegistryAPIToken gets registry access token via ecr.GetAuthorizationToken.
func (x *Controller) GetRegistryAPIToken(registry string) (string, error) {
	ecrRegion, err := extractRegionFromRegistry(registry)
	if err != nil {
		return "", err
	}
	client, err := x.adaptors.NewECR(ecrRegion)
	if err != nil {
		return "", err
	}

	input := &ecr.GetAuthorizationTokenInput{}
	output, err := client.GetAuthorizationToken(input)
	if err != nil {
		return "", golambda.WrapError(err, "Fail to get auth token of registry to fetch image manifest")
	}

	if len(output.AuthorizationData) == 0 {
		return "", golambda.NewError("Fail to get auth token of registry. No output.AuthorizationData").With("output", output)
	}
	if len(output.AuthorizationData) > 1 {
		logger.With("count", len(output.AuthorizationData)).Warn("Too many auth token from ecr.GetAuthorizationToken")
	}

	return aws.StringValue(output.AuthorizationData[0].AuthorizationToken), nil
}

// GetImageManifest returns manifest of a target image
func (x *Controller) GetImageManifest(target *model.Image, authToken string) (*ImageManifestResult, error) {
	reference := target.Digest
	if reference == "" {
		reference = target.Tag
	}

	url := fmt.Sprintf("https://%s/v2/%s/manifests/%s", target.Registry, target.Repo, reference)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, golambda.WrapError(err, "Fail to create a new HTTP request to get docker registry manifests").With("url", url)
	}
	req.Header.Add("Authorization", "Basic "+authToken)
	req.Header.Add("Accept", "*/*")

	resp, err := x.adaptors.HTTP.Do(req)
	if err != nil {
		return nil, golambda.WrapError(err, "Fail to send http request to registry").With("target", target)
	} else if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, golambda.WrapError(err, "Fail to request to registry").With("url", url).With("status", resp.StatusCode).With("body", string(body))
	}

	var manifest ImageManifestResult
	body, err := ioutil.ReadAll(resp.Body) // Do not use json.Decoder for trouble shooting

	if err != nil {
		return nil, golambda.WrapError(err, "Fail to read body from registry")
	}
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, golambda.WrapError(err, "Fail to unmarshal response from registry").With("body", string(body))
	}

	switch manifest.MediaType {
	case "application/vnd.docker.distribution.manifest.v2+json":
		return &manifest, nil

	case "application/vnd.docker.distribution.manifest.list.v2+json":
		for _, m := range manifest.Manifests {
			if m.Platform.Architecture == "amd64" {
				newTarget := target
				newTarget.Tag = ""
				newTarget.Digest = m.Digest

				return x.GetImageManifest(newTarget, authToken)
			}

			logger.With("manifest", m).Warn("Unsupported platform")
		}

		return nil, fmt.Errorf("No available")

	default:
		logger.With("manifest", manifest).Error("Unsupported manifest media type")
		return nil, fmt.Errorf("Unsupported manifest media type: %s", manifest.MediaType)
	}
}

func (x *Controller) GetImageEnv(manifest *ImageManifestResult, target *model.Image, authToken string) ([]string, error) {
	url := fmt.Sprintf("https://%s/v2/%s/blobs/%s", target.Registry, target.Repo, manifest.Config.Digest)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, golambda.WrapError(err, "Fail to create a new HTTP request to get docker registry manifests").With("url", url)
	}
	req.Header.Add("Authorization", "Basic "+authToken)
	req.Header.Add("Accept", manifest.MediaType)

	resp, err := x.adaptors.HTTP.Do(req)
	if err != nil {
		return nil, golambda.WrapError(err, "Fail to send http request to get blobs").With("url", url).With("target", target)
	} else if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, golambda.WrapError(err, "Fail to request to registry").With("url", url).With("status", resp.StatusCode).With("body", body)
	}

	var config imageConfigData
	body, err := ioutil.ReadAll(resp.Body) // Do not use json.Decoder for trouble shooting
	if err != nil {
		return nil, golambda.WrapError(err, "Fail to read body from registry")
	}
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, golambda.WrapError(err, "Fail to unmarshal config from registry").With("body", body)
	}

	return config.Config.Env, nil
}
