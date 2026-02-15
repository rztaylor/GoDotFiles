package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rztaylor/GoDotFiles/internal/platform"
)

// ConditionEvaluator is an interface for mockability, though simple function might suffice.
// Let's stick to function for now.

// EvaluateCondition checks if a condition string matches the current platform.
// Supports boolean expressions with parentheses:
//   - "os == 'linux' OR os == 'wsl'"
//   - "(os == 'linux' OR os == 'wsl') AND arch == 'amd64'"
//
// Predicates support fields: os, distro, hostname, arch
// Predicates support operators: ==, !=, =~ (regex)
func EvaluateCondition(condition string, p *platform.Platform) (bool, error) {
	tokens, err := tokenizeCondition(condition)
	if err != nil {
		return false, err
	}

	parser := &conditionParser{tokens: tokens, platform: p}
	result, err := parser.parseExpr()
	if err != nil {
		return false, err
	}
	if parser.pos != len(parser.tokens) {
		return false, fmt.Errorf("unexpected token %q", parser.tokens[parser.pos])
	}
	return result, nil
}

func evaluatePredicate(field, op, value string, p *platform.Platform) (bool, error) {
	value = strings.Trim(value, "'\"")

	var actual string
	switch field {
	case "os":
		actual = p.OS
	case "distro":
		actual = p.Distro
	case "hostname":
		actual = p.Hostname
	case "arch":
		actual = p.Arch
	default:
		return false, fmt.Errorf("unknown field: %s", field)
	}

	switch op {
	case "==":
		return actual == value, nil
	case "!=":
		return actual != value, nil
	case "=~":
		matched, err := regexp.MatchString(value, actual)
		return matched, err
	default:
		return false, fmt.Errorf("unknown operator: %s", op)
	}
}

type conditionParser struct {
	tokens   []string
	pos      int
	platform *platform.Platform
}

// expr := term { OR term }
func (p *conditionParser) parseExpr() (bool, error) {
	left, err := p.parseTerm()
	if err != nil {
		return false, err
	}

	for p.hasNext() {
		tok := strings.ToUpper(p.peek())
		if tok != "OR" {
			break
		}
		p.pos++

		right, err := p.parseTerm()
		if err != nil {
			return false, err
		}
		left = left || right
	}

	return left, nil
}

// term := factor { AND factor }
func (p *conditionParser) parseTerm() (bool, error) {
	left, err := p.parseFactor()
	if err != nil {
		return false, err
	}

	for p.hasNext() {
		tok := strings.ToUpper(p.peek())
		if tok != "AND" {
			break
		}
		p.pos++

		right, err := p.parseFactor()
		if err != nil {
			return false, err
		}
		left = left && right
	}

	return left, nil
}

// factor := '(' expr ')' | predicate
func (p *conditionParser) parseFactor() (bool, error) {
	if !p.hasNext() {
		return false, fmt.Errorf("unexpected end of condition")
	}

	if p.peek() == "(" {
		p.pos++
		value, err := p.parseExpr()
		if err != nil {
			return false, err
		}
		if !p.hasNext() || p.peek() != ")" {
			return false, fmt.Errorf("missing closing parenthesis")
		}
		p.pos++
		return value, nil
	}

	return p.parsePredicate()
}

// predicate := field op value
func (p *conditionParser) parsePredicate() (bool, error) {
	if p.pos+2 >= len(p.tokens) {
		return false, fmt.Errorf("invalid condition format")
	}

	field := p.tokens[p.pos]
	p.pos++
	op := p.tokens[p.pos]
	p.pos++
	value := p.tokens[p.pos]
	p.pos++

	return evaluatePredicate(field, op, value, p.platform)
}

func (p *conditionParser) hasNext() bool {
	return p.pos < len(p.tokens)
}

func (p *conditionParser) peek() string {
	return p.tokens[p.pos]
}

func tokenizeCondition(condition string) ([]string, error) {
	var tokens []string
	runes := []rune(strings.TrimSpace(condition))

	for i := 0; i < len(runes); {
		r := runes[i]

		if r == ' ' || r == '\t' || r == '\n' {
			i++
			continue
		}

		if r == '(' || r == ')' {
			tokens = append(tokens, string(r))
			i++
			continue
		}

		if r == '"' || r == '\'' {
			quote := r
			start := i
			i++
			for i < len(runes) && runes[i] != quote {
				i++
			}
			if i >= len(runes) {
				return nil, fmt.Errorf("unterminated quoted value in condition")
			}
			i++
			tokens = append(tokens, string(runes[start:i]))
			continue
		}

		start := i
		for i < len(runes) {
			cur := runes[i]
			if cur == ' ' || cur == '\t' || cur == '\n' || cur == '(' || cur == ')' {
				break
			}
			i++
		}
		tokens = append(tokens, string(runes[start:i]))
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty condition")
	}

	return tokens, nil
}

// CheckConditions evaluates a list of ProfileConditions against the platform.
func CheckConditions(conditions []ProfileCondition, p *platform.Platform) (includes, includeApps, excludeApps []string, err error) {
	for _, c := range conditions {
		match, evalErr := EvaluateCondition(c.If, p)
		if evalErr != nil {
			return nil, nil, nil, evalErr
		}
		if match {
			includes = append(includes, c.Includes...)
			includeApps = append(includeApps, c.IncludeApps...)
			excludeApps = append(excludeApps, c.ExcludeApps...)
		}
	}
	return
}
