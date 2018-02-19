package util

import "strings"

func UnescapeManifestLabel(label string) string {
	// un-escape special characters
	seedStr := label
	seedStr = strings.Replace(seedStr, "\\\"", "\"", -1)
	seedStr = strings.Replace(seedStr, "\\\"", "\"", -1) //extra replace to fix extra back slashes added by docker build command
	seedStr = strings.Replace(seedStr, "\\$", "$", -1)
	seedStr = strings.Replace(seedStr, "\\/", "/", -1)
	seedStr = strings.TrimSpace(seedStr)
	seedStr = strings.TrimSuffix(strings.TrimPrefix(seedStr, "'\""), "\"'")

	if seedStr[0] == '"' {  //fix quoted string
		seedStr = seedStr[1:len(seedStr)-1]
	}

	return seedStr
}