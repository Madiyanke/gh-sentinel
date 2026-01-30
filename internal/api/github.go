package api

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

type SentinelClient struct {
	Client *github.Client
	Owner, Repo string
}

func NewClient() (*SentinelClient, error) {
	out, err := exec.Command("gh", "repo", "view", "--json", "owner,name", "--jq", ".owner.login + \"/\" + .name").Output()
	if err != nil { return nil, fmt.Errorf("GitHub context not found") }
	parts := strings.Split(strings.TrimSpace(string(out)), "/")
	
	token, _ := exec.Command("gh", "auth", "token").Output()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: strings.TrimSpace(string(token))})
	tc := oauth2.NewClient(context.Background(), ts)
	return &SentinelClient{Client: github.NewClient(tc), Owner: parts[0], Repo: parts[1]}, nil
}

func (s *SentinelClient) GetFailedWorkflows() ([]*github.WorkflowRun, error) {
	opts := &github.ListWorkflowRunsOptions{ListOptions: github.ListOptions{PerPage: 5}}
	runs, _, err := s.Client.Actions.ListRepositoryWorkflowRuns(context.Background(), s.Owner, s.Repo, opts)
	return runs.WorkflowRuns, err
}

func (s *SentinelClient) GetJobLogs(runID int64) (string, error) {
	jobs, _, _ := s.Client.Actions.ListWorkflowJobs(context.Background(), s.Owner, s.Repo, runID, nil)
	for _, job := range jobs.Jobs {
		if job.GetConclusion() == "failure" {
			logs, _, _ := s.Client.Actions.GetWorkflowJobLogs(context.Background(), s.Owner, s.Repo, job.GetID(), 1)
			return logs.String(), nil
		}
	}
	return "No failure logs found.", nil
}

func (s *SentinelClient) GetWorkflowPath(runID int64) string {
	out, _ := exec.Command("gh", "api", fmt.Sprintf("repos/%s/%s/actions/runs/%d", s.Owner, s.Repo, runID), "--jq", ".path").Output()
	return strings.TrimSpace(string(out))
}