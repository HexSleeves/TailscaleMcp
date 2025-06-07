#!/bin/bash

# A comprehensive script to automate versioning, building, and publishing of a Node.js project.

set -e # Exit on any error

# --- Configuration ---
# Your primary git branch
MAIN_BRANCH="main"
DOCKER_HUB_USERNAME="hexsleeves"
DOCKER_HUB_REPO="tailscale-mcp-server"

# --- Colors for output ---
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# --- Helper Functions ---

print_status() {
  echo -e "${BLUE}[INFO]${NC} $1"
}
print_success() {
  echo -e "${GREEN}[SUCCESS]${NC} $1"
}
print_warning() {
  echo -e "${YELLOW}[WARNING]${NC} $1"
}
print_error() {
  echo -e "${RED}[ERROR]${NC} $1"
  exit 1
}

prompt_yes_no() {
  local prompt="$1"
  local response
  while true; do
    read -p "$prompt (y/n): " response
    case $response in
    [Yy]*) return 0 ;;
    [Nn]*) return 1 ;;
    *) echo "Please answer yes (y) or no (n)." ;;
    esac
  done
}

# --- Package & Git Info Functions ---

get_current_version() {
  node -p "require('./package.json').version"
}
get_package_name() {
  node -p "require('./package.json').name"
}
get_github_owner() {
  # Tries to automatically determine the GitHub user/org from the git remote URL
  local remote_url
  remote_url=$(git remote get-url origin 2>/dev/null) || return
  # Extracts 'user' from 'git@github.com:user/repo.git' or 'https://github.com/user/repo.git'
  echo "$remote_url" | sed -n -e 's#.*github.com[:/]\([^/]*\)/.*#\1#p'
}

# --- Core Workflow Functions ---

bump_version() {
  local bump_type="$1"
  local new_version

  if [ "$bump_type" == "custom" ]; then
    read -p "Enter the custom version number: " new_version
    npm version "$new_version" --no-git-tag-version
  else
    # Use npm version to bump and capture the new version
    new_version=$(npm version "$bump_type" --no-git-tag-version | sed 's/v//')
  fi

  print_success "Version bumped to $new_version"

  # Commit the version change
  print_status "Committing version bump..."
  git add package.json package-lock.json
  git commit -m "chore: bump version to $new_version"

  # Create git tag
  print_status "Creating git tag v$new_version..."
  git tag "v$new_version"

  echo "$new_version"
}

publish_npm() {
  local version="$1"
  print_status "Preparing to publish to npm..."

  if prompt_yes_no "Run 'npm run build' first?"; then
    npm run build
  fi

  if prompt_yes_no "Run tests before publishing?"; then
    npm test
  fi

  if prompt_yes_no "Ready to publish version $version to npm?"; then
    npm publish --access public
    print_success "Published to npm successfully!"
  else
    print_warning "NPM publish aborted by user."
  fi
}

build_docker() {
  local version="$1"
  local docker_name="$DOCKER_HUB_USERNAME/$DOCKER_HUB_REPO"
  print_status "Building Docker image: $docker_name:$version"
  docker build -t "$docker_name:$version" -t "$docker_name:latest" .
  print_success "Built Docker image: $docker_name:$version"
}

publish_docker() {
  local version="$1"
  local package_name="$2"
  local docker_name

  docker_name="$DOCKER_HUB_USERNAME/$DOCKER_HUB_REPO"
  build_docker "$version"

  if prompt_yes_no "Push image '$docker_name' to Docker Hub?"; then
    print_status "Pushing to Docker Hub: $docker_name:$version..."
    docker push "$docker_name:$version"

    print_status "Pushing to Docker Hub: $docker_name:latest..."
    docker push "$docker_name:latest"
    print_success "Published to Docker Hub successfully!"
  else
    print_warning "Docker Hub push aborted by user."
  fi
}

publish_ghcr() {
  local version="$1"
  local package_name="$2"
  local docker_name
  local github_owner
  local ghcr_image

  docker_name="$DOCKER_HUB_USERNAME/$DOCKER_HUB_REPO"
  github_owner=$(get_github_owner)

  if [ -z "$github_owner" ]; then
    print_warning "Could not determine GitHub owner from git remote."
    read -p "Enter GitHub username/organization: " github_owner
    if [ -z "$github_owner" ]; then
      print_error "GitHub owner is required for GHCR publish."
    fi
  else
    print_status "Detected GitHub owner: $github_owner"
  fi

  lowercase_github_owner=$(echo "$github_owner" | tr '[:upper:]' '[:lower:]')
  ghcr_image="ghcr.io/$lowercase_github_owner/$DOCKER_HUB_REPO"

  print_status "Tagging image for GHCR: $ghcr_image:$version"
  docker tag "$docker_name:latest" "$ghcr_image:$version"
  docker tag "$docker_name:latest" "$ghcr_image:latest"

  if prompt_yes_no "Push image to GitHub Container Registry?"; then
    print_status "Pushing to GHCR..."
    docker push "$ghcr_image:$version"
    docker push "$ghcr_image:latest"
    print_success "Published to GitHub Container Registry successfully!"
  else
    print_warning "GHCR push aborted by user."
  fi
}

# --- Main Script Logic ---
main() {
  print_status "🚀 Starting publish workflow..."

  # --- Pre-flight Checks ---
  if ! git rev-parse --git-dir >/dev/null 2>&1; then
    print_error "Not a git repository. Aborting."
  fi
  if ! git diff-index --quiet HEAD --; then
    print_warning "You have uncommitted changes."
    if ! prompt_yes_no "Continue anyway?"; then
      print_error "Aborted by user."
    fi
  fi

  local current_version
  local package_name
  current_version=$(get_current_version)
  package_name=$(get_package_name)

  print_status "Package: $package_name"
  print_status "Current version: $current_version"

  # --- Version Bumping ---
  local new_version="$current_version"
  local bump_type=""

  echo ""
  print_status "Select a version bump type:"

  # Pre-calculate version bumps to avoid repeated npx semver calls
  local patch_ver minor_ver major_ver
  patch_ver=$(npx semver "$current_version" -i patch)
  minor_ver=$(npx semver "$current_version" -i minor)
  major_ver=$(npx semver "$current_version" -i major)

  local options=(
    "patch ($patch_ver)"
    "minor ($minor_ver)"
    "major ($major_ver)"
    "prerelease"
    "custom"
    "skip"
  )
  PS3="Your choice: " # Prompt for the select menu
  select opt in "${options[@]}"; do
    case $opt in
    "patch"*)
      bump_type="patch"
      break
      ;;
    "minor"*)
      bump_type="minor"
      break
      ;;
    "major"*)
      bump_type="major"
      break
      ;;
    "prerelease")
      bump_type="prerelease"
      break
      ;;
    "custom")
      bump_type="custom"
      break
      ;;
    "skip")
      bump_type="skip"
      break
      ;;
    *)
      print_warning "Invalid option. Please try again."
      ;;
    esac
  done

  if [[ -n "$bump_type" && "$bump_type" != "skip" ]]; then
    new_version=$(bump_version "$bump_type")
  else
    print_status "Skipping version bump. Using current version: $current_version"
  fi

  echo ""
  print_status "Working with version: $new_version"

  # --- Publishing Options ---
  local did_publish_docker=false
  echo ""
  if prompt_yes_no "Publish to npm?"; then
    publish_npm "$new_version"
  fi

  echo ""
  if prompt_yes_no "Build and publish Docker image?"; then
    # We only need to build the Docker image once
    publish_docker "$new_version" "$package_name"
    did_publish_docker=true
  fi

  echo ""
  if prompt_yes_no "Publish to GitHub Container Registry?"; then
    if ! $did_publish_docker; then
      print_warning "Docker image was not built in the previous step."
      if prompt_yes_no "Build it now?"; then
        build_docker "$new_version"
      else
        print_error "Cannot publish to GHCR without a built image."
      fi
    fi
    publish_ghcr "$new_version" "$package_name"
  fi

  # --- Final Git Push ---
  if [ "$new_version" != "$current_version" ]; then
    echo ""
    if prompt_yes_no "Push git commits and tags to origin?"; then
      git push origin "$MAIN_BRANCH"
      git push origin "v$new_version"
      print_success "Git changes and tags pushed!"
    fi
  fi

  echo ""
  print_success "🎉 Publish workflow completed!"
  print_status "Final version: $new_version"
}

# Run the main function with all passed arguments
main "$@"
