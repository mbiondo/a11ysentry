# Documentation Index

Welcome to A11ySentry's comprehensive documentation. This index will help you find the right documentation for your needs.

---

## 🚀 Getting Started

### New to A11ySentry?
Start here if you're just learning about the project:

1. **[Main README](../README.md)** - Overview, vision, and quick start
2. **[Installation](../README.md#-installation)** - Install on Windows, macOS, or Linux
3. **[Usage](../README.md#-usage)** - Basic CLI and TUI commands
4. **[Examples](../examples/README.md)** - See A11ySentry in action

### Quick Reference

```bash
# Analyze a file
a11ysentry path/to/file.html

# Open TUI dashboard
a11ysentry --tui

# Register MCP in AI agents
a11ysentry mcp --register

# Output as JSON
a11ysentry --format json file.html
```

---

## 📖 Documentation by Topic

### Architecture & Design

| Document | Purpose | Audience |
|----------|---------|----------|
| **[Architecture Deep Dive](./ARCHITECTURE_DEEP_DIVE.md)** | Detailed architecture, pipeline stages, data flow | Developers, Architects |
| **[API Reference](./API_REFERENCE.md)** | Core types, interfaces, CLI commands | All Users |
| **[USN Specification](./API_REFERENCE.md#universal-semantic-node-usn)** | Universal Semantic Node schema | Adapter Developers |

**Key Concepts:**
- Adapter-Hub Pattern
- Universal Semantic Node (USN)
- Four-stage pipeline (Ingestion → Normalization → Analysis → Emission)
- Stateless analysis

### Development & Contribution

| Document | Purpose | Audience |
|----------|---------|----------|
| **[Developer Guide](./DEVELOPER_GUIDE.md)** | How to build, test, and extend | Contributors |
| **[Contributing Guidelines](../CONTRIBUTING.md)** | Git workflow, commit conventions | All Contributors |
| **[Code of Conduct](../CODE_OF_CONDUCT.md)** | Community standards | Everyone |
| **[Examples Documentation](../examples/README.md)** | Platform-specific patterns | Testers, Contributors |

**Common Tasks:**
- [Creating a New Adapter](./DEVELOPER_GUIDE.md#creating-a-new-adapter)
- [Adding a New WCAG Rule](./DEVELOPER_GUIDE.md#adding-a-new-wcag-rule)
- [Running Tests](./DEVELOPER_GUIDE.md#testing-guidelines)
- [USN Mapping Examples](./DEVELOPER_GUIDE.md#usn-mapping-examples)

### Integration & Usage

| Document | Purpose | Audience |
|----------|---------|----------|
| **[MCP Integration Guide](./MCP_INTEGRATION.md)** | AI agent setup and tools | AI Agent Users |
| **[CLI Commands](./API_REFERENCE.md#cli-commands)** | Command-line usage | All Users |
| **[TUI Dashboard](./API_REFERENCE.md#tui-dashboard)** | Interactive dashboard | All Users |
| [**Output Formats**](./API_REFERENCE.md#output-formats) | Text, JSON, JSON-LD | CI/CD Engineers |

**AI Agent Integration:**
- [Supported Agents](./MCP_INTEGRATION.md#supported-ai-agents)
- [Manual Configuration](./MCP_INTEGRATION.md#manual-configuration)
- [Available Tools](./MCP_INTEGRATION.md#available-tools)
- [Troubleshooting](./MCP_INTEGRATION.md#troubleshooting)

### Deployment & Maintenance

| Document | Purpose | Audience |
|----------|---------|----------|
| **[Release Guide](./RELEASE.md)** | Build, tag, and publish | Maintainers |
| **[Security Policy](../SECURITY.md)** | Vulnerability reporting | Security Researchers |
| **[Changelog](../CHANGELOG.md)** | Version history | All Users |
| **[License](../LICENSE)** | MIT License | Legal/Compliance |

**Release Tasks:**
- [Version Numbering](./RELEASE.md#versioning)
- [GoReleaser Setup](./RELEASE.md#build-configuration)
- [CI/CD Pipeline](./RELEASE.md#cicd-pipeline)
- [Rollback Procedure](./RELEASE.md#rollback-procedure)

---

## 🎯 Documentation by Role

### I'm a...

#### Developer / Contributor

**Start with:**
1. [Developer Guide](./DEVELOPER_GUIDE.md)
2. [API Reference](./API_REFERENCE.md)
3. [Examples](../examples/README.md)

**Then explore:**
- [Creating Adapters](./DEVELOPER_GUIDE.md#creating-a-new-adapter)
- [Adding Rules](./DEVELOPER_GUIDE.md#adding-a-new-wcag-rule)
- [Testing Guidelines](./DEVELOPER_GUIDE.md#testing-guidelines)

#### AI Agent User

**Start with:**
1. [MCP Integration Guide](./MCP_INTEGRATION.md)
2. [Usage - AI Agents](../README.md#ai-agents--mcp)

**Then explore:**
- [Supported Agents](./MCP_INTEGRATION.md#supported-ai-agents)
- [Available Tools](./MCP_INTEGRATION.md#available-tools)
- [Integration Examples](./MCP_INTEGRATION.md#integration-examples)

#### DevOps / Release Engineer

**Start with:**
1. [Release Guide](./RELEASE.md)
2. [Installation Scripts](../install.ps1, ../install.sh)

**Then explore:**
- [GoReleaser Configuration](./RELEASE.md#build-configuration)
- [CI/CD Pipeline](./RELEASE.md#cicd-pipeline)
- [Distribution Channels](./RELEASE.md#distribution-channels)

#### Architect / Technical Lead

**Start with:**
1. [Architecture Deep Dive](./ARCHITECTURE_DEEP_DIVE.md)
2. [API Reference](./API_REFERENCE.md)

**Then explore:**
- [Design Decisions](./ARCHITECTURE_DEEP_DIVE.md#design-decisions)
- [Performance Characteristics](./ARCHITECTURE_DEEP_DIVE.md#performance-characteristics)
- [Future Considerations](./ARCHITECTURE_DEEP_DIVE.md#future-architecture-considerations)

#### End User / QA Engineer

**Start with:**
1. [Main README](../README.md)
2. [CLI Commands](./API_REFERENCE.md#cli-commands)

**Then explore:**
- [Output Formats](./API_REFERENCE.md#output-formats)
- [Error Codes](./API_REFERENCE.md#error-codes)
- [Examples](../examples/README.md)

---

## 📚 Document Types

### Tutorials (Step-by-Step)

- [Getting Started](../README.md#-installation)
- [Creating a New Adapter](./DEVELOPER_GUIDE.md#creating-a-new-adapter)
- [Adding a WCAG Rule](./DEVELOPER_GUIDE.md#adding-a-new-wcag-rule)
- [MCP Registration](./MCP_INTEGRATION.md#installation)

### How-To Guides (Task-Oriented)

- [Analyze Files](./API_REFERENCE.md#direct-analysis)
- [Use TUI Dashboard](./API_REFERENCE.md#tui-dashboard)
- [Configure AI Agents](./MCP_INTEGRATION.md#manual-configuration)
- [Run Tests](./DEVELOPER_GUIDE.md#testing-guidelines)

### Reference (Technical Details)

- [API Reference](./API_REFERENCE.md) - Types, interfaces, commands
- [Error Codes](./API_REFERENCE.md#error-codes) - WCAG violations
- [Platforms](./API_REFERENCE.md#platforms) - Supported platforms
- [Semantic Roles](./API_REFERENCE.md#semantic-roles) - USN roles

### Explanations (Concepts & Theory)

- [Architecture Deep Dive](./ARCHITECTURE_DEEP_DIVE.md) - Why and how
- [USN Concept](./ARCHITECTURE_DEEP_DIVE.md#stage-2-normalization-usn-mapping) - Abstraction layer
- [Design Decisions](./ARCHITECTURE_DEEP_DIVE.md#design-decisions) - Trade-offs
- [WCAG Standards](https://www.w3.org/WAI/WCAG22/quickref/) - External resource

---

## 🔍 Quick Navigation

### By File Path

```
docs/
├── API_REFERENCE.md           # Types, interfaces, CLI
├── ARCHITECTURE_DEEP_DIVE.md  # Architecture details
├── DEVELOPER_GUIDE.md         # How to contribute
├── MCP_INTEGRATION.md         # AI agent setup
├── RELEASE.md                 # Build & deploy
├── INDEX.md                   # This file
├── architecture.md            # High-level overview
└── standards.md               # Project standards
```

### By Topic

**Accessibility:**
- [WCAG Rules](./API_REFERENCE.md#error-codes)
- [USN Mapping](./DEVELOPER_GUIDE.md#usn-mapping-examples)
- [Platform Support](../README.md#-key-features)

**Development:**
- [Go Code Structure](./DEVELOPER_GUIDE.md#project-structure)
- [Testing](./DEVELOPER_GUIDE.md#testing-guidelines)
- [Debugging](./DEVELOPER_GUIDE.md#debugging-tips)

**Deployment:**
- [Building](./RELEASE.md#build--run)
- [Releasing](./RELEASE.md#release-process)
- [Installers](./RELEASE.md#installation-scripts)

---

## 🆘 Getting Help

### Documentation Not Enough?

1. **Check Examples:** [examples/README.md](../examples/README.md)
2. **Search Issues:** [GitHub Issues](https://github.com/mbiondo/a11ysentry/issues)
3. **Ask in Discussions:** [GitHub Discussions](https://github.com/mbiondo/a11ysentry/discussions)
4. **Report Bug:** [New Issue](https://github.com/mbiondo/a11ysentry/issues/new)

### Common Questions

**Q: How do I add support for a new platform?**  
A: See [Creating a New Adapter](./DEVELOPER_GUIDE.md#creating-a-new-adapter)

**Q: Which AI agents are supported?**  
A: See [Supported AI Agents](./MCP_INTEGRATION.md#supported-ai-agents)

**Q: How do I run tests?**  
A: See [Testing Guidelines](./DEVELOPER_GUIDE.md#testing-guidelines)

**Q: What WCAG rules are implemented?**  
A: See [Error Codes](./API_REFERENCE.md#error-codes)

**Q: How do I release a new version?**  
A: See [Release Process](./RELEASE.md#release-process)

---

## 📝 Contributing to Documentation

Found a typo? Want to improve documentation?

1. **Fork the repository**
2. **Edit documentation files**
3. **Submit PR** with clear description

**Documentation Conventions:**
- Use Markdown formatting
- Include code examples
- Link to related documents
- Keep language clear and concise

---

## 🔗 External Resources

### WCAG & Accessibility

- [WCAG 2.2 Guidelines](https://www.w3.org/WAI/WCAG22/quickref/)
- [WAI-ARIA Authoring Practices](https://www.w3.org/WAI/ARIA/apg/)
- [WebAIM Checklist](https://webaim.org/standards/wcag/checklist)

### Platform Accessibility

- [Web (MDN)](https://developer.mozilla.org/en-US/docs/Web/Accessibility)
- [Android](https://developer.android.com/guide/topics/ui/accessibility)
- [iOS](https://developer.apple.com/documentation/accessibility)
- [Flutter](https://docs.flutter.dev/development/ui/accessibility)
- [React Native](https://reactnative.dev/docs/accessibility)

### Tools & Technologies

- [Model Context Protocol](https://modelcontextprotocol.io/)
- [Bubbletea TUI Framework](https://github.com/charmbracelet/bubbletea)
- [GoReleaser](https://goreleaser.com/)

---

## 📊 Documentation Status

| Document | Status | Last Updated |
|----------|--------|--------------|
| API Reference | ✅ Complete | 2026-04-30 |
| Architecture Deep Dive | ✅ Complete | 2026-04-30 |
| Developer Guide | ✅ Complete | 2026-04-30 |
| MCP Integration Guide | ✅ Complete | 2026-04-30 |
| Release Guide | ✅ Complete | 2026-04-30 |
| Examples Documentation | ✅ Complete | 2026-04-30 |
| Architecture (High-Level) | ✅ Complete | 2026-04-30 |
| Standards | ✅ Complete | 2026-04-30 |

**Legend:**
- ✅ Complete: Comprehensive coverage
- ⚠️ Needs Update: Requires expansion
- ❌ Missing: Not yet created

---

## 🎯 Next Steps

**Pick your path:**

- **New User?** → [Main README](../README.md)
- **Developer?** → [Developer Guide](./DEVELOPER_GUIDE.md)
- **AI Agent User?** → [MCP Integration](./MCP_INTEGRATION.md)
- **Release Manager?** → [Release Guide](./RELEASE.md)
- **Architect?** → [Architecture Deep Dive](./ARCHITECTURE_DEEP_DIVE.md)

---

**Happy documenting! 📚✨**
