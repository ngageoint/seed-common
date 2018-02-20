package v2

import (
	"strings"

	"github.com/heroku/docker-registry-client/registry"
	"github.com/ngageoint/seed-common/objects"
	"github.com/ngageoint/seed-common/util"
)

type v2registry struct {
	r        *registry.Registry
	Org      string
	Username string
	Password string
	Print    util.PrintCallback
}

func New(url, org, username, password string) (*v2registry, error) {
	if util.PrintUtil == nil {
		util.InitPrinter(util.PrintErr)
	}

	reg, err := registry.New(url, username, password)
	if reg != nil {
		return &v2registry{r: reg, Org: org, Username: username, Password: password, Print: util.PrintUtil}, err
	}
	return nil, err
}

func (r *v2registry) Name() string {
	return "V2"
}

func (r *v2registry) Ping() error {
	_, err := r.r.Repositories()
	return err
}

func (r *v2registry) Repositories() ([]string, error) {
	return r.r.Repositories()
}

func (r *v2registry) Tags(repository string) ([]string, error) {
	return r.r.Tags(repository)
}

func (r *v2registry) Images() ([]string, error) {
	url := r.r.URL + "/v2/_catalog"
	r.Print("Searching %s for Seed images...\n", url)
	repositories, err := r.r.Repositories()

	var images []string
	for _, repo := range repositories {
		if !strings.HasSuffix(repo, "-seed") {
			continue
		}
		tags, err := r.Tags(repo)
		if err != nil {
			r.Print(err.Error())
			continue
		}
		for _, tag := range tags {
			images = append(images, repo+":"+tag)
		}
	}

	return images, err
}

func (r *v2registry) ImagesWithManifests() ([]objects.Image, error) {
	imageNames, err := r.Images()

	if err != nil {
		return nil, err
	}

	images := []objects.Image{}

	for _, imgstr := range imageNames {
		temp := strings.Split(imgstr, ":")
		if len(temp) != 2 {
			r.Print("ERROR: Invalid seed name: %s. Unable to split into name/tag pair\n", imgstr)
			continue
		}
		manifest, err := r.GetImageManifest(temp[0], temp[1])
		if err != nil {
			//skip images with empty manifests
			r.Print("ERROR: Error reading v2 manifest for %s: %s\n Skipping.\n", imgstr, err.Error())
			continue
		}

		imageStruct := objects.Image{Name: imgstr, Registry: r.r.URL, Org: r.Org, Manifest: manifest}
		images = append(images, imageStruct)
	}

	return images, err
}

func (r *v2registry) GetImageManifest(repoName, tag string) (string, error) {
	manifest := ""
	mv2, err := r.r.ManifestV2(repoName, tag)
	if err == nil {
		resp, err := r.r.DownloadLayer(repoName, mv2.Config.Digest)
		if err == nil {
			manifest, err = objects.GetSeedManifestFromBlob(resp)
		}
	}

	return manifest, err
}
