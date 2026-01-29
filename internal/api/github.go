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
	Owner  string
	Repo   string
}

func NewClient() (*SentinelClient, error) {
	// Auto-détection du repo via 'gh repo view'
	outRepo, err := exec.Command("gh", "repo", "view", "--json", "owner,name").Output()
	if err != nil {
		return nil, fmt.Errorf("veuillez vous assurer d'être dans un dépôt GitHub : %w", err)
	}
	
	// Parsing simplifié (on évite une dépendance JSON lourde pour 2 champs)
	parts := strings.Split(strings.ReplaceAll(string(outRepo), "\"", ""), ",")
	var owner, name string
	for _, p := range parts {
		if strings.Contains(p, "login:") { owner = strings.Split(p, ":")[1] }
		if strings.Contains(p, "name:") { name = strings.Split(p, ":")[1] }
	}

	outToken, err := exec.Command("gh", "auth", "token").Output()
	if err != nil {
		return nil, fmt.Errorf("token introuvable. Lancez 'gh auth login' : %w", err)
	}
	token := strings.TrimSpace(string(outToken))

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)

	return &SentinelClient{
		Client: github.NewClient(tc),
		Owner:  strings.TrimSpace(owner),
		Repo:   strings.TrimSpace(name),
	}, nil
}

func (s *SentinelClient) GetFailedWorkflows() ([]*github.WorkflowRun, error) {
	opts := &github.ListWorkflowRunsOptions{
		Status: "failure",
		ListOptions: github.ListOptions{PerPage: 10},
	}
	runs, _, err := s.Client.Actions.ListRepositoryWorkflowRuns(context.Background(), s.Owner, s.Repo, opts)
	return runs.WorkflowRuns, err
}

func (s *SentinelClient) GetJobLogs(runID int64) (string, error) {
	jobs, _, err := s.Client.Actions.ListWorkflowJobs(context.Background(), s.Owner, s.Repo, runID, nil)
	if err != nil { return "", err }

	for _, job := range jobs.Jobs {
		if job.GetConclusion() == "failure" {
			// On récupère les logs de l'étape spécifique en échec (plus précis)
			logs, _, err := s.Client.Actions.GetWorkflowJobLogs(context.Background(), s.Owner, s.Repo, job.GetID(), 1)
			if err != nil { return "", err }
			return logs.String(), nil
		}
	}
	return "", fmt.Errorf("logs introuvables")
}