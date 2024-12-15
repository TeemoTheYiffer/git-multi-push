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

func handleMerge(gitOp *git.GitOperation) error {
	// Sync with remotes first
	fmt.Println("Synchronizing with remotes...")
	if err := gitOp.SyncWithRemotes(); err != nil {
		return fmt.Errorf("failed to sync with remotes: %v", err)
	}

	currentBranch, err := gitOp.GetCurrentBranch()
	if err != nil {
		return err
	}

	// Ask if user wants to merge
	fmt.Printf("\nCurrent branch: %s\n", currentBranch)
	merge := readUserInput("Would you like to merge your changes? [y/N]: ")
	if strings.ToLower(merge) != "y" {
		return nil
	}

	// Show available branches
	branches, err := gitOp.ListBranches()
	if err != nil {
		return err
	}

	fmt.Println("\nAvailable branches:")
	for i, branch := range branches {
		if branch != currentBranch { // Only show branches that aren't current
			fmt.Printf("%d: %s\n", i+1, branch)
		}
	}

	if len(branches) <= 1 {
		return fmt.Errorf("no other branches available to merge into")
	}

	// Get target branch
	targetBranch := readUserInput("\nEnter the branch name to merge into: ")
	if targetBranch == currentBranch {
		return fmt.Errorf("cannot merge a branch into itself")
	}

	found := false
	for _, branch := range branches {
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
		// ... [setup code remains the same]
	}

	// Check if we're in a git repository
	isRepo, repoPath := gitOp.IsGitRepo()
	if !isRepo {
		logger.Fatal("Not in a git repository")
	}
	logger.Printf("Operating on git repository at: %s", repoPath)

	// Handle merge if requested
	if err := handleMerge(gitOp); err != nil {
		logger.Fatal(err)
	}

	// Push to remotes
	if err := gitOp.Push(*forcePush); err != nil {
		logger.Fatal(err)
	}

	fmt.Println("Operations completed successfully")
}
