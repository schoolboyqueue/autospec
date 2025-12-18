---
layout: default
title: Home
nav_order: 1
description: "autospec - AI-powered software specification and implementation workflows"
permalink: /
---

# autospec
{: .fs-9 }

Spec-Driven Development Automation
{: .fs-6 .fw-300 }

Stop AI slop. Build features systematically with AI-powered specification workflows.
{: .fs-5 .fw-300 }

[Get Started](/autospec/quickstart){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 .mr-2 }
[View on GitHub](https://github.com/ariel-frischer/autospec){: .btn .fs-5 .mb-4 .mb-md-0 }

---

## What is autospec?

autospec brings structure to AI coding: **spec → plan → tasks → implement** - all in one command.

Built for Claude Code and inspired by [GitHub SpecKit](https://github.com/github/spec-kit), autospec reimagines the specification workflow with **YAML-first artifacts** for programmatic access and validation.

```bash
# Generate everything: spec → plan → tasks → implement
autospec run -a "Add user authentication with OAuth"

# Or step by step with review between stages
autospec run -s "Add feature"  # Generate spec
# Review and edit spec.yaml
autospec run -pti              # Plan, tasks, implement
```

---

## Key Features

<div class="features-grid">

### Automated Workflow Orchestration
{: .text-purple-000 }
Runs stages in dependency order with automatic retry on failure. No manual intervention needed.

### YAML-First Artifacts
{: .text-purple-000 }
Machine-readable `spec.yaml`, `plan.yaml`, `tasks.yaml` for programmatic access and validation.

### Smart Validation
{: .text-purple-000 }
Validates artifacts exist and meet completeness criteria before proceeding to the next stage.

### Session Isolation
{: .text-purple-000 }
Per-phase or per-task execution reduces API costs by 80%+ on large specs through context isolation.

### Configurable Retry Logic
{: .text-purple-000 }
Automatic retries with persistent state tracking. Resume from failures without losing progress.

### Performance Optimized
{: .text-purple-000 }
Sub-second validation (<10ms per check), <50ms startup. Built for speed.

</div>

---

## Quick Start

### Prerequisites

- [Claude Code CLI](https://code.claude.com/docs/en/setup) installed and configured
- Git

### Installation

```bash
curl -fsSL https://raw.githubusercontent.com/ariel-frischer/autospec/main/install.sh | sh
```

### First Workflow

```bash
# Navigate to your project
cd your-project

# Check dependencies
autospec doctor

# Initialize autospec configuration
autospec init

# Create your first specification
autospec run -a "Add user authentication with OAuth"
```

[View Full Quickstart Guide](/autospec/quickstart){: .btn .btn-outline }

---

## The Workflow

autospec runs four core stages in sequence:

| Stage | Command | Creates | Description |
|:------|:--------|:--------|:------------|
| **specify** | `autospec specify "desc"` | `spec.yaml` | Feature specification with requirements |
| **plan** | `autospec plan` | `plan.yaml` | Implementation design and architecture |
| **tasks** | `autospec tasks` | `tasks.yaml` | Actionable task breakdown with dependencies |
| **implement** | `autospec implement` | — | Executes tasks, updates status |

Each artifact is validated before proceeding to the next stage, ensuring quality at every step.

---

## Documentation

{: .fs-6 .fw-300 }

| Section | Description |
|:--------|:------------|
| [Quickstart](/autospec/quickstart) | Get up and running in 5 minutes |
| [Reference](/autospec/reference/) | Complete CLI command reference |
| [Guides](/autospec/guides/) | Configuration, troubleshooting, FAQ |
| [Architecture](/autospec/architecture/) | System design and internals |

---

## What Makes autospec Different?

| Feature | GitHub SpecKit | autospec |
|:--------|:--------------|:---------|
| Output Format | Markdown | **YAML** (machine-readable) |
| Validation | Manual review | **Automatic** with retry logic |
| Context Efficiency | Full prompt each time | **Smart YAML injection** + **phase-isolated sessions** |
| Status Updates | Manual | **Auto-updates** spec.yaml & tasks.yaml |
| Phase Orchestration | Manual | **Automated** with dependencies |
| Session Isolation | Single session | **Per-phase/task** (80%+ cost savings) |
| Implementation | Shell scripts | **Go** (type-safe, single binary) |

---

## License

autospec is distributed under the [MIT License](https://github.com/ariel-frischer/autospec/blob/main/LICENSE).
