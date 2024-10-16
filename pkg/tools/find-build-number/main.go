package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

const (
	otcOrg  = "SumoLogic"
	otcRepo = "sumologic-otel-collector"
)

const (
	workflowIDAnnotation     = "Workflow ID"
	otcVersionAnnotation     = "OTC Version"
	otcSumoVersionAnnotation = "OTC Sumo Version"
)

type Inputs struct {
	Ref         string `yaml:"ref"`
	GithubToken string `yaml:"github_token"`
}

type Outputs struct {
	BuildVersion string `json:"build_version"`
	BuildNumber  string `json:"build_number"`
}

type jobMetadata struct {
	WorkflowID     string
	OTCVersion     string
	OTCSumoVersion string
}

func getJobMetadata(ctx context.Context, client *github.Client, ref string) (jobMetadata, error) {
	var meta jobMetadata

	list, _, err := client.Checks.ListCheckRunsForRef(ctx, otcOrg, otcRepo, ref, nil)
	if err != nil {
		return meta, err
	}

	for _, checkRun := range list.CheckRuns {
		if !strings.HasPrefix(*checkRun.Name, "Trigger Packaging") {
			continue
		}
		annotations, _, err := client.Checks.ListCheckRunAnnotations(ctx, otcOrg, otcRepo, *checkRun.ID, nil)
		if err != nil {
			return meta, err
		}
		for _, anno := range annotations {
			switch *anno.Title {
			case otcVersionAnnotation:
				meta.OTCVersion = *anno.Message
			case otcSumoVersionAnnotation:
				meta.OTCSumoVersion = *anno.Message
			case workflowIDAnnotation:
				meta.WorkflowID = *anno.Message
			}
		}
	}

	return meta, nil
}

func main() {
	var inputs Inputs

	if err := yaml.NewDecoder(os.Stdin).Decode(&inputs); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse STDIN as JSON: %s\n", err)
		fmt.Println(`Expected input: {"github_token": "xxxxxx", "ref": "abcdef"}`)
		os.Exit(1)
	}

	if inputs.GithubToken == "" {
		fmt.Fprintf(os.Stderr, "The github_token JSON attribute must be set\n")
		os.Exit(1)
	}

	if inputs.Ref == "" {
		fmt.Fprintln(os.Stderr, "The ref JSON attribute must be set (commit SHA or tag)")
		os.Exit(1)
	}

	// setup github client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: inputs.GithubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	githubClient := github.NewClient(tc)

	meta, err := getJobMetadata(ctx, githubClient, inputs.Ref)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var workflowID int64
	if _, err := fmt.Sscanf(meta.WorkflowID, "%d", &workflowID); err != nil {
		fmt.Fprintf(os.Stderr, "invalid workflow id: %q: %s\n", meta.WorkflowID, err)
		os.Exit(1)
	}

	run, _, err := githubClient.Actions.GetWorkflowRunByID(ctx, otcOrg, otcRepo, workflowID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	outputs := Outputs{
		BuildVersion: fmt.Sprintf("%s-sumo-%s", meta.OTCVersion, meta.OTCSumoVersion),
		BuildNumber:  fmt.Sprintf("%d", *run.RunNumber),
	}

	outputJSON, err := json.Marshal(&outputs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal outputs JSON: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(string(outputJSON))
}
