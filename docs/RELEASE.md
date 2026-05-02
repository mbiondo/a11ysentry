# Release & Deployment Guide

## Overview

This document covers the release process, deployment strategies, and maintenance procedures for A11ySentry.

---

## Cambios de release (actualización de documentación)

- Eliminado examples/web. Ya no forma parte de la distribución ni del flujo de desarrollo.
- Añadidos cuatro ejemplos por framework: Astro, Next.js, Nuxt y SvelteKit.
- Actualizadas las rutas y comandos de compilación para el desarrollo: binario en cmd/a11ysentry, uso de go.work sin cli.
- Nuxt y SvelteKit quedan documentados como scanners integrados en el sistema de detección y se explican las rutas de descubrimiento en monorepos.

## Versioning

A11ySentry follows [Semantic Versioning](https://semver.org/):

- **MAJOR.MINOR.PATCH** (e.g., 1.2.3)
- **MAJOR:** Breaking changes to API or USN schema
- **MINOR:** New features, adapters, or WCAG rules (backward compatible)
- **PATCH:** Bug fixes, performance improvements (backward compatible)

### Version Tags

```bash
# Tag format
v1.2.3

# Pre-release format
v1.2.3-beta.1
v1.2.3-rc.2
```

---

## Release Process

### Prerequisites

- Go 1.24+
- GitHub account with repo access
- GoReleaser installed (`go install github.com/goreleaser/goreleaser@latest`)

### Step 1: Update Changelog

Edit `CHANGELOG.md`:

```markdown
## [0.0.2] - 2026-05-02

### Added
- Advanced WCAG rules for landmarks, modals, and fieldsets.
- WCAG 2.1.1 keyboard navigation rules.
- Responsive TUI refactor and auto-resolving MCP.
- Global install scripts and simplified CLI usage.

## [0.0.1] - 2026-04-30
```

### Step 2: Run Tests

```bash
# All tests
go test ./...
```

### Step 3: Update Version

Update version in `cmd/a11ysentry/main.go`:

```go
var (
    Version = "0.0.2"
)
```

### Step 4: Commit Changes

```bash
git add CHANGELOG.md cmd/a11ysentry/main.go
git commit -m "chore: bump version to 0.0.2"
git push origin main
```

### Step 5: Create Git Tag

```bash
# Create annotated tag
git tag -a v0.0.2 -m "Release v0.0.2"

# Push tag to GitHub
git push origin v0.0.2
```

### Step 6: Build with GoReleaser

```bash
# Test build (no publish)
goreleaser release --snapshot --clean

# Production build
goreleaser release --clean
```

**GoReleaser Configuration:**

`.goreleaser.yaml` handles:
- Multi-platform builds (Windows, macOS, Linux)
- Binary compression (ZIP, TAR.GZ)
- Checksum generation
- GitHub release creation

### Step 7: Verify Release

1. **Check GitHub Releases:** https://github.com/mbiondo/a11ysentry/releases
2. **Verify Binaries:** Download and test on each platform
3. **Test Installers:** Run `install.ps1` and `install.sh`

---

## Build Configuration

### GoReleaser Setup

```yaml
# .goreleaser.yaml
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - id: a11ysentry
    main: ./cmd/a11ysentry
    binary: a11ysentry
    env:
      - CGO_ENABLED=0
    goos:
      - windows
      - darwin
      - linux
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_
      {{- title .Os }}_
      {{- if eq .Arch "amd64" }}x86_64
      {{- else }}{{ .Arch }}{{ end }}
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'
```

### Platform-Specific Builds

| OS | Architecture | Binary Name | Archive |
|----|--------------|-------------|---------|
| Windows | x64 | `a11ysentry.exe` | `a11ysentry_Windows_x86_64.zip` |
| macOS | x64/arm64 | `a11ysentry` | `a11ysentry_Darwin_x86_64.tar.gz` |
| Linux | x64/arm64 | `a11ysentry` | `a11ysentry_Linux_x86_64.tar.gz` |

---

## Installation Scripts

### Windows (PowerShell)

**File:** `install.ps1`

```powershell
# Download latest release
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/mbiondo/a11ysentry/releases/latest"
$asset = $release.assets | Where-Object { $_.name -like "*Windows*x86_64.zip" }

# Download and extract
Invoke-WebRequest -Uri $asset.browser_download_url -OutFile "$env:TEMP\a11ysentry.zip"
Expand-Archive -Path "$env:TEMP\a11ysentry.zip" -DestinationPath "$env:TEMP\a11ysentry"

# Install binary
Move-Item -Path "$env:TEMP\a11ysentry\a11ysentry.exe" -Destination "C:\Program Files\a11ysentry\" -Force

# Add to PATH
$currentUserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentUserPath -notlike "*a11ysentry*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentUserPath;C:\Program Files\a11ysentry\", "User")
}

# Register MCP
& "C:\Program Files\a11ysentry\a11ysentry.exe" mcp --register
```

### Unix (Bash)

**File:** `install.sh`

```bash
#!/bin/bash

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Download latest release
RELEASE_URL=$(curl -s https://api.github.com/repos/mbiondo/a11ysentry/releases/latest | \
  grep "browser_download_url.*${OS}.*${ARCH}" | \
  cut -d '"' -f 4)

# Download and extract
curl -L -o /tmp/a11ysentry.tar.gz "$RELEASE_URL"
tar -xzf /tmp/a11ysentry.tar.gz -C /tmp

# Install binary
sudo mv /tmp/a11ysentry /usr/local/bin/
sudo chmod +x /usr/local/bin/a11ysentry

# Register MCP
a11ysentry mcp --register
```

---

## CI/CD Pipeline

### GitHub Actions Workflow

**.github/workflows/release.yml:**

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Pre-Release Checks

**.github/workflows/ci.yml:**

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      
      - name: Run tests
        run: go test ./...
      
      - name: Run linters
        run: |
          go fmt ./...
          go vet ./...
      
      - name: Test examples
        run: |
          go build -o a11ysentry ./cli
          ./a11ysentry examples/example-astro/src/pages/index.astro
```

---

## Distribution Channels

### 1. GitHub Releases

Primary distribution method. Binaries available at:
https://github.com/mbiondo/a11ysentry/releases

### 2. Package Managers (Future)

#### Homebrew (macOS)

```ruby
# Formula: a11ysentry.rb
class A11ysentry < Formula
  desc "Universal Accessibility Engine for Multi-Platform UI"
  homepage "https://github.com/mbiondo/a11ysentry"
  url "https://github.com/mbiondo/a11ysentry/archive/v0.0.2.tar.gz"
  sha256 "..."
  
  def install
    system "go", "build", *std_go_args, "./cli"
  end
  
  test do
    system "#{bin}/a11ysentry", "--version"
  end
end
```

#### Scoop (Windows)

```json
{
  "version": "0.0.2",
  "description": "Universal Accessibility Engine",
  "homepage": "https://github.com/mbiondo/a11ysentry",
  "url": "https://github.com/mbiondo/a11ysentry/releases/download/v0.0.2/a11ysentry_Windows_x86_64.zip",
  "bin": "a11ysentry.exe"
}
```

### 3. Go Install

```bash
go install github.com/mbiondo/a11ysentry/cli@latest
```

---

## Deployment Verification

### Post-Installation Tests

```bash
# Verify binary is accessible
a11ysentry --version

# Test basic analysis
           a11ysentry examples/example-astro/src/pages/index.astro

# Verify MCP registration
a11ysentry mcp --check-mcp

# Test TUI
a11ysentry --tui
```

### Smoke Test Script

```bash
#!/bin/bash
# tests/smoke_test.sh

echo "Running smoke tests..."

# Test 1: Version
if ! a11ysentry --version > /dev/null 2>&1; then
    echo "❌ Version check failed"
    exit 1
fi
echo "✅ Version check passed"

# Test 2: File analysis
 if ! a11ysentry examples/example-astro/src/pages/index.astro > /dev/null 2>&1; then
    echo "❌ File analysis failed"
    exit 1
fi
echo "✅ File analysis passed"

# Test 3: JSON output
 if ! a11ysentry --format json examples/example-astro/src/pages/index.astro > /dev/null 2>&1; then
    echo "❌ JSON output failed"
    exit 1
fi
echo "✅ JSON output passed"

# Test 4: MCP check
if ! a11ysentry mcp --check-mcp > /dev/null 2>&1; then
    echo "❌ MCP check failed"
    exit 1
fi
echo "✅ MCP check passed"

echo "✅ All smoke tests passed!"
```

---

## Maintenance

### Bug Fix Release

1. **Create Hotfix Branch:**
   ```bash
   git checkout -b hotfix/issue-description main
   ```

2. **Fix Issue:** Implement and test fix

3. **Update PATCH Version:**
   ```bash
   # 0.0.2 → 1.2.1
   ```

4. **Follow Release Process** (Steps 1-7 above)

### Feature Release

1. **Create Feature Branch:**
   ```bash
   git checkout -b feat/new-adapter main
   ```

2. **Implement Feature:** Add adapter, rules, etc.

3. **Update MINOR Version:**
   ```bash
   # 0.0.2 → 1.3.0
   ```

4. **Follow Release Process**

### Breaking Change Release

1. **Create Major Branch:**
   ```bash
   git checkout -b breaking/usn-v2 main
   ```

2. **Implement Breaking Changes:** Document migration path

3. **Update MAJOR Version:**
   ```bash
   # 0.0.2 → 2.0.0
   ```

4. **Write Migration Guide** in `CHANGELOG.md`

5. **Follow Release Process**

---

## Rollback Procedure

### If Release Has Critical Bug

1. **Delete Tag:**
   ```bash
   git tag -d v0.0.2
   git push origin :refs/tags/v0.0.2
   ```

2. **Delete Release:**
   - Go to GitHub Releases
   - Delete release v0.0.2
   - Delete associated binaries

3. **Fix Issues:**
   ```bash
   git checkout -b hotfix/critical-bug main
   # Implement fix
   ```

4. **Create New Release:**
   ```bash
   git tag v1.2.1
   git push origin v1.2.1
   ```

### Notify Users

Post announcement in release notes:

```markdown
## Deprecation Notice

Version v0.0.2 has been deprecated due to [issue description].
Please upgrade to v1.2.1 immediately.
```

---

## Monitoring & Analytics

### Download Tracking

GitHub Releases provides download counts:
- Track per-binary downloads
- Monitor platform distribution
- Identify popular versions

### Error Reporting

Implement crash reporting (future):

```go
// cli/main.go
func main() {
    defer func() {
        if r := recover(); r != nil {
            // Send to Sentry/Crashlytics
            reportCrash(r)
            os.Exit(1)
        }
    }()
    
    // ... rest of main
}
```

### User Feedback

Monitor:
- GitHub Issues
- Discussions
- Social media mentions

---

## Security Updates

### Vulnerability Response

1. **Receive Report:** Via SECURITY.md process
2. **Assess Severity:** Critical, High, Medium, Low
3. **Develop Fix:** In private branch
4. **Test Thoroughly:** Ensure no regressions
5. **Release Patch:** Expedited release process
6. **Public Disclosure:** After fix is available

### Dependency Updates

```bash
# Check for outdated dependencies
go list -u -m all

# Update dependencies
go get -u ./...

# Verify tests still pass
go test ./...
```

---

## Release Checklist

### Pre-Release

- [ ] All tests passing
- [ ] Code formatted (`go fmt ./...`)
- [ ] Linters clean (`go vet ./...`)
- [ ] CHANGELOG.md updated
- [ ] Version bumped
- [ ] Examples tested
- [ ] Documentation updated

### Release

- [ ] Git tag created
- [ ] GoReleaser build successful
- [ ] GitHub Release published
- [ ] Binaries verified
- [ ] Installers tested

### Post-Release

- [ ] Smoke tests passing
- [ ] MCP registration verified
- [ ] User announcement posted
- [ ] Documentation deployed
- [ ] Package managers updated (if applicable)

---

## Troubleshooting

### Issue: GoReleaser Fails

**Symptoms:**
```
release failed: configuration error: invalid yaml
```

**Solution:**
```bash
# Validate config
goreleaser check

# Test build
goreleaser release --snapshot --clean
```

### Issue: Binary Not Found After Install

**Symptoms:**
```
bash: a11ysentry: command not found
```

**Solution:**
```bash
# Verify installation directory
which a11ysentry

# Check PATH
echo $PATH

# Re-run installer
curl -sSL https://.../install.sh | bash
```

### Issue: MCP Registration Fails

**Symptoms:**
```
⚠️  Registration finished with some warnings
```

**Solution:**
```bash
# Check if AI agent is installed
ls -la ~/Library/Application\ Support/Claude/

# Manually add config
cat >> ~/Library/Application\ Support/Claude/claude_desktop_config.json <<EOF
{
  "mcpServers": {
    "a11ysentry": {
      "command": "a11ysentry",
      "args": ["mcp"]
    }
  }
}
EOF
```

---

## Resources

- [GoReleaser Documentation](https://goreleaser.com/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Semantic Versioning Spec](https://semver.org/)
- [CHANGELOG.md Guidelines](https://keepachangelog.com/)
F
```

---

## Resources

- [GoReleaser Documentation](https://goreleaser.com/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Semantic Versioning Spec](https://semver.org/)
- [CHANGELOG.md Guidelines](https://keepachangelog.com/)
