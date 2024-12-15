package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"git-multi-push/pkg/git"
)

func readUserInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func handleCommit(gitOp *git.GitOperation) error {
	hasChanges, err := gitOp.HasUncommittedChanges()
	if err != nil {
		return err
	}

	if !hasChanges {
		fmt.Println("No changes to commit")
		return nil
	}

	fmt.Println("\nCurrent git status:")
	if err := gitOp.ShowStatus(); err != nil {
		return err
	}

	commit := readUserInput("\nWould you like to commit these changes? [y/N]: ")
	if strings.ToLower(commit) != "y" {
		return fmt.Errorf("changes must be committed before pushing. Operation cancelled")
	}

	message := readUserInput("Enter commit message: ")
	if message == "" {
		return fmt.Errorf("commit message cannot be empty")
	}

	if err := gitOp.Commit(message); err != nil {
		return err
	}

	fmt.Println("Changes committed successfully")
	return nil
}

func handleMerge(gitOp *git.GitOperation) error {
	// Get list of branches first
	branches, err := gitOp.ListBranches()
	if err != nil {
		return err
	}

	currentBranch, err := gitOp.GetCurrentBranch()
	if err != nil {
		return err
	}

	// Filter out current branch from available branches
	availableBranches := []string{}
	for _, branch := range branches {
		if branch != currentBranch {
			availableBranches = append(availableBranches, branch)
		}
	}

	// If no other branches available, skip merge prompt
	if len(availableBranches) == 0 {
		fmt.Println("\nNo other branches available for merging.")
		return nil
	}

	// Ask if user wants to merge
	fmt.Printf("\nCurrent branch: %s\n", currentBranch)
	merge := readUserInput("Would you like to merge your changes? [y/N]: ")
	if strings.ToLower(merge) != "y" {
		return nil
	}

	// Show available branches
	fmt.Println("\nAvailable branches:")
	for i, branch := range availableBranches {
		fmt.Printf("%d: %s\n", i+1, branch)
	}

	// Get target branch
	targetBranch := readUserInput("\nEnter the branch name to merge into: ")
	found := false
	for _, branch := range availableBranches {
		if branch == targetBranch {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("branch '%s' not found", targetBranch)
	}

	// Get commit message
	message := readUserInput("Enter merge commit message: ")
	if message == "" {
		message = fmt.Sprintf("Merge branch '%s' into %s", currentBranch, targetBranch)
	}

	// Perform merge
	if err := gitOp.MergeBranch(currentBranch, targetBranch, message); err != nil {
		return err
	}

	fmt.Printf("Successfully merged '%s' into '%s'\n", currentBranch, targetBranch)
	return nil
}

func main() {
	// Parse command line flags
	forcePush := flag.Bool("force", false, "Force push to remotes")
	setupMode := flag.Bool("setup", false, "Run setup configuration")
	flag.Parse()

	// Setup logging
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Initialize git operations
	gitOp := git.NewGitOperation(logger)

	// Check git installation
	if err := gitOp.CheckGitInstalled(); err != nil {
		logger.Fatal(err)
	}

	// Handle setup mode
	if *setupMode {
		// ... setup code remains the same ...
		return
	}

	// Check if we're in a git repository
	isRepo, repoPath := gitOp.IsGitRepo()
	if !isRepo {
		logger.Fatal("Not in a git repository")
	}
	logger.Printf("Operating on git repository at: %s", repoPath)

	// Step 1: Sync with remotes
	fmt.Println("Synchronizing with remotes...")
	if err := gitOp.SyncWithRemotes(); err != nil {
		logger.Printf("Warning: Failed to sync with remotes: %v", err)
		// Continue anyway as this might be first push
	}

	// Step 2: Handle commits if there are changes
	if err := handleCommit(gitOp); err != nil {
		logger.Fatal(err)
	}

	// Step 3: Handle merge if requested
	if err := handleMerge(gitOp); err != nil {
		logger.Fatal(err)
	}

	// Step 4: Push to remotes
	if err := gitOp.Push(*forcePush); err != nil {
		logger.Fatal(err)
	}

	fmt.Println("Operations completed successfully")
}
