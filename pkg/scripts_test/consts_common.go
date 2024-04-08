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

func getLatestAppReleaseVersion() (string, error) {
	githubApiBaseUrl, err := url.Parse(GithubApiBaseUrl)
	if err != nil {
		return "", err
	}
	githubApiLatestReleaseUrl := githubApiBaseUrl.JoinPath(
		"repos",
		GithubOrg,
		GithubAppRepository,
		"releases",
		"latest")
	response, err := http.Get(githubApiLatestReleaseUrl.String()) //nolint:noctx
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get release: %s", response.Status)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&release)
	if err != nil {
		return "", err
	}

	return release.TagName, nil
}

func init() {
	latestReleaseVersion, err := getLatestAppReleaseVersion()
	if err != nil {
		fmt.Printf("error fetching release: %v", err)
		os.Exit(1)
	}
	if latestReleaseVersion == "" {
		fmt.Println("No app release versions found")
		os.Exit(1)
	}
	latestAppVersion = latestReleaseVersion
}
