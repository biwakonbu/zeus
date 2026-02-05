package generator

import "embed"

// 配布対象エージェント・スキル
// go generate でプロジェクトルートの .claude/ からコピーされる
//
//go:generate sh -c "mkdir -p assets/agents assets/skills/zeus-suggest assets/skills/zeus-risk-analysis assets/skills/zeus-e2e-tester"
//go:generate sh -c "cp ../../.claude/agents/zeus-orchestrator.md assets/agents/"
//go:generate sh -c "cp ../../.claude/agents/zeus-planner.md assets/agents/"
//go:generate sh -c "cp ../../.claude/agents/zeus-reviewer.md assets/agents/"
//go:generate sh -c "cp ../../.claude/skills/zeus-suggest/SKILL.md assets/skills/zeus-suggest/"
//go:generate sh -c "cp ../../.claude/skills/zeus-risk-analysis/SKILL.md assets/skills/zeus-risk-analysis/"
//go:generate sh -c "cp -r ../../.claude/skills/zeus-e2e-tester/. assets/skills/zeus-e2e-tester/"

//go:embed assets/agents/*.md
var agentFS embed.FS

//go:embed assets/skills/zeus-suggest/SKILL.md
//go:embed assets/skills/zeus-risk-analysis/SKILL.md
//go:embed assets/skills/zeus-e2e-tester/SKILL.md
var skillFS embed.FS
