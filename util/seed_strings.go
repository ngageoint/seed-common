package util

import (
	"errors"
	"fmt"
	"strings"
)

func UnescapeManifestLabel(label string) string {
	// un-escape special characters
	seedStr := label
	seedStr = strings.Replace(seedStr, "\\\"", "\"", -1)
	seedStr = strings.Replace(seedStr, "\\\"", "\"", -1) //extra replace to fix extra back slashes added by docker build command
	seedStr = strings.Replace(seedStr, "\\$", "$", -1)
	seedStr = strings.Replace(seedStr, "\\/", "/", -1)
	seedStr = strings.TrimSpace(seedStr)
	seedStr = strings.TrimSuffix(strings.TrimPrefix(seedStr, "'\""), "\"'")

	if seedStr[0] == '"' { //fix quoted string
		seedStr = seedStr[1 : len(seedStr)-1]
	}

	return seedStr
}

//parses a seed image name into an array with the short name, full name, job version and package version
//the full name may include the registry or organization, if any and the short name should match the
//name in the manifest
func ParseSeedImageName(name string) ([]string, error) {
	result := []string{"", "", "", ""}
	parts := strings.Split(name, ":")
	if len(parts) < 2 {
		return result, errors.New("ERROR: No colons in seed image name")
	}
	if len(parts) > 2 {
		return result, errors.New("ERROR: More than one colon in seed image name")
	}

	result[3] = parts[1] //package version

	test := parts[0][len(parts[0])-5:]
	if test != "-seed" {
		msg := fmt.Sprintf("ERROR: Expected -seed, found %v", test)
		return result, errors.New(msg)
	}
	temp := parts[0][0 : len(parts[0])-5]

	pos := strings.LastIndexByte(temp, '-')

	result[2] = temp[pos+1:] //job version

	result[1] = temp[:pos] //full name

	pos = strings.LastIndexByte(result[1], '/')

	result[0] = result[1][pos+1:] //short name

	return result, nil
}
