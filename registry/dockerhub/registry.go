package dockerhub

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/heroku/docker-registry-client/registry"
	"github.com/ngageoint/seed-common/constants"
	"github.com/ngageoint/seed-common/util"
)

//DockerHubRegistry type representing a Docker Hub registry
type DockerHubRegistry struct {
	URL    string
	Client *http.Client
	Org    string
	v2Base *registry.Registry
	Print  util.PrintCallback
}

//New creates a new docker hub registry from the given URL
func New(registryUrl, org, username, password string) (*DockerHubRegistry, error) {
	if util.PrintUtil == nil {
		util.InitPrinter(util.PrintErr, os.Stderr, os.Stdout)
	}
	url := strings.TrimSuffix(registryUrl, "/")

	reg, _ := registry.New("https://registry-1.docker.io/", username, password)

	registry := &DockerHubRegistry{
		URL:    url,
		Client: &http.Client{},
		Org:    org,
		v2Base: reg,
		Print:  util.PrintUtil,
	}

	return registry, nil
}

func (r *DockerHubRegistry) url(pathTemplate string, args ...interface{}) string {
	pathSuffix := fmt.Sprintf(pathTemplate, args...)
	url := fmt.Sprintf("%s%s", r.URL, pathSuffix)
	return url
}

func (r *DockerHubRegistry) Name() string {
	return "DockerHubRegistry"
}

func (r *DockerHubRegistry) Ping() error {
	url := r.url("/v2/repositories/%s/", constants.DefaultOrg)
	resp, err := r.Client.Get(url)
	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return errors.New(resp.Status)
		}
	}
	return err
}
