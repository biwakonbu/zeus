package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/biwakonbu/zeus/internal/core"
	"github.com/biwakonbu/zeus/internal/doctor"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "システムの健全性を診断",
	RunE:  runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	zeus := getZeus(cmd)

	// IntegrityChecker を作成（ハンドラーがある場合）
	var d *doctor.Doctor
	if zeus != nil {
		checker := createIntegrityChecker(zeus)
		d = doctor.NewWithIntegrity(".", checker)
	} else {
		d = doctor.New(".")
	}

	result, err := d.Diagnose(ctx)
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Println(cyan("Zeus Doctor - System Diagnosis"))
	fmt.Println("═══════════════════════════════════════════════════════════")

	for _, check := range result.Checks {
		var icon string
		switch check.Status {
		case "pass":
			icon = color.GreenString("✓")
		case "warn":
			icon = color.YellowString("⚠")
		case "fail":
			icon = color.RedString("✗")
		}
		fmt.Printf("%s %s: %s\n", icon, check.Check, check.Message)
	}

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("Overall: %s\n", result.Overall)

	if result.FixableCount > 0 {
		fmt.Printf("\n%d issue(s) can be fixed automatically. Run 'zeus fix' to repair.\n", result.FixableCount)
	}

	return nil
}

// createIntegrityChecker は Zeus インスタンスから IntegrityChecker を作成
// 全てのハンドラーは任意であり、取得できたものだけが設定される
func createIntegrityChecker(zeus *core.Zeus) *core.IntegrityChecker {
	registry := zeus.GetRegistry()
	if registry == nil {
		return nil
	}

	// Objective ハンドラーを取得（nil でも IntegrityChecker は動作する）
	var objH *core.ObjectiveHandler
	if objHandler, ok := registry.Get("objective"); ok {
		objH, _ = objHandler.(*core.ObjectiveHandler)
	}

	checker := core.NewIntegrityChecker(objH)

	// Consideration ハンドラーを設定
	if conHandler, ok := registry.Get("consideration"); ok {
		if conH, ok := conHandler.(*core.ConsiderationHandler); ok {
			checker.SetConsiderationHandler(conH)
		}
	}

	// Decision ハンドラーを設定
	if decHandler, ok := registry.Get("decision"); ok {
		if decH, ok := decHandler.(*core.DecisionHandler); ok {
			checker.SetDecisionHandler(decH)
		}
	}

	// Problem ハンドラーを設定
	if probHandler, ok := registry.Get("problem"); ok {
		if probH, ok := probHandler.(*core.ProblemHandler); ok {
			checker.SetProblemHandler(probH)
		}
	}

	// Risk ハンドラーを設定
	if riskHandler, ok := registry.Get("risk"); ok {
		if riskH, ok := riskHandler.(*core.RiskHandler); ok {
			checker.SetRiskHandler(riskH)
		}
	}

	// Assumption ハンドラーを設定
	if assumHandler, ok := registry.Get("assumption"); ok {
		if assumH, ok := assumHandler.(*core.AssumptionHandler); ok {
			checker.SetAssumptionHandler(assumH)
		}
	}

	// Quality ハンドラーを設定
	if qualHandler, ok := registry.Get("quality"); ok {
		if qualH, ok := qualHandler.(*core.QualityHandler); ok {
			checker.SetQualityHandler(qualH)
		}
	}

	// UseCase ハンドラーを設定
	if ucHandler, ok := registry.Get("usecase"); ok {
		if ucH, ok := ucHandler.(*core.UseCaseHandler); ok {
			checker.SetUseCaseHandler(ucH)
		}
	}

	// Subsystem ハンドラーを設定
	if subHandler, ok := registry.Get("subsystem"); ok {
		if subH, ok := subHandler.(*core.SubsystemHandler); ok {
			checker.SetSubsystemHandler(subH)
		}
	}

	// Activity ハンドラーを設定
	if actHandler, ok := registry.Get("activity"); ok {
		if actH, ok := actHandler.(*core.ActivityHandler); ok {
			checker.SetActivityHandler(actH)
		}
	}

	// Actor ハンドラーを設定
	if actorHandler, ok := registry.Get("actor"); ok {
		if actorH, ok := actorHandler.(*core.ActorHandler); ok {
			checker.SetActorHandler(actorH)
		}
	}

	return checker
}
