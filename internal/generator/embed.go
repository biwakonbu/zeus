package generator

import "embed"

// 配布対象エージェント・スキル
// テンプレート正本は assets/ 配下。{{.ProjectName}} 等のテンプレート変数を含む。
// zeus update-claude が assets/ のテンプレートを展開して .claude/ に出力する。
// 注意: .claude/ は生成物であり、assets/ にコピーバックしてはならない。

//go:embed assets/agents/*.md
var agentFS embed.FS

//go:embed assets/skills/zeus-suggest/SKILL.md
//go:embed assets/skills/zeus-risk-analysis/SKILL.md
//go:embed assets/skills/zeus-e2e-tester/SKILL.md
var skillFS embed.FS
