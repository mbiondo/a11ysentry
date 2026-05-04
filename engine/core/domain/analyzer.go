package domain

import (
	"context"
	"sync"
)

type accessibilityAnalyzer struct {
	rules []Rule
}

// NewAnalyzer returns a new accessibility analyzer with default rules.
func NewAnalyzer() Analyzer {
	return &accessibilityAnalyzer{
		rules: getDefaultRules(),
	}
}

func (a *accessibilityAnalyzer) Analyze(ctx context.Context, nodes []USN, cfg ProjectConfig) ([]Violation, error) {
	analysisCtx := BuildAnalysisContext(nodes, cfg)
	var allViolations []Violation
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, rule := range a.rules {
		// Skip disabled rules
		if ruleCfg, found := cfg.Rules[rule.ErrorCode()]; found && !ruleCfg.Enabled {
			continue
		}

		wg.Add(1)
		go func(r Rule) {
			defer wg.Done()
			violations, err := r.Execute(ctx, analysisCtx)
			if err == nil {
				mu.Lock()
				// Apply severity overrides from config and filter ignored rules
				if ruleCfg, found := cfg.Rules[r.ErrorCode()]; found && ruleCfg.Severity != "" {
					for i := range violations {
						violations[i].Severity = ruleCfg.Severity
					}
				}

				// Filter violations based on inline a11y-ignore comments
				var filtered []Violation
				for _, v := range violations {
					ignored := false
					for _, ignoredCode := range v.SourceRef.IgnoredRules {
						if ignoredCode == v.ErrorCode || ignoredCode == "all" {
							ignored = true
							break
						}
					}
					if !ignored {
						filtered = append(filtered, v)
					}
				}

				allViolations = append(allViolations, filtered...)
				mu.Unlock()
			}
		}(rule)
	}

	wg.Wait()

	return deduplicateViolations(allViolations), nil
}

func deduplicateViolations(violations []Violation) []Violation {
	seen := make(map[string]bool)
	unique := violations[:0]
	for _, v := range violations {
		key := v.ErrorCode + "|" + v.SourceRef.RawHTML
		if !seen[key] {
			seen[key] = true
			unique = append(unique, v)
		}
	}
	return unique
}

func getDefaultRules() []Rule {
	return []Rule{
		&ruleWCAG111{},
		&ruleWCAG131{},
		&ruleWCAG135{},
		&ruleWCAG141{},
		&ruleWCAG143{},
		&ruleWCAG1411{},
		&ruleWCAG244{},
		&ruleWCAG246{},
		&ruleAccessibleNames{},
		&ruleParsing{},
		&ruleLandmarks{},
		&ruleFocus{},
		&ruleWCAG413{},
		&ruleTouchTargets{},
		&rulePageLevel{},
		&ruleBestPractices{},
	}
}
