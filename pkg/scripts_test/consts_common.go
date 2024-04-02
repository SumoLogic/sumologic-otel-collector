package sumologic_scripts_tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

const (
	GithubOrg           = "SumoLogic"
	GithubAppRepository = "sumologic-otel-collector"
	GithubApiBaseUrl    = "https://api.github.com"
)

var (
	latestAppVersion string
)

func getAppReleaseVersions() ([]string, error) {
	githubApiBaseUrl, err := url.Parse(GithubApiBaseUrl)
	if err != nil {
		return nil, err
	}
	githubApiReleasesUrl := githubApiBaseUrl.JoinPath(
		"repos",
		GithubOrg,
		GithubAppRepository,
		"releases")
	response, err := http.Get(githubApiReleasesUrl.String()) //nolint:noctx
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get releases: %s", response.Status)
	}

	var releases []struct {
		TagName string `json:"tag_name"`
	}
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&releases)
	if err != nil {
		return nil, err
	}

	releaseVersions := make([]string, len(releases))
	for i, release := range releases {
		releaseVersions[i] = release.TagName
	}
	return releaseVersions, nil
}

func init() {
	releaseVersions, err := getAppReleaseVersions()
	if err != nil {
		fmt.Printf("error fetching releases: %v", err)
		os.Exit(1)
	}
	// the Github API returns these in descending creation order by default
	latestReleaseVersion := releaseVersions[0]
	if latestReleaseVersion == "" {
		fmt.Println("No app release versions found")
		os.Exit(1)
	}
	latestAppVersion = latestReleaseVersion
}
