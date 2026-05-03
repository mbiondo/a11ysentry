# Skill Registry - A11ySentry

## Project Standards

### Universal Semantic Node (USN)
All platform-specific UI elements MUST be mapped to the USN schema. This allows the engine to apply unified WCAG rules across Web, Mobile, Desktop, and Gaming. USN nodes must include:
- `Role`: Normalized semantic role.
- `Label`: Accessible name.
- `Traits`: Platform attributes (styles, ARIA, props).
- `Source`: Accurate file, line, and column info.

### Compliance
Validators MUST strictly follow WCAG 2.2 / Section 508 standards. Logic MUST be deterministic and binary (Pass/Fail) based on engineering data, never LLM inference.
- **Hierarchy Rule**: Heading levels must not skip.
- **Contrast Rule**: Minimum 4.5:1 for text, 3:1 for UI components.
- **Context Rule**: Accessibility states (hidden, disabled) must propagate recursively down the component tree.


## User Skills
- [a11ysentry-mcp](.agents/skills/a11ysentry-mcp/SKILL.md): Expert guidance for auditing and fixing accessibility violations.

