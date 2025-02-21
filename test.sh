#!/bin/bash

# Ensure we're inside a Git repository.
if ! git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
  echo "Error: This is not a Git repository."
  exit 1
fi

# Set the locale to 'C' to avoid illegal byte sequence errors.
export LC_ALL=C

# Loop to create 1,000 random branches.
for i in $(seq 1 100); do
  # Generate a random branch name prefixed with "rand_".
  branch_name="rand_$(tr -dc 'a-z0-9' </dev/urandom | head -c 8)"
  
  # Ensure the branch name is unique.
  while git rev-parse --verify "$branch_name" >/dev/null 2>&1; do
    branch_name="rand_$(tr -dc 'a-z0-9' </dev/urandom | head -c 8)"
  done
  
  git branch "$branch_name"
done

echo "Created 100 random Git branches."
