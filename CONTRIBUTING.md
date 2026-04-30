# Contributing to A11ySentry

First off, thank you for considering contributing to A11ySentry! It's people like you that make it a universal standard.

## 📜 Code of Conduct
By participating in this project, you agree to abide by our professional and inclusive standards.

## 🛠️ Development Workflow

### 1. Conventional Commits
We strictly follow [Conventional Commits](https://www.conventionalcommits.org/). This ensures a clean and automated changelog.
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation changes
- `style`: Formatting, missing semi colons, etc; no code change
- `refactor`: Refactoring production code
- `test`: Adding missing tests, refactoring tests; no production code change
- `chore`: Updating build tasks, package manager configs, etc; no production code change

**Example:** `feat(web): add support for aria-describedby validation`

### 2. Branching Strategy
- Main branch: `main`
- Feature branches: `feat/feature-name`
- Bug fixes: `fix/issue-description`

### 3. Pull Request Process
1. Create a new branch.
2. Ensure all tests pass: `go test ./...`.
3. Add tests for your changes.
4. Open a PR with a clear description of the "What" and "Why".
5. Wait for the A11ySentry CI and maintainer review.

## 🐛 Issues
- Use the provided **Bug Report** or **Feature Request** templates.
- Always include a code snippet or a file from `/examples` that reproduces the issue.

## 🏗️ Architecture Philosophy
- **USN First**: All platform-specific code must map to a `Universal Semantic Node`.
- **Deterministic Logic**: Rules should be binary (Pass/Fail) and based on WCAG standards.
- **Performance**: Adapters must be fast (< 100ms per file).
