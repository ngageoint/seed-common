package v2

import (
	"errors"
	"os"
	"strings"

	"github.com/heroku/docker-registry-client/registry"
	"github.com/ngageoint/seed-common/objects"
	"github.com/ngageoint/seed-common/util"
)

type v2registry struct {
	r        *registry.Registry
	Hostname string
	Org      string
	Username string
	Password string
	Print    util.PrintCallback
}

func New(url, org, username, password string) (*v2registry, error) {
	if util.PrintUtil == nil {
		util.InitPrinter(util.PrintErr, os.Stderr, os.Stdout)
	}

	reg, err := registry.New(url, username, password)
	if reg != nil {
		host := strings.Replace(url, "https://", "", 1)
		host = strings.Replace(host, "http://", "", 1)
		return &v2registry{r: reg, Hostname: host, Org: org, Username: username, Password: password, Print: util.PrintUtil}, err
	}
	return nil, err
}

func (v2 *v2registry) Name() string {
	return "V2"
}

func (v2 *v2registry) Ping() error {
	err := v2.r.Ping()
	return err
}

func (v2 *v2registry) Repositories() ([]string, error) {
	repositories, err := v2.r.Repositories()
	var repos []string
	for _, repo := range repositories {
		if !strings.HasSuffix(repo, "-seed") {
			continue
		}
		if v2.Org != "" && !strings.HasPrefix(repo, v2.Org + "/") {
			continue
		}
		_, err2 := v2.Tags(repo)
		if err2 != nil {
			continue
		}
		repos = append(repos, repo)
	}
	return repos, err
}

func (v2 *v2registry) Tags(repository string) ([]string, error) {
	return v2.r.Tags(repository)
}

func (v2 *v2registry) Images() ([]string, error) {
	url := v2.r.URL + "/v2/_catalog"
	v2.Print("Searching %s for Seed images...\n", url)
	repositories, err := v2.r.Repositories()

	var images []string
	for _, repo := range repositories {
		if !strings.HasSuffix(repo, "-seed") {
			continue
		}
		if v2.Org != "" && !strings.HasPrefix(repo, v2.Org + "/") {
			continue
		}
		tags, err := v2.Tags(repo)
		if err != nil {
			continue
		}
		for _, tag := range tags {
			images = append(images, repo+":"+tag)
		}
	}

	return images, err
}

func (v2 *v2registry) ImagesWithManifests() ([]objects.Image, error) {
	imageNames, err := v2.Images()
	v2.Print("Images found in V2 Registry %s with Org %s: \n %v", v2.Hostname, v2.Org, imageNames)
	v2.Print("Getting Manifests for %d images in V2 Registry %s with Org %s", len(imageNames), v2.Hostname, v2.Org)

	if err != nil {
		return nil, err
	}

	images := []objects.Image{}

	for _, imgstr := range imageNames {
		v2.Print("Getting manifest for %s", imgstr)
		temp := strings.Split(imgstr, ":")
		if len(temp) != 2 {
			v2.Print("ERROR: Invalid seed name: %s. Unable to split into name/tag pair\n", imgstr)
			continue
		}
		manifest, err := v2.GetImageManifest(temp[0], temp[1])
		if err != nil {
			//skip images with empty manifests
			v2.Print("ERROR: Error reading v2 manifest for %s: %s\n Skipping.\n", imgstr, err.Error())
			continue
		}

		imgOrg := v2.Org
		if imgOrg == "" {
			index := strings.LastIndex("/", imgstr)
			if index > 0 {
				imgOrg = imgstr[:index]
			}
		}
		imageStruct := objects.Image{Name: imgstr, Registry: v2.Hostname, Org: imgOrg, Manifest: manifest}
		images = append(images, imageStruct)
	}

	return images, err
}

func (v2 *v2registry) GetImageManifest(repoName, tag string) (string, error) {
	manifest := ""
	mv2, err := v2.r.ManifestV2(repoName, tag)
	if err == nil {
		resp, err := v2.r.DownloadLayer(repoName, mv2.Config.Digest)
		if err == nil {
			manifest, err = objects.GetSeedManifestFromBlob(resp)
		}
	}

	if err == nil && manifest == "" {
		err = errors.New("Empty seed manifest!")
	}

	return manifest, err
}

func (v2 *v2registry) RemoveImage(repoName, tag string) error {
	digest, err := v2.r.ManifestDigestV2(repoName, tag)

	if err == nil {
		err = v2.r.DeleteManifest(repoName, digest)
	}

	return err
}