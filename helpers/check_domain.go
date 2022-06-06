package helpers

import (
	"strings"
)

func CheckDomain(urlToShorten string, serverURL string) bool {

	if urlToShorten == serverURL {
		return false
	}

	// Catch different cases e.g. localhost:80, http://localhost, https://localhost...
	newURL := strings.Replace(urlToShorten, "http://", "", 1)
	newURL = strings.Replace(urlToShorten, "http://", "", 1)
	newURL = strings.Replace(newURL, "https://", "", 1)
	newURL = strings.Replace(newURL, "www.", "", 1)
	newURL = strings.Split(newURL, "/")[0]

	if newURL == serverURL {
		return false
	}

	return true

}
