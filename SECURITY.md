# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

We take the security of A11ySentry seriously. If you believe you have found a security vulnerability, please report it to us as described below.

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via email at [INSERT EMAIL] or create a draft security advisory on GitHub.

### What to Include

Please include the following information in your report:

- A description of the vulnerability and its impact
- Steps to reproduce the issue
- Your assessment of the severity (if applicable)
- Any potential mitigations you're aware of

### Response Time

We will acknowledge your report within 48 hours and provide a detailed response within 7 days, including:

- Confirmation that we have received your report
- Our assessment of the severity
- A timeline for when we expect to provide a fix

### Disclosure Policy

We ask that you keep the vulnerability confidential until we have issued a fix and publicly disclosed the issue. We will work with you to ensure we don't negatively impact your users or the broader ecosystem.

## Security Best Practices

When contributing to A11ySentry, please follow these security best practices:

- Never commit secrets, API keys, or credentials to the repository
- Validate all user inputs in the CLI and MCP server
- Keep dependencies up to date
- Follow secure coding practices for Go
