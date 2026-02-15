---
description: Fast project orientation and code discovery
---

1. Print repository and working tree status:
`pwd && git rev-parse --abbrev-ref HEAD && git status --short`

2. List high-signal project files:
`ls -1 README.md AGENTS.md TASKS.md CHANGELOG.md IMPLEMENTATION_PLAN.md docs/architecture/overview.md docs/architecture/components.md docs/reference/cli.md docs/reference/yaml-schemas.md`

3. Map package surface area:
`rg --files internal | sed 's|/[^/]*$||' | sort -u`

4. Find CLI command entry points:
`rg "^func new.*Cmd|Use:\\s+\"" internal/cli`

5. Find primary orchestration entry points:
`rg "func \\(.*\\) (Apply|Restore|Link|Resolve|Install|Run)" internal/engine internal/apps internal/packages`

6. Find open roadmap work:
`sed -n '1,260p' TASKS.md`
