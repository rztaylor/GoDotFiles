package engine

import (
	"regexp"
	"strings"

	"github.com/rztaylor/GoDotFiles/internal/apps"
)

// RiskFinding describes a high-risk command found in configuration.
type RiskFinding struct {
	App      string
	Location string
	Command  string
	Reason   string
}

var highRiskPatterns = []struct {
	re     *regexp.Regexp
	reason string
}{
	{
		re:     regexp.MustCompile(`(?i)\b(curl|wget)\b[^\n|;]*\|\s*(sh|bash|zsh)\b`),
		reason: "pipes remote content directly into a shell",
	},
	{
		re:     regexp.MustCompile(`(?i)\b(bash|sh|zsh)\b\s+-c\s+.*\b(curl|wget)\b`),
		reason: "executes downloaded content via shell -c",
	},
	{
		re:     regexp.MustCompile(`(?i)\$\(.*\b(curl|wget)\b.*\)`),
		reason: "uses command substitution with remote content",
	},
}

// DetectHighRiskConfigurations scans app bundles for high-risk shell commands.
func DetectHighRiskConfigurations(bundles []*apps.Bundle) []RiskFinding {
	var findings []RiskFinding

	for _, b := range bundles {
		if b == nil {
			continue
		}
		if b.Hooks != nil {
			findings = append(findings, detectCommands(b.Name, "hooks.pre_install", b.Hooks.PreInstall)...)
			findings = append(findings, detectCommands(b.Name, "hooks.post_install", b.Hooks.PostInstall)...)
			findings = append(findings, detectCommands(b.Name, "hooks.pre_link", b.Hooks.PreLink)...)
			findings = append(findings, detectCommands(b.Name, "hooks.post_link", b.Hooks.PostLink)...)
			for _, h := range b.Hooks.Apply {
				if reason, ok := detectHighRiskReason(h.Run); ok {
					findings = append(findings, RiskFinding{
						App:      b.Name,
						Location: "hooks.apply.run",
						Command:  strings.TrimSpace(h.Run),
						Reason:   reason,
					})
				}
			}
		}

		if b.Package != nil && b.Package.Custom != nil && strings.TrimSpace(b.Package.Custom.Script) != "" {
			if reason, ok := detectHighRiskReason(b.Package.Custom.Script); ok {
				findings = append(findings, RiskFinding{
					App:      b.Name,
					Location: "package.custom.script",
					Command:  strings.TrimSpace(b.Package.Custom.Script),
					Reason:   reason,
				})
			}
		}
	}

	return findings
}

func detectCommands(app, location string, commands []string) []RiskFinding {
	var findings []RiskFinding
	for _, cmd := range commands {
		if reason, ok := detectHighRiskReason(cmd); ok {
			findings = append(findings, RiskFinding{
				App:      app,
				Location: location,
				Command:  strings.TrimSpace(cmd),
				Reason:   reason,
			})
		}
	}
	return findings
}

func detectHighRiskReason(command string) (string, bool) {
	cmd := strings.TrimSpace(command)
	if cmd == "" {
		return "", false
	}
	for _, p := range highRiskPatterns {
		if p.re.MatchString(cmd) {
			return p.reason, true
		}
	}
	return "", false
}
