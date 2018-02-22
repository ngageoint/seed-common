package registry

import (
	"testing"
	"github.com/ngageoint/seed-common/util"
	"os"
	"strings"
)

func TestMain(m *testing.M) {
	util.InitPrinter(util.PrintErr)
	util.StartRegistry()

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