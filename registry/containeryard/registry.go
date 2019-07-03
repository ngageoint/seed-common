package containeryard

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/heroku/docker-registry-client/registry"
	"github.com/ngageoint/seed-common/util"
)

//ContainerYardRegistry type representing a Container Yard registry
type ContainerYardRegistry struct {
	URL      string
	Hostname string
	Client   *http.Client
	Org      string
	Username string
	Password string
	v2Base   *registry.Registry
	Print    util.PrintCallback
}

func (r *ContainerYardRegistry) Name() string {
	return "ContainerYardRegistry"
}

//New creates a new docker hub registry from the given URL
func New(registryUrl, org, username, password string) (*ContainerYardRegistry, error) {
	if util.PrintUtil == nil {
		util.InitPrinter(util.PrintErr, os.Stderr, os.Stdout)
	}
	url := strings.TrimSuffix(registryUrl, "/")
	reg, err := registry.New(url, username, password)

	host := strings.Replace(url, "https://", "", 1)
	host = strings.Replace(host, "http://", "", 1)

	registry := &ContainerYardRegistry{
		URL:      url,
		Hostname: host,
		Client:   &http.Client{},
		Org:      org,
		Username: username,
		Password: password,
		v2Base:   reg,
		Print:    util.PrintUtil,
	}

	return registry, err
}

func (r *ContainerYardRegistry) url(pathTemplate string, args ...interface{}) string {
	pathSuffix := fmt.Sprintf(pathTemplate, args...)
	url := fmt.Sprintf("%s%s", r.URL, pathSuffix)
	return url
}

func (r *ContainerYardRegistry) Ping() error {
	//query that should quickly return an empty json response
	url := r.url("/search?q=NoImagesWithThisName&t=json")
	var response Response
	err := r.getContainerYardJson(url, &response)
	return err
}
