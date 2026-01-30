package context

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"gh-sentinel/internal/errors"
)

// RepoContext holds information about the current repository
type RepoContext struct {
	Owner         string
	Name          string
	FullName      string
	DefaultBranch string
	IsPrivate     bool
}

// ghRepoResponse matches the structure returned by gh repo view --json
type ghRepoResponse struct {
	Owner struct {
		Login string `json:"login"`
	} `json:"owner"`
	Name             string `json:"name"`
	NameWithOwner    string `json:"nameWithOwner"`
	DefaultBranchRef struct {
		Name string `json:"name"`
	} `json:"defaultBranchRef"`
	IsPrivate bool `json:"isPrivate"`
}

// DetectRepository uses gh CLI to detect current repository context
func DetectRepository() (*RepoContext, error) {
	// Check if gh CLI is available
	if _, err := exec.LookPath("gh"); err != nil {
		return nil, errors.AuthError("detect_repository", fmt.Errorf("gh CLI not found in PATH"))
	}

	// Get repository information
	cmd := exec.Command("gh", "repo", "view", "--json", "owner,name,nameWithOwner,defaultBranchRef,isPrivate")
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.GitHubAPIError("detect_repository", fmt.Errorf("not in a git repository or gh not authenticated"))
	}

	// Parse JSON response
	var response ghRepoResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, errors.ValidationError("detect_repository", fmt.Sprintf("failed to parse repository information: %v", err))
	}

	// Validate required fields
	if response.Owner.Login == "" || response.Name == "" {
		return nil, errors.ValidationError("detect_repository", "missing required repository information")
	}

	ctx := &RepoContext{
		Owner:         response.Owner.Login,
		Name:          response.Name,
		FullName:      response.NameWithOwner,
		DefaultBranch: response.DefaultBranchRef.Name,
		IsPrivate:     response.IsPrivate,
	}

	// Fallback for FullName if not provided
	if ctx.FullName == "" {
		ctx.FullName = fmt.Sprintf("%s/%s", ctx.Owner, ctx.Name)
	}

	return ctx, nil
}

// GetAuthToken retrieves the GitHub authentication token from gh CLI
func GetAuthToken() (string, error) {
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return "", errors.AuthError("get_auth_token", err)
	}
	
	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", errors.AuthError("get_auth_token", fmt.Errorf("empty token received"))
	}
	
	return token, nil
}

// CheckAuthentication verifies that gh CLI is authenticated
func CheckAuthentication() error {
	cmd := exec.Command("gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		return errors.AuthError("check_authentication", fmt.Errorf("not authenticated with GitHub - run 'gh auth login'"))
	}
	return nil
}
