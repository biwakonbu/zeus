package generator

import "embed"

// 配布対象エージェント・スキル
// go generate でプロジェクトルートの .claude/ からコピーされる
//
//go:generate sh -c "mkdir -p assets/agents assets/skills/zeus-project-scan assets/skills/zeus-activity-suggest assets/skills/zeus-risk-analysis assets/skills/zeus-wbs-design"
//go:generate sh -c "cp ../../.claude/agents/zeus-orchestrator.md assets/agents/"
//go:generate sh -c "cp ../../.claude/agents/zeus-planner.md assets/agents/"
//go:generate sh -c "cp ../../.claude/agents/zeus-reviewer.md assets/agents/"
//go:generate sh -c "cp ../../.claude/skills/zeus-project-scan/SKILL.md assets/skills/zeus-project-scan/"
//go:generate sh -c "cp ../../.claude/skills/zeus-activity-suggest/SKILL.md assets/skills/zeus-activity-suggest/"
//go:generate sh -c "cp ../../.claude/skills/zeus-risk-analysis/SKILL.md assets/skills/zeus-risk-analysis/"
//go:generate sh -c "cp ../../.claude/skills/zeus-wbs-design/SKILL.md assets/skills/zeus-wbs-design/"

//go:embed assets/agents/*.md
var agentFS embed.FS

//go:embed assets/skills/zeus-project-scan/SKILL.md
//go:embed assets/skills/zeus-activity-suggest/SKILL.md
//go:embed assets/skills/zeus-risk-analysis/SKILL.md
//go:embed assets/skills/zeus-wbs-design/SKILL.md
var skillFS embed.FS
