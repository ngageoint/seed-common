package registry

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ngageoint/seed-common/objects"
	"github.com/ngageoint/seed-common/util"
)

func TestMain(m *testing.M) {
	util.InitPrinter(util.PrintErr)

	code := m.Run()

	os.Exit(code)
}

func TestCreateRegistry(t *testing.T) {
	cases := []struct {
		url      string
		org      string
		username string
		password string
		expect   bool
		errStr   string
	}{
		{"hub.docker.com", "johnptobe", "", "", true, ""},
		{"localhost:5000", "", "", "", false, "authentication required"},
		{"localhost:5000", "", "wronguser", "wrongpass", false, "authentication required"},
		{"localhost:5000", "", "testuser", "testpassword", true, ""},
	}

	for _, c := range cases {
		_, err := CreateRegistry(c.url, c.org, c.username, c.password)

		if err != nil && c.expect == true {
			t.Errorf("CreateRegistry returned an error: %v\n", err)
		}
		if c.expect == false && err != nil && !strings.Contains(err.Error(), c.errStr) {
			t.Errorf("CreateRegistry returned an error: %v\n expected %v", err, c.errStr)
		}
		if c.expect == false && err == nil {
			t.Errorf("CreateRegistry did not return an error when one was expected: %v", c.errStr)
		}
	}
}

func TestRepositories(t *testing.T) {
	cases := []struct {
		url      string
		org      string
		username string
		password string
		expect   string
		errStr   string
	}{
		{"hub.docker.com", "johnptobe", "", "", "[my-job-0.1.0-seed my-job-0.1.2-seed extractor-0.1.0-seed addition-job-0.0.1-seed]", ""},
	}

	for _, c := range cases {
		reg, err := CreateRegistry(c.url, c.org, c.username, c.password)

		images, err := reg.Repositories()

		resultStr := fmt.Sprintf("%s", images)

		if err == nil && c.expect != resultStr {
			t.Errorf("Repositories returned %v, expected %v\n", resultStr, c.expect)
		}
		if err != nil && !strings.Contains(err.Error(), c.errStr) {
			t.Errorf("Repositories returned an error: %v\n expected %v", err, c.errStr)
		}
		if err == nil && c.errStr != "" {
			t.Errorf("Repositories did not return an error when one was expected: %v", c.errStr)
		}
	}
}

func TestTags(t *testing.T) {
	cases := []struct {
		url      string
		org      string
		username string
		password string
		repo     string
		expect   string
		errStr   string
	}{
		{"hub.docker.com", "johnptobe", "", "", "my-job-0.1.0-seed", "[latest 0.1.0]", ""},
	}

	for _, c := range cases {
		reg, err := CreateRegistry(c.url, c.org, c.username, c.password)

		tags, err := reg.Tags(c.repo)

		resultStr := fmt.Sprintf("%s", tags)

		if err == nil && c.expect != resultStr {
			t.Errorf("Tags returned %v, expected %v\n", resultStr, c.expect)
		}
		if err != nil && !strings.Contains(err.Error(), c.errStr) {
			t.Errorf("Tags returned an error: %v\n expected %v", err, c.errStr)
		}
		if err == nil && c.errStr != "" {
			t.Errorf("Tags did not return an error when one was expected: %v", c.errStr)
		}
	}
}

func TestImages(t *testing.T) {
	cases := []struct {
		url      string
		org      string
		username string
		password string
		expect   string
		errStr   string
	}{
		{"hub.docker.com", "johnptobe-typo", "", "", "[]", ""},
		{"hub.docker.com", "johnptobe", "", "", "[my-job-0.1.0-seed:latest my-job-0.1.0-seed:0.1.0 my-job-0.1.2-seed:2.0.0 extractor-0.1.0-seed:0.1.0 addition-job-0.0.1-seed:1.0.0]", ""},
	}

	for _, c := range cases {
		reg, err := CreateRegistry(c.url, c.org, c.username, c.password)

		images, err := reg.Images()

		resultStr := fmt.Sprintf("%s", images)

		if err == nil && c.expect != resultStr {
			t.Errorf("Images returned %v, expected %v\n", resultStr, c.expect)
		}
		if err != nil && !strings.Contains(err.Error(), c.errStr) {
			t.Errorf("Images returned an error: %v\n expected %v", err, c.errStr)
		}
		if err == nil && c.errStr != "" {
			t.Errorf("Images did not return an error when one was expected: %v", c.errStr)
		}
	}
}

func TestImagesWithManifests(t *testing.T) {
	cases := []struct {
		url      string
		org      string
		username string
		password string
		expectedNames   string
		expectedOrg string
		expectedReg string
		errStr   string
	}{
		{"hub.docker.com", "johnptobe-typo", "", "", "[]", "johnptobe-typo", "docker.io", ""},
		{"hub.docker.com", "johnptobe", "", "", "[my-job-0.1.0-seed:latest my-job-0.1.0-seed:0.1.0 my-job-0.1.2-seed:2.0.0 extractor-0.1.0-seed:0.1.0 addition-job-0.0.1-seed:1.0.0]", "johnptobe", "docker.io", ""},
	}

	for _, c := range cases {
		reg, err := CreateRegistry(c.url, c.org, c.username, c.password)

		images, err := reg.ImagesWithManifests()
		names := []string{}
		for _, i := range images {
			names = append(names, i.Name)
			seed, err := objects.SeedFromManifestString(i.Manifest)
			if err != nil  {
				t.Errorf("Error parsing seed manifest for %v/%v/%v, %v", c.url, c.org, i.Name, err)
			}
			if !strings.Contains(i.Name, seed.Job.Name) {
				t.Errorf("ImagesWithManifests name: %v does not match up with manifest name: %v\n", i.Name, seed.Job.Name)
			}
			if !strings.Contains(i.Name, seed.Job.JobVersion) {
				t.Errorf("ImagesWithManifests name: %v does not match up with manifest job version: %v\n", i.Name, seed.Job.JobVersion)
			}
		}

		resultStr := fmt.Sprintf("%s", names)

		if err == nil && c.expectedNames != resultStr {
			t.Errorf("ImagesWithManifests returned %v, expected %v\n", resultStr, c.expectedNames)
		}
		if err != nil && !strings.Contains(err.Error(), c.errStr) {
			t.Errorf("ImagesWithManifests returned an error: %v\n expected %v", err, c.errStr)
		}
		if err == nil && c.errStr != "" {
			t.Errorf("ImagesWithManifests did not return an error when one was expected: %v", c.errStr)
		}
	}
}