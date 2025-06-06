# GitHub Actions CI/CD Pipeline

This repository includes a comprehensive GitHub Actions setup for continuous integration, deployment, and maintenance of the Tailscale MCP Server.

## 🚀 Workflows Overview

### 1. **CI Workflow** (`.github/workflows/ci.yml`)

**Triggers:** Push to `main`/`develop`, Pull Requests
**Purpose:** Continuous Integration and Quality Assurance

**Jobs:**

- **Lint & Type Check**: TypeScript type checking
- **Test**: Multi-version Node.js testing (18, 20, 22) with coverage
- **Build**: Project compilation and artifact verification
- **Security Audit**: Dependency vulnerability scanning

### 2. **Release Workflow** (`.github/workflows/release.yml`)

**Triggers:** Push to `main`, Manual dispatch
**Purpose:** Automated versioning, releases, and NPM publishing

**Features:**

- Semantic versioning based on commit messages
- Automatic changelog generation
- GitHub release creation with assets
- NPM publishing (stable/beta tags)
- Manual release type selection

### 3. **CodeQL Workflow** (`.github/workflows/codeql.yml`)

**Triggers:** Push, Pull Requests, Weekly schedule
**Purpose:** Security analysis and vulnerability detection

**Features:**

- JavaScript/TypeScript security scanning
- SARIF report generation
- GitHub Security tab integration

### 4. **Docker Workflow** (`.github/workflows/docker.yml`)

**Triggers:** Push to `main`, Tags, Pull Requests
**Purpose:** Container image building and publishing

**Features:**

- Multi-architecture builds (AMD64, ARM64)
- GitHub Container Registry publishing
- Trivy security scanning
- Optimized caching

### 5. **Dependency Update Workflow** (`.github/workflows/dependency-update.yml`)

**Triggers:** Weekly schedule, Manual dispatch
**Purpose:** Automated dependency maintenance

**Features:**

- Minor/patch version updates
- Automated testing
- Pull request creation
- Safety checks

## 🔧 Setup Instructions

### Required Secrets

Add these secrets to your GitHub repository settings:

```bash
# NPM Publishing
NPM_TOKEN=your_npm_token_here

# Optional: Enhanced GitHub token for advanced features
GITHUB_TOKEN=your_github_token_here  # Usually auto-provided
```

### NPM Token Setup

1. Go to [npmjs.com](https://www.npmjs.com) and log in
2. Navigate to Access Tokens in your account settings
3. Create a new token with "Automation" type
4. Add it as `NPM_TOKEN` secret in GitHub repository settings

### Repository Settings

1. **Enable GitHub Actions**: Go to Settings → Actions → General
2. **Workflow Permissions**: Set to "Read and write permissions"
3. **Branch Protection**: Configure rules for `main` branch:
   - Require status checks to pass
   - Require branches to be up to date
   - Include administrators

## 📋 Commit Message Conventions

The release workflow uses commit messages to determine version bumps:

```bash
# Patch version (0.1.0 → 0.1.1)
fix: resolve login issue
docs: update README

# Minor version (0.1.0 → 0.2.0)
feat: add new authentication method

# Major version (0.1.0 → 1.0.0)
feat!: redesign API interface
feat: new feature

BREAKING CHANGE: API interface changed
```

## 🎯 Manual Release Process

### Using GitHub UI

1. Go to Actions → Release workflow
2. Click "Run workflow"
3. Select release type (patch/minor/major/prerelease)
4. Click "Run workflow"

### Using Git Tags

```bash
# Create and push a tag
git tag v1.0.0
git push origin v1.0.0
```

## 🐳 Docker Usage

### Pull from GitHub Container Registry

```bash
# Latest version

# Pull from Docker Hub
docker pull hexsleeves/tailscale-mcp-server
docker pull ghcr.io/hexsleeves/tailscale-mcp-server:latest


# Specific version

# Pull from Docker Hub
docker pull hexsleeves/tailscale-mcp-server:v1.0.0

# Pull from GitHub Container Registry
docker pull ghcr.io/hexsleeves/tailscale-mcp-server:v1.0.0
```

### Run Container

```bash
docker run -d \
  --name tailscale-mcp \
  -e TAILSCALE_API_KEY=your_api_key \
  -e TAILSCALE_TAILNET=your_tailnet \
  -e LOG_LEVEL=1 \
  ghcr.io/hexsleeves/tailscale-mcp-server:latest
```

## 🔍 Monitoring and Maintenance

### Workflow Status

- Check the Actions tab for workflow runs
- Review failed builds and address issues
- Monitor security alerts in the Security tab

### Dependency Updates

- Review automated dependency update PRs
- Test thoroughly before merging
- Check for breaking changes in updated packages

### Security Scanning

- Review CodeQL alerts in Security tab
- Address Trivy container scan findings
- Keep dependencies updated

## 🛠️ Troubleshooting

### Common Issues

**Build Failures:**

- Check Node.js version compatibility
- Verify all dependencies are properly installed
- Review TypeScript compilation errors

**Test Failures:**

- Ensure all tests pass locally first
- Check for environment-specific issues
- Review test coverage requirements

**Release Issues:**

- Verify NPM_TOKEN is valid and has publish permissions
- Check package.json version format
- Ensure all required files are included in build

**Docker Build Issues:**

- Review Dockerfile syntax
- Check .dockerignore patterns
- Verify multi-stage build dependencies

### Getting Help

1. Check workflow logs in the Actions tab
2. Review this documentation
3. Check GitHub Actions documentation
4. Open an issue with detailed error information

## 📈 Metrics and Analytics

The workflows provide several metrics:

- Build success/failure rates
- Test coverage percentages
- Security vulnerability counts
- Dependency update frequency
- Release cadence

Monitor these through:

- GitHub Actions dashboard
- Repository Insights
- Security tab alerts
- NPM package statistics
