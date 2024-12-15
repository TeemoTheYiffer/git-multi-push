package git

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Config struct {
	GithubUsername string `json:"github_username"`
	GithubRepo     string `json:"github_repo"`
	GitlabUsername string `json:"gitlab_username"`
	GitlabRepo     string `json:"gitlab_repo"`
}

type GitOperation struct {
	logger *log.Logger
	config *Config
}

// Helper function to extract repo name from URL or plain name
func ExtractRepoName(input string) string {
	// Handle full URLs
	if strings.Contains(input, "://") {
		// Remove .git suffix if present
		input = strings.TrimSuffix(input, ".git")
		// Get the last part of the path
		parts := strings.Split(input, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	// Return as-is if it's just a name
	return strings.TrimSuffix(input, ".git")
}

func NewGitOperation(logger *log.Logger) *GitOperation {
	return &GitOperation{
		logger: logger,
	}
}

func (g *GitOperation) GetConfigDir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "git-multi-push")
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "git-multi-push")
}

func (g *GitOperation) LoadConfig() error {
	configPath := filepath.Join(g.GetConfigDir(), "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("config not found, run setup first: %v", err)
	}

	g.config = &Config{}
	if err := json.Unmarshal(data, g.config); err != nil {
		return fmt.Errorf("invalid config format: %v", err)
	}
	return nil
}

func (g *GitOperation) SaveConfig(config *Config) error {
	configDir := g.GetConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}

	g.config = config
	return nil
}

func (g *GitOperation) CheckGitInstalled() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git is not installed: %v", err)
	}
	return nil
}

func (g *GitOperation) IsGitRepo() (bool, string) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return false, ""
	}
	return true, strings.TrimSpace(string(output))
}

func (g *GitOperation) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func (g *GitOperation) ListBranches() ([]string, error) {
	cmd := exec.Command("git", "branch")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %v", err)
	}

	branches := []string{}
	for _, branch := range strings.Split(string(output), "\n") {
		// Remove the '* ' from current branch and any whitespace
		branch = strings.TrimSpace(strings.TrimPrefix(branch, "*"))
		if branch != "" {
			branches = append(branches, branch)
		}
	}
	return branches, nil
}

func (g *GitOperation) FetchAllRemotes() error {
	cmd := exec.Command("git", "fetch", "--all")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to fetch remotes: %s", string(output))
	}
	return nil
}

func (g *GitOperation) ListRemoteBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "-r")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list remote branches: %s", string(output))
	}

	branches := []string{}
	for _, branch := range strings.Split(string(output), "\n") {
		branch = strings.TrimSpace(branch)
		if branch != "" && !strings.Contains(branch, "->") {
			// Remove 'origin/' prefix
			branch = strings.TrimPrefix(branch, "origin/")
			branches = append(branches, branch)
		}
	}
	return branches, nil
}

func (g *GitOperation) SyncWithRemotes() error {
	// Fetch from all remotes
	if err := g.FetchAllRemotes(); err != nil {
		return err
	}

	currentBranch, err := g.GetCurrentBranch()
	if err != nil {
		return err
	}

	// Try to pull from each remote
	remotes := []string{"github", "gitlab"}
	for _, remote := range remotes {
		pullCmd := exec.Command("git", "pull", remote, currentBranch, "--allow-unrelated-histories")
		output, err := pullCmd.CombinedOutput()
		g.logger.Printf("Syncing with %s: %s", remote, string(output))
		if err != nil {
			g.logger.Printf("Warning: Could not pull from %s: %v", remote, err)
			// Continue with other remotes even if one fails
		}
	}

	return nil
}

func (g *GitOperation) ValidateMerge(fromBranch, toBranch string) error {
	if fromBranch == toBranch {
		return fmt.Errorf("cannot merge a branch into itself")
	}
	return nil
}

// Update the existing MergeBranch method
func (g *GitOperation) MergeBranch(fromBranch, toBranch, message string) error {
	// Validate the merge
	if err := g.ValidateMerge(fromBranch, toBranch); err != nil {
		return err
	}

	// First checkout the target branch
	checkoutCmd := exec.Command("git", "checkout", toBranch)
	if output, err := checkoutCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to checkout %s: %s", toBranch, string(output))
	}

	// Then merge with the specified message
	mergeArgs := []string{"merge", fromBranch}
	if message != "" {
		mergeArgs = append(mergeArgs, "-m", message)
	}

	mergeCmd := exec.Command("git", mergeArgs...)
	if output, err := mergeCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to merge %s into %s: %s", fromBranch, toBranch, string(output))
	}

	return nil
}

func (g *GitOperation) Push(forcePush bool) error {
	// First get the root directory of the git repo
	isRepo, rootDir := g.IsGitRepo()
	if !isRepo {
		return fmt.Errorf("not in a git repository")
	}

	// Log the repository location for clarity
	g.logger.Printf("Operating on git repository at: %s", rootDir)

	if err := g.LoadConfig(); err != nil {
		return err
	}

	remotes := map[string]string{
		"github": fmt.Sprintf("git@github.com:%s/%s.git", g.config.GithubUsername, g.config.GithubRepo),
		"gitlab": fmt.Sprintf("git@gitlab.com:%s/%s.git", g.config.GitlabUsername, g.config.GitlabRepo),
	}

	for name, url := range remotes {
		if err := g.addRemote(name, url); err != nil {
			return err
		}
		if err := g.pushToRemote(name, forcePush); err != nil {
			return err
		}
	}

	return nil
}

func (g *GitOperation) addRemote(name, url string) error {
	checkCmd := exec.Command("git", "remote", "get-url", name)
	if checkCmd.Run() == nil {
		cmd := exec.Command("git", "remote", "set-url", name, url)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update remote %s: %v", name, err)
		}
	} else {
		cmd := exec.Command("git", "remote", "add", name, url)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to add remote %s: %v", name, err)
		}
	}
	return nil
}

func (g *GitOperation) pushToRemote(remote string, forcePush bool) error {
	args := []string{"push", remote}
	if forcePush {
		args = append(args, "--force")
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		// Check for protected branch error
		if strings.Contains(outputStr, "protected branch") {
			return fmt.Errorf(`failed to push to %s: %s

GitLab protected branch detected. You have several options:

1. Use a development branch instead:
   git checkout -b development
   ./git-multi-push

2. Unprotect the branch in GitLab:
   - Go to GitLab repository → Settings → Repository → Protected Branches
   - Unprotect or modify permissions for the branch

3. Use GitLab's web interface to merge changes

See README for more detailed instructions on working with protected branches.`, remote, outputStr)
		}

		// Check for fetch first error
		if strings.Contains(outputStr, "fetch first") {
			return fmt.Errorf(`failed to push to %s: %s

To resolve this, you can either:
1. Pull and merge changes (recommended):
   git pull %s main --allow-unrelated-histories

2. Force push (use with caution):
   ./git-multi-push --force

See README for more detailed instructions.`, remote, outputStr, remote)
		}

		return fmt.Errorf("failed to push to %s: %s", remote, outputStr)
	}

	g.logger.Printf("Successfully pushed to %s", remote)
	return nil
}
