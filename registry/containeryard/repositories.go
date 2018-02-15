package containeryard

import (
	"github.com/ngageoint/seed-common/objects"
	"github.com/ngageoint/seed-common/util"
	"strings"
)

type Response struct {
	Results Results
}

type Results struct {
	Community map[string]*Image
	Imports   map[string]*Image
}

type Image struct {
	Author    string
	Compliant bool
	Error     bool
	Labels    map[string]string
	Obsolete  bool
	Pulls     string
	Stars     int
	Tags      map[string]Tag
}

type Tag struct {
	Age     int
	Created string
	Digest  string
	Size    string
}

//Result struct representing JSON result
type Result struct {
	Name string
}

func (registry *ContainerYardRegistry) Repositories() ([]string, error) {
	url := registry.url("/search?q=%s&t=json", "-seed")
	repos := make([]string, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response Response

	err = registry.getContainerYardJson(url, &response)
	if err == nil {
		for repoName, _ := range response.Results.Community {
			repos = append(repos, repoName)
		}
		for repoName, _ := range response.Results.Imports {
			repos = append(repos, repoName)
		}
	}
	return repos, err

}

func (registry *ContainerYardRegistry) Tags(repository string) ([]string, error) {
	url := registry.url("/search?q=%s&t=json", repository)
	registry.Print("Searching %s for Seed images...\n", url)
	tags := make([]string, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response Response

	err = registry.getContainerYardJson(url, &response)
	if err == nil {
		for _, image := range response.Results.Community {
			for tagName, _ := range image.Tags {
				tags = append(tags, tagName)
			}
		}
		for _, image := range response.Results.Imports {
			for tagName, _ := range image.Tags {
				tags = append(tags, tagName)
			}
		}
	}
	return tags, err
}

//Images returns all seed images on the registry
func (registry *ContainerYardRegistry) Images() ([]string, error) {
	images, err := registry.ImagesWithManifests()
	imageStrs := []string{}
	for _, img := range images {
		imageStrs = append(imageStrs, img.Name)
	}
	return imageStrs, err
}

//Images returns all seed images on the registry along with their manifests, if available
func (registry *ContainerYardRegistry) ImagesWithManifests() ([]objects.Image, error) {
	//TODO: Update after container yard generates unique manifests for each tag
	url := registry.url("/search?q=%s&t=json", "-seed")
	repos := make([]objects.Image, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response Response

	err = registry.getContainerYardJson(url, &response)
	if err == nil {
		for repoName, image := range response.Results.Community {
			manifestLabel := ""
			for name, value := range image.Labels {
				if name == "com.ngageoint.seed.manifest" {
					manifestLabel = util.UnescapeManifestLabel(value)
					manifestLabel = manifestLabel[1:len(manifestLabel)-1]
				}
			}
			if manifestLabel == "" {
				continue
			}
			for tagName, _ := range image.Tags {
				manifestLabel, err = registry.GetImageManifest(repoName, tagName)
				imageStr := repoName + ":" + tagName
				img := objects.Image{Name: imageStr, Registry: registry.URL, Org: registry.Org, Manifest: manifestLabel}
				repos = append(repos, img)
			}
		}
		for repoName, image := range response.Results.Imports {
			manifestLabel := ""
			for name, value := range image.Labels {
				if name == "com.ngageoint.seed.manifest" {
					manifestLabel = util.UnescapeManifestLabel(value)
					manifestLabel = manifestLabel[1:len(manifestLabel)-1]
				}
			}
			if manifestLabel == "" {
				continue
			}
			for tagName, _ := range image.Tags {
				manifestLabel, err = registry.GetImageManifest(repoName, tagName)
				imageStr := repoName + ":" + tagName
				img := objects.Image{Name: imageStr, Registry: registry.URL, Org: registry.Org, Manifest: manifestLabel}
				repos = append(repos, img)
			}
		}
	}
	return repos, nil
}

func (registry *ContainerYardRegistry) GetImageManifest(repoName, tag string) (string, error) {
	//remove http(s) prefix for docker pull command
	url := strings.Replace(registry.URL, "http://", "", 1)
	url = strings.Replace(url, "https://", "", 1)
	//username := registry.Username
	//password := registry.Password

	manifest := ""
	digest, err := registry.v2Base.ManifestDigest(repoName, tag)
	if err == nil {
		resp, err := registry.v2Base.DownloadLayer(repoName, digest)
		if err == nil {
			manifest, err = objects.GetSeedManifestFromBlob(resp)
		}
	}

	// falling back to docker pull may result in lots of large pulls for non-seed images that somehow snuck through,
	// always slowing down scans
	/*if err != nil {
		// fallback to docker pull
		registry.Print("ERROR: Could not get seed manifest from v2 API: %s\n", err.Error())
		registry.Print("Falling back to docker pull\n")
		imageName, err := util.DockerPull(image, url, r.Org, username, password)
		if err == nil {
			manifest, err = util.GetSeedManifestFromImage(imageName)
		}
		if err != nil {
			registry.Print("ERROR: Could not get manifest: %s\n", err.Error())
		}
	}*/
	return manifest, err
}
