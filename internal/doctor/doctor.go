package doctor

import (
	"context"
	"path/filepath"

	"github.com/biwakonbu/zeus/internal/core"
	"github.com/biwakonbu/zeus/internal/yaml"
)

// CheckResult は診断結果
type CheckResult struct {
	Check   string
	Status  string // pass, warn, fail
	Message string
	Fixable bool
	FixFunc func(ctx context.Context) error
}

// DiagnosisResult は診断結果全体
type DiagnosisResult struct {
	Overall      string // healthy, degraded, unhealthy
	Checks       []CheckResult
	FixableCount int
}

// FixResult は修復結果
type FixResult struct {
	Fixes  []FixAction
	DryRun bool
}

// FixAction は修復アクション
type FixAction struct {
	Action   string
	Executed bool
}

// Doctor は診断・修復を行う
type Doctor struct {
	zeusPath         string
	fileManager      *yaml.FileManager
	integrityChecker *core.IntegrityChecker
}

// New は新しい Doctor を作成
func New(projectPath string) *Doctor {
	zeusPath := filepath.Join(projectPath, ".zeus")
	return &Doctor{
		zeusPath:    zeusPath,
		fileManager: yaml.NewFileManager(zeusPath),
	}
}

// NewWithIntegrity は IntegrityChecker を含む Doctor を作成
func NewWithIntegrity(projectPath string, checker *core.IntegrityChecker) *Doctor {
	zeusPath := filepath.Join(projectPath, ".zeus")
	return &Doctor{
		zeusPath:         zeusPath,
		fileManager:      yaml.NewFileManager(zeusPath),
		integrityChecker: checker,
	}
}

// Diagnose はシステムを診断（Context対応）
func (d *Doctor) Diagnose(ctx context.Context) (*DiagnosisResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	checks := []CheckResult{}

	// 設定ファイル存在チェック
	checks = append(checks, d.checkConfigExists(ctx))

	// タスクファイル存在チェック
	checks = append(checks, d.checkTasksExists(ctx))

	// 状態ファイル存在チェック
	checks = append(checks, d.checkStateExists(ctx))

	// 参照整合性チェック（IntegrityChecker が設定されている場合）
	if d.integrityChecker != nil {
		checks = append(checks, d.checkIntegrity(ctx)...)
	}

	// 全体の健全性を計算
	overall := d.calculateOverall(checks)
	fixableCount := 0
	for _, check := range checks {
		if check.Status == "fail" && check.Fixable {
			fixableCount++
		}
	}

	return &DiagnosisResult{
		Overall:      overall,
		Checks:       checks,
		FixableCount: fixableCount,
	}, nil
}

// Fix は問題を修復（Context対応）
func (d *Doctor) Fix(ctx context.Context, dryRun bool) (*FixResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	diagnosis, err := d.Diagnose(ctx)
	if err != nil {
		return nil, err
	}

	fixes := []FixAction{}
	for _, check := range diagnosis.Checks {
		if check.Status == "fail" && check.Fixable && check.FixFunc != nil {
			if dryRun {
				fixes = append(fixes, FixAction{Action: check.Message, Executed: false})
			} else {
				if err := check.FixFunc(ctx); err != nil {
					return nil, err
				}
				fixes = append(fixes, FixAction{Action: check.Message, Executed: true})
			}
		}
	}

	return &FixResult{Fixes: fixes, DryRun: dryRun}, nil
}

func (d *Doctor) checkConfigExists(ctx context.Context) CheckResult {
	if d.fileManager.Exists(ctx, "zeus.yaml") {
		return CheckResult{
			Check:   "config_exists",
			Status:  "pass",
			Message: "zeus.yaml found",
			Fixable: false,
		}
	}

	return CheckResult{
		Check:   "config_exists",
		Status:  "fail",
		Message: "zeus.yaml not found - run 'zeus init' to create",
		Fixable: true,
		FixFunc: func(ctx context.Context) error {
			zeus := core.New(filepath.Dir(d.zeusPath))
			_, err := zeus.Init(ctx)
			return err
		},
	}
}

func (d *Doctor) checkTasksExists(ctx context.Context) CheckResult {
	if d.fileManager.Exists(ctx, "tasks/active.yaml") {
		return CheckResult{
			Check:   "tasks_exists",
			Status:  "pass",
			Message: "Task files found",
			Fixable: false,
		}
	}

	return CheckResult{
		Check:   "tasks_exists",
		Status:  "warn",
		Message: "Task files missing",
		Fixable: true,
		FixFunc: func(ctx context.Context) error {
			taskStore := &core.TaskStore{Tasks: []core.Task{}}
			if err := d.fileManager.EnsureDir(ctx, "tasks"); err != nil {
				return err
			}
			return d.fileManager.WriteYaml(ctx, "tasks/active.yaml", taskStore)
		},
	}
}

func (d *Doctor) checkStateExists(ctx context.Context) CheckResult {
	if d.fileManager.Exists(ctx, "state/current.yaml") {
		return CheckResult{
			Check:   "state_exists",
			Status:  "pass",
			Message: "State file found",
			Fixable: false,
		}
	}

	return CheckResult{
		Check:   "state_exists",
		Status:  "warn",
		Message: "State file missing",
		Fixable: true,
		FixFunc: func(ctx context.Context) error {
			state := &core.ProjectState{
				Timestamp: core.Now(),
				Summary:   core.TaskStats{},
				Health:    core.HealthUnknown,
				Risks:     []string{},
			}
			if err := d.fileManager.EnsureDir(ctx, "state"); err != nil {
				return err
			}
			return d.fileManager.WriteYaml(ctx, "state/current.yaml", state)
		},
	}
}

func (d *Doctor) calculateOverall(checks []CheckResult) string {
	failed := 0
	warned := 0

	for _, check := range checks {
		switch check.Status {
		case "fail":
			failed++
		case "warn":
			warned++
		}
	}

	if failed > 0 {
		return "unhealthy"
	}
	if warned > 0 {
		return "degraded"
	}
	return "healthy"
}

// checkIntegrity は参照整合性をチェック
func (d *Doctor) checkIntegrity(ctx context.Context) []CheckResult {
	var checks []CheckResult

	result, err := d.integrityChecker.CheckAll(ctx)
	if err != nil {
		checks = append(checks, CheckResult{
			Check:   "integrity_check",
			Status:  "fail",
			Message: "Integrity check failed: " + err.Error(),
			Fixable: false,
		})
		return checks
	}

	// 参照エラーのチェック
	if len(result.ReferenceErrors) > 0 {
		for _, refErr := range result.ReferenceErrors {
			checks = append(checks, CheckResult{
				Check:   "reference_integrity",
				Status:  "fail",
				Message: refErr.Error(),
				Fixable: false, // 参照エラーは自動修復不可
			})
		}
	} else {
		checks = append(checks, CheckResult{
			Check:   "reference_integrity",
			Status:  "pass",
			Message: "All entity references are valid",
			Fixable: false,
		})
	}

	// 循環参照のチェック
	if len(result.CycleErrors) > 0 {
		for _, cycleErr := range result.CycleErrors {
			checks = append(checks, CheckResult{
				Check:   "cycle_check",
				Status:  "fail",
				Message: cycleErr.Error(),
				Fixable: false, // 循環参照は自動修復不可
			})
		}
	} else {
		checks = append(checks, CheckResult{
			Check:   "cycle_check",
			Status:  "pass",
			Message: "No circular references detected",
			Fixable: false,
		})
	}

	// 警告のチェック（TASK-015: サブシステム参照チェック追加）
	if len(result.Warnings) > 0 {
		for _, warn := range result.Warnings {
			checks = append(checks, CheckResult{
				Check:   "subsystem_reference",
				Status:  "warn",
				Message: warn.Warning(),
				Fixable: false, // 警告は自動修復不可（参照先の作成は手動で行う）
			})
		}
	} else {
		checks = append(checks, CheckResult{
			Check:   "subsystem_reference",
			Status:  "pass",
			Message: "All UseCase subsystem references are valid",
			Fixable: false,
		})
	}

	return checks
}
