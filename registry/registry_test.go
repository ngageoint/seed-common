package registry

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ngageoint/seed-common/util"
)

func TestMain(m *testing.M) {
	util.InitPrinter(util.PrintErr)
	//util.StartRegistry()

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
		{"hub.docker.com", "johnptobe", "", "", "[my-job-0.1.0-seed]", ""},
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
