# Design: Initialize SpecHarbor CLI Project

## Overview

The initial implementation creates a minimal but extensible Go CLI. The project is organized by package responsibility so future OpenSpec changes can add behavior incrementally.

## Package layout

```text
cmd/specharbor       CLI entrypoint
internal/cli         command routing
internal/config      configuration management
internal/scanner     project stack detection
internal/generator   spec generation strategies
internal/validator   OpenSpec validation
internal/prompt      agent prompt generation
internal/ai          provider integrations
internal/openspec    OpenSpec file operations
internal/archive     archive workflow
internal/review      review workflow
```

## CLI command placeholders

- init
- scan
- generate
- prompt
- validate
- review
- archive
- config
- version
- help

## Rationale

Keeping the first implementation small allows the project to compile and gives future changes clear boundaries.
