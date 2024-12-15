# Git Multi-Push

A Go app for pushing Git repositories to multiple remote sources (GitHub and/or GitLab).

## Prerequisites

- Working knowledge of basic git commands
- Git installed and configured on your system
- SSH keys set up for your GitHub/GitLab accounts

## What This Tool Does

✅ This tool DOES:
- Push to multiple remotes simultaneously
- Configure remote repositories
- Support force pushing when needed
- Work from any directory in your git repository

❌ This tool does NOT:
- Create or manage branches
- Handle merges
- Resolve conflicts
- Replace standard git commands
- Manage your git workflow


### Scenario 1: Working on a Feature
```bash
# 1. Create feature branch (use git)
git checkout -b feature/awesome

# 2. Make changes and commit (use git)
git add .
git commit -m "Add awesome feature"

# 3. Switch to main branch (use git)
git checkout main

# 4. Merge your changes (use git)
git merge feature/awesome

# 5. Push to all remotes (use git-multi-push)
./git-multi-push
```

### Scenario 2: Force Push Needed
```bash
# 1. Handle your branch changes (use git)
git checkout feature/branch
git rebase main

# 2. Force push to all remotes (use git-multi-push)
./git-multi-push --force
```

## Building

1. Clone or download this repository:
```bash
git clone <repository-url>
cd git-multi-push
```

2. Build the application:
```bash
# Make build script executable (Linux/macOS only)
chmod +x build.sh

# Build for your current platform
./build.sh

# Or build for a specific platform
./build.sh windows
./build.sh linux
./build.sh darwin

# Or build for all platforms
./build.sh all

# Clean build artifacts
./build.sh clean
```

3. The compiled binary will be in the `build` directory.

## Setup

1. First, clone your existing repository from either GitHub or GitLab:
```bash
# If starting with GitHub repository
git clone git@github.com:username/repository.git

# Or if starting with GitLab repository
git clone git@gitlab.com:username/repository.git
```

2. Move into your repository directory:
```bash
cd repository
```

3. Copy the git-multi-push tool into your repository directory or add it to your PATH.

4. Run the setup configuration and enter your information:
```bash
$ ./git-multi-push --setup
Enter GitHub information:
(Just the repository name, not the full URL)
GitHub username: TeemoTheYiffer
GitHub repository name (e.g., 'git-multi-push'): git-multi-push

Enter GitLab information (press Enter to skip):
GitLab username: TeemoTheYiffer
GitLab repository name (e.g., 'git-multi-push'): git-multi-push

Configuration to be saved:
GitHub: TeemoTheYiffer/git-multi-push
GitLab: TeemoTheYiffer/git-multi-push

Is this correct? [Y/n]: y
Configuration saved successfully
```

## Usage

### Command Line Options

- `--setup`: Run initial configuration
- `--force`: Force push to remotes
- `--help`: Show help message

### Push and Merge Example
```bash
$ /path/to/git-multi-push
Operating on git repository at: /path/to/repo
Current branch: feature/awesome
Would you like to merge your changes? [y/N]: y

Available branches:
1: main
2: development
3: feature/awesome

Enter the branch name to merge into: main
Enter merge commit message: Added awesome feature
Successfully merged 'feature/awesome' into 'main'
Operations completed successfully
```

## Best Practices

1. **Development Workflow**
   ```bash
   # Create a development branch
   git checkout -b development

   # Make your changes
   git add .
   git commit -m "Your changes"

   # Push to development branch
   ./git-multi-push

   # Create merge request
   ```

2. **Repository Setup**
   - GitHub: Generally more permissive by default
   - GitLab: More restrictive, protects `main` branch
   - Consider standardizing protection rules across both platforms

3. **Recommended Branch Structure**
   ```
   main (protected)
   ├── development (semi-protected)
   └── feature/* (unprotected)
   ```

## Troubleshooting

1. "Not in a git repository"
   - Make sure you're running the command from within a Git repository
   - Run `git status` to verify it's a valid Git repository

2. "Permission denied"
   - Make sure the binary is executable: `chmod +x git-multi-push`

3. "Failed to push to remote"
   - Verify your SSH keys are set up correctly
   - Check your repository permissions
   - Ensure your local repository is up to date

### GitLab Protected Branches

If you see this error:
```
remote: GitLab: You are not allowed to force push code to a protected branch on this project.
```

This occurs because GitLab protects certain branches (like `main` or `master`) by default. You have several options:

#### Option 1: Unprotect the Branch (If you have admin access)
1. Go to your GitLab repository
2. Navigate to Settings → Repository → Protected Branches
3. Find the protected branch (usually `main`)
4. Click "Unprotect" or modify permissions to allow force push

#### Option 2: Use a Different Branch (Recommended)
```bash
# Create and switch to a new branch
git checkout -b development

# Push to the new branch
./git-multi-push

# Then merge through GitLab's web interface
```

#### Option 3: Configure Branch Protection Rules
If you're the repository owner, you can:
1. Keep protection but allow force push for maintainers
2. Customize protection rules per branch
3. Set up different rules for different user roles

### Different Commit Histories

If you see an error like:
```
! [rejected]        main -> main (fetch first)
error: failed to push some refs to 'gitlab.com:user/repo.git'
```

This means one of your remotes has commits that aren't in your local repository. Here's how to resolve it:

**Option 1: Pull and Merge (Recommended for first-time setup)**
   ```bash
   # Add the remote if it doesn't exist
   git remote add gitlab git@gitlab.com:username/repository.git

   # Fetch the remote repository
   git fetch gitlab

   # Merge the remote changes
   git pull gitlab main --allow-unrelated-histories

   # Resolve any conflicts if they occur
   
   # Then try pushing again
   ./git-multi-push
   ```

**Option 2: Force Push (Use with caution)**
   ```bash
   # Only use if you're sure you want to overwrite remote changes
   ./git-multi-push --force
   ```

### Initial Setup Workflow

When setting up a repository for the first time with multiple remotes:

1. **Clone your primary repository**
   ```bash
   git clone git@github.com:username/repository.git
   cd repository
   ```

2. **Add and sync with second remote**
   ```bash
   # Add GitLab remote
   git remote add gitlab git@gitlab.com:username/repository.git

   # Fetch GitLab repository
   git fetch gitlab

   # If GitLab has existing content, merge it
   git pull gitlab main --allow-unrelated-histories

   # Configure git-multi-push
   ./git-multi-push --setup

   # Push to both remotes
   ./git-multi-push
   ```

### Best Practices for Multiple Remotes

1. **Before making changes**
   ```bash
   # Update from both remotes
   git pull github main
   git pull gitlab main
   ```

2. **After making changes**
   ```bash
   # Push to both using git-multi-push
   ./git-multi-push
   ```

3. **If pushes are rejected**
   - Pull from the rejecting remote
   - Resolve any conflicts
   - Try pushing again
   - Use force push only if you're certain about overwriting remote changes

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

