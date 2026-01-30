package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gh-sentinel/internal/config"
	sentinelContext "gh-sentinel/internal/context"
	"gh-sentinel/internal/errors"
	"gh-sentinel/internal/logger"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

// Client wraps GitHub API client with enhanced functionality
type Client struct {
	client  *github.Client
	repo    *sentinelContext.RepoContext
	config  *config.Config
	logger  *logger.Logger
	ctx     context.Context
}

// NewClient creates a new GitHub client with automatic authentication
func NewClient(cfg *config.Config, log *logger.Logger) (*Client, error) {
	// Check authentication
	if err := sentinelContext.CheckAuthentication(); err != nil {
		return nil, err
	}

	// Detect repository context
	repo, err := sentinelContext.DetectRepository()
	if err != nil {
		return nil, err
	}

	// Get auth token
	token, err := sentinelContext.GetAuthToken()
	if err != nil {
		return nil, err
	}

	// Create authenticated client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	
	ghClient := github.NewClient(tc)
	ghClient.UserAgent = cfg.UserAgent

	log.Info("Authenticated as repository: %s", repo.FullName)

	return &Client{
		client: ghClient,
		repo:   repo,
		config: cfg,
		logger: log,
		ctx:    ctx,
	}, nil
}

// GetRepository returns the repository context
func (c *Client) GetRepository() *sentinelContext.RepoContext {
	return c.repo
}

// WorkflowRun represents a simplified workflow run
type WorkflowRun struct {
	ID          int64
	Name        string
	DisplayTitle string
	Status      string
	Conclusion  string
	Event       string
	HeadSHA     string    // Commit SHA for this run
	CreatedAt   time.Time
	UpdatedAt   time.Time
	WorkflowPath string
	RunNumber   int
	Attempt     int
}

// ListWorkflowRuns retrieves recent workflow runs
func (c *Client) ListWorkflowRuns(limit int) ([]*WorkflowRun, error) {
	if limit <= 0 {
		limit = 10
	}

	opts := &github.ListWorkflowRunsOptions{
		ListOptions: github.ListOptions{PerPage: limit},
	}

	runs, _, err := c.client.Actions.ListRepositoryWorkflowRuns(
		c.ctx,
		c.repo.Owner,
		c.repo.Name,
		opts,
	)
	if err != nil {
		return nil, errors.GitHubAPIError("list_workflow_runs", err)
	}

	var result []*WorkflowRun
	for _, run := range runs.WorkflowRuns {
		// Construct workflow path from workflow name
		// The API doesn't provide the exact path, so we construct it
		workflowName := run.GetName()
		if workflowName == "" {
			workflowName = "unknown"
		}
		workflowPath := ".github/workflows/" + strings.ToLower(strings.ReplaceAll(workflowName, " ", "-")) + ".yml"
		
		result = append(result, &WorkflowRun{
			ID:          run.GetID(),
			Name:        run.GetName(),
			DisplayTitle: run.GetDisplayTitle(),
			Status:      run.GetStatus(),
			Conclusion:  run.GetConclusion(),
			Event:       run.GetEvent(),
			HeadSHA:     run.GetHeadSHA(),
			CreatedAt:   run.GetCreatedAt().Time,
			UpdatedAt:   run.GetUpdatedAt().Time,
			WorkflowPath: workflowPath,
			RunNumber:   run.GetRunNumber(),
			Attempt:     run.GetRunAttempt(),
		})
	}

	c.logger.Debug("Retrieved %d workflow runs", len(result))
	return result, nil
}

// GetFailedWorkflowRuns retrieves only failed workflow runs from the latest push
func (c *Client) GetFailedWorkflowRuns(limit int) ([]*WorkflowRun, error) {
	runs, err := c.ListWorkflowRuns(limit * 2) // Fetch more to ensure we get latest commit
	if err != nil {
		return nil, err
	}

	if len(runs) == 0 {
		return []*WorkflowRun{}, nil
	}

	// Find the most recent commit SHA (latest push)
	latestCommitSHA := runs[0].HeadSHA
	
	// Only return failed runs from the latest commit
	var failed []*WorkflowRun
	for _, run := range runs {
		// Only consider runs from the latest commit
		if run.HeadSHA != latestCommitSHA {
			continue
		}
		
		// Check if it failed
		if run.Conclusion == "failure" || (run.Status == "completed" && run.Conclusion != "success") {
			failed = append(failed, run)
		}
	}

	c.logger.Info("Found %d failed runs from latest commit (%s)", len(failed), latestCommitSHA[:7])
	return failed, nil
}

// GetWorkflowJobLogs retrieves logs for all failed jobs in a workflow run
func (c *Client) GetWorkflowJobLogs(runID int64) (string, error) {
	jobs, _, err := c.client.Actions.ListWorkflowJobs(
		c.ctx,
		c.repo.Owner,
		c.repo.Name,
		runID,
		nil,
	)
	if err != nil {
		return "", errors.GitHubAPIError("list_workflow_jobs", err)
	}

	var logBuilder strings.Builder
	failedCount := 0
	
	// First pass: try to get logs from explicitly failed jobs
	for _, job := range jobs.Jobs {
		if job.GetConclusion() == "failure" {
			failedCount++
			logBuilder.WriteString(fmt.Sprintf("\n=== Job: %s (ID: %d) ===\n", job.GetName(), job.GetID()))
			
			// Get job logs
			logs, _, err := c.client.Actions.GetWorkflowJobLogs(
				c.ctx,
				c.repo.Owner,
				c.repo.Name,
				job.GetID(),
				2, // Follow redirects
			)
			if err != nil {
				c.logger.Warn("Failed to get logs for job %d: %v", job.GetID(), err)
				continue
			}

			logBuilder.WriteString(logs.String())
			logBuilder.WriteString("\n")
		}
	}

	// If no failed jobs found, try cancelled or incomplete jobs
	if failedCount == 0 {
		c.logger.Debug("No failed jobs found, checking cancelled/skipped jobs")
		for _, job := range jobs.Jobs {
			conclusion := job.GetConclusion()
			status := job.GetStatus()
			
			// Include cancelled, timed_out, or still in_progress jobs
			if conclusion == "cancelled" || conclusion == "timed_out" || 
			   (status == "completed" && conclusion != "success" && conclusion != "skipped") {
				failedCount++
				logBuilder.WriteString(fmt.Sprintf("\n=== Job: %s (Status: %s, Conclusion: %s) ===\n", 
					job.GetName(), status, conclusion))
				
				logs, _, err := c.client.Actions.GetWorkflowJobLogs(
					c.ctx,
					c.repo.Owner,
					c.repo.Name,
					job.GetID(),
					2,
				)
				if err != nil {
					logBuilder.WriteString(fmt.Sprintf("[Could not retrieve logs: %v]\n", err))
					continue
				}

				logBuilder.WriteString(logs.String())
				logBuilder.WriteString("\n")
			}
		}
	}

	// If still no logs, the workflow might have failed at configuration level
	if failedCount == 0 {
		c.logger.Warn("Workflow run marked as failed but contains no failed/cancelled jobs")
		return "", errors.ValidationError("get_workflow_job_logs", 
			"workflow failed but no job logs available (possible configuration error)")
	}

	c.logger.Debug("Retrieved logs from %d jobs", failedCount)
	
	// Truncate if needed
	result := logBuilder.String()
	if len(result) > c.config.MaxLogSize {
		truncated := "... [LOGS TRUNCATED FOR SAFETY] ...\n" + result[len(result)-c.config.MaxLogSize:]
		c.logger.Warn("Logs truncated from %d to %d characters", len(result), len(truncated))
		return truncated, nil
	}

	return result, nil
}

// ListWorkflowFiles retrieves all workflow YAML files from .github/workflows
func (c *Client) ListWorkflowFiles() ([]string, error) {
	_, directoryContent, _, err := c.client.Repositories.GetContents(
		c.ctx,
		c.repo.Owner,
		c.repo.Name,
		".github/workflows",
		nil,
	)
	if err != nil {
		return nil, errors.GitHubAPIError("list_workflow_files", err)
	}

	var files []string
	for _, file := range directoryContent {
		name := file.GetName()
		if name != "" && (strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml")) {
			files = append(files, name)
		}
	}

	c.logger.Debug("Found %d workflow files", len(files))
	return files, nil
}

// GetWorkflowFileContent retrieves the content of a workflow file
func (c *Client) GetWorkflowFileContent(path string) (string, error) {
	// Ensure path starts with .github/workflows
	if !strings.HasPrefix(path, ".github/workflows/") {
		path = ".github/workflows/" + strings.TrimPrefix(path, "/")
	}

	fileContent, _, _, err := c.client.Repositories.GetContents(
		c.ctx,
		c.repo.Owner,
		c.repo.Name,
		path,
		nil,
	)
	if err != nil {
		return "", errors.GitHubAPIError("get_workflow_file_content", err).WithPath(path)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return "", errors.ValidationError("get_workflow_file_content", "failed to decode file content").WithPath(path)
	}

	return content, nil
}
