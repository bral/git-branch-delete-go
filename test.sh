#!/bin/bash

# Ensure we're inside a Git repository.
if ! git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
  echo "Error: This is not a Git repository."
  exit 1
fi

# Set the locale to 'C' to avoid illegal byte sequence errors.
export LC_ALL=C

# Default to 5 branches if no count is provided
count=${1:-5}

echo "Creating $count test branches..."

# Store the current branch
current_branch=$(git rev-parse --abbrev-ref HEAD)

# Loop to create test branches
for i in $(seq 1 $count); do
  # Generate a random branch name prefixed with "test_"
  branch_name="test_$(tr -dc 'a-f0-9' </dev/urandom | head -c 8)"

  # Ensure the branch name is unique
  while git rev-parse --verify "$branch_name" >/dev/null 2>&1; do
    branch_name="test_$(tr -dc 'a-f0-9' </dev/urandom | head -c 8)"
  done

  # Create branch and switch to it
  git checkout -b "$branch_name"

  # Create an empty commit
  git commit --allow-empty -m "Test commit for $branch_name"

  # Push to remote
  if git push -u origin "$branch_name"; then
    echo "Created and pushed branch: $branch_name"
  else
    echo "Warning: Failed to push branch $branch_name"
  fi
done

# Return to original branch
git checkout "$current_branch"

echo -e "\nCreated $count test branches successfully! 🎉"
echo "Run 'git-branch-delete interactive --all' to clean them up"
