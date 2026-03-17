package main

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// A tiny single-file demo agent supporting two execution styles:
// 1) ReAct: think -> act -> observe in a loop
// 2) Plan-and-Execute: build a short plan first, then execute step by step
//
// It intentionally uses rule-based reasoning instead of an external LLM so it can
// run out-of-the-box. You can later replace the planner/decider with model calls.

type ToolResult struct {
	Name    string
	Input   string
	Output  string
	Success bool
}

type Tool func(string) ToolResult

type Agent struct {
	Tools map[string]Tool
}

type Step struct {
	Tool   string
	Input  string
	Reason string
}

func main() {
	mode := flag.String("mode", "react", "agent mode: react or plan")
	question := flag.String("q", "", "question/task to solve")
	verbose := flag.Bool("v", true, "show reasoning logs")
	flag.Parse()

	if strings.TrimSpace(*question) == "" {
		fmt.Println("Usage:")
		fmt.Println(`  go run react_plan_agent_single.go -mode react -q "what time is it and count words in hello world"`)
		fmt.Println(`  go run react_plan_agent_single.go -mode plan  -q "uppercase hello and then count words in hello world"`)
		os.Exit(1)
	}

	agent := NewAgent()
	var answer string
	var err error

	switch strings.ToLower(strings.TrimSpace(*mode)) {
	case "react":
		answer, err = agent.RunReAct(*question, *verbose)
	case "plan", "plan-and-execute", "pae":
		answer, err = agent.RunPlanAndExecute(*question, *verbose)
	default:
		err = fmt.Errorf("unknown mode: %s", *mode)
	}

	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

	fmt.Println("\n=== FINAL ANSWER ===")
	fmt.Println(answer)
}

func NewAgent() *Agent {
	return &Agent{Tools: map[string]Tool{
		"time":       toolTime,
		"word_count": toolWordCount,
		"uppercase":  toolUppercase,
		"lowercase":  toolLowercase,
		"reverse":    toolReverse,
		"calc":       toolCalc,
	}}
}

func (a *Agent) RunReAct(question string, verbose bool) (string, error) {
	maxSteps := 4
	observations := []ToolResult{}
	remaining := strings.TrimSpace(question)

	for i := 1; i <= maxSteps; i++ {
		step, final, done := decideNextReActStep(remaining, observations)
		if verbose {
			fmt.Printf("[ReAct] step=%d\n", i)
			if done {
				fmt.Printf("  Thought: I already have enough information.\n")
				fmt.Printf("  Final: %s\n", final)
			} else {
				fmt.Printf("  Thought: %s\n", step.Reason)
				fmt.Printf("  Action: %s(%q)\n", step.Tool, step.Input)
			}
		}
		if done {
			return final, nil
		}

		tool, ok := a.Tools[step.Tool]
		if !ok {
			return "", fmt.Errorf("tool not found: %s", step.Tool)
		}
		result := tool(step.Input)
		observations = append(observations, result)
		if verbose {
			fmt.Printf("  Observation: %s\n", result.Output)
		}

		remaining = updateRemainingGoal(question, observations)
	}

	if len(observations) == 0 {
		return fallbackDirectAnswer(question)
	}
	return synthesizeAnswer(question, observations), nil
}

func (a *Agent) RunPlanAndExecute(question string, verbose bool) (string, error) {
	plan := makePlan(question)
	if len(plan) == 0 {
		return fallbackDirectAnswer(question)
	}
	if verbose {
		fmt.Println("[Plan] Generated steps:")
		for i, s := range plan {
			fmt.Printf("  %d. %s(%q) -- %s\n", i+1, s.Tool, s.Input, s.Reason)
		}
	}

	results := []ToolResult{}
	for i, step := range plan {
		tool, ok := a.Tools[step.Tool]
		if !ok {
			return "", fmt.Errorf("tool not found: %s", step.Tool)
		}
		result := tool(step.Input)
		results = append(results, result)
		if verbose {
			fmt.Printf("[Execute] step=%d result=%s\n", i+1, result.Output)
		}
	}
	return synthesizeAnswer(question, results), nil
}

func decideNextReActStep(question string, observations []ToolResult) (Step, string, bool) {
	needed := makePlan(question)
	completed := map[string]int{}
	for _, obs := range observations {
		completed[obs.Name]++
	}
	for _, step := range needed {
		if completed[step.Tool] == 0 {
			return step, "", false
		}
		completed[step.Tool]--
	}
	return Step{}, synthesizeAnswer(question, observations), true
}

func makePlan(question string) []Step {
	q := strings.TrimSpace(question)
	lower := strings.ToLower(q)
	steps := []Step{}

	if strings.Contains(lower, "time") || strings.Contains(lower, "几点") || strings.Contains(lower, "时间") || strings.Contains(lower, "date") {
		steps = append(steps, Step{Tool: "time", Input: q, Reason: "Need current time/date information."})
	}

	if expr := extractExpression(q); expr != "" {
		steps = append(steps, Step{Tool: "calc", Input: expr, Reason: "Need arithmetic calculation."})
	}

	if txt := extractQuotedText(q, []string{"uppercase", "upper", "大写"}); txt != "" {
		steps = append(steps, Step{Tool: "uppercase", Input: txt, Reason: "Convert text to uppercase."})
	}
	if txt := extractQuotedText(q, []string{"lowercase", "lower", "小写"}); txt != "" {
		steps = append(steps, Step{Tool: "lowercase", Input: txt, Reason: "Convert text to lowercase."})
	}
	if txt := extractQuotedText(q, []string{"reverse", "倒序", "反转"}); txt != "" {
		steps = append(steps, Step{Tool: "reverse", Input: txt, Reason: "Reverse the text."})
	}
	if txt := extractQuotedText(q, []string{"count words in", "word count", "统计单词", "数单词"}); txt != "" {
		steps = append(steps, Step{Tool: "word_count", Input: txt, Reason: "Count words in the target text."})
	}

	if len(steps) == 0 {
		// Basic fallback: if no tool clearly matches, answer directly.
		return nil
	}
	return steps
}

func updateRemainingGoal(question string, observations []ToolResult) string {
	// For this simple demo, we keep the original goal and rely on completed tool tracking.
	return question
}

func synthesizeAnswer(question string, observations []ToolResult) string {
	if len(observations) == 0 {
		ans, _ := fallbackDirectAnswer(question)
		return ans
	}
	parts := make([]string, 0, len(observations))
	for _, obs := range observations {
		label := obs.Name
		switch obs.Name {
		case "time":
			label = "当前时间"
		case "calc":
			label = "计算结果"
		case "uppercase":
			label = "大写结果"
		case "lowercase":
			label = "小写结果"
		case "reverse":
			label = "倒序结果"
		case "word_count":
			label = "单词统计"
		}
		parts = append(parts, fmt.Sprintf("%s: %s", label, obs.Output))
	}
	return strings.Join(parts, " | ")
}

func fallbackDirectAnswer(question string) (string, error) {
	q := strings.TrimSpace(question)
	if q == "" {
		return "", errors.New("empty question")
	}
	return "这是一个基础版 agent，目前更擅长处理时间、算术、大小写、倒序、单词统计这类简单任务。你的问题超出内置工具范围：" + q, nil
}

func toolTime(_ string) ToolResult {
	now := time.Now()
	return ToolResult{Name: "time", Success: true, Output: now.Format("2006-01-02 15:04:05 MST")}
}

func toolWordCount(input string) ToolResult {
	fields := strings.Fields(input)
	return ToolResult{Name: "word_count", Input: input, Success: true, Output: fmt.Sprintf("%d words", len(fields))}
}

func toolUppercase(input string) ToolResult {
	return ToolResult{Name: "uppercase", Input: input, Success: true, Output: strings.ToUpper(input)}
}

func toolLowercase(input string) ToolResult {
	return ToolResult{Name: "lowercase", Input: input, Success: true, Output: strings.ToLower(input)}
}

func toolReverse(input string) ToolResult {
	r := []rune(input)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return ToolResult{Name: "reverse", Input: input, Success: true, Output: string(r)}
}

func toolCalc(input string) ToolResult {
	val, err := evalExpr(input)
	if err != nil {
		return ToolResult{Name: "calc", Input: input, Success: false, Output: "calc error: " + err.Error()}
	}
	return ToolResult{Name: "calc", Input: input, Success: true, Output: trimFloat(val)}
}

func extractQuotedText(question string, hints []string) string {
	lower := strings.ToLower(question)
	matched := false
	for _, h := range hints {
		if strings.Contains(lower, strings.ToLower(h)) {
			matched = true
			break
		}
	}
	if !matched {
		return ""
	}
	re := regexp.MustCompile(`"([^"]+)"|'([^']+)'`)
	m := re.FindStringSubmatch(question)
	if len(m) >= 3 {
		if m[1] != "" {
			return m[1]
		}
		return m[2]
	}

	// Chinese-style fallback: take text after the keyword.
	for _, h := range hints {
		idx := strings.Index(lower, strings.ToLower(h))
		if idx >= 0 {
			tail := strings.TrimSpace(question[idx+len(h):])
			tail = strings.Trim(tail, " ：:，,.。")
			if tail != "" {
				return tail
			}
		}
	}
	return ""
}

func extractExpression(question string) string {
	// Prefer text after explicit calculate keywords.
	lower := strings.ToLower(question)
	for _, key := range []string{"calculate", "calc", "compute", "算", "计算"} {
		if idx := strings.Index(lower, key); idx >= 0 {
			tail := strings.TrimSpace(question[idx+len(key):])
			tail = strings.Trim(tail, " ：:，,.。")
			if looksLikeMathExpr(tail) {
				return tail
			}
		}
	}
	// Fallback: find longest math-looking chunk.
	re := regexp.MustCompile(`[0-9\s\+\-\*/\(\)\.]+`)
	candidates := re.FindAllString(question, -1)
	sort.Slice(candidates, func(i, j int) bool { return len(candidates[i]) > len(candidates[j]) })
	for _, c := range candidates {
		c = strings.TrimSpace(c)
		if looksLikeMathExpr(c) {
			return c
		}
	}
	return ""
}

func looksLikeMathExpr(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	hasDigit := false
	for _, r := range s {
		if unicode.IsDigit(r) {
			hasDigit = true
			continue
		}
		if strings.ContainsRune("+-*/(). ", r) {
			continue
		}
		return false
	}
	return hasDigit
}

// --- Tiny arithmetic parser: supports + - * / and parentheses ---

type parser struct {
	input string
	pos   int
}

func evalExpr(s string) (float64, error) {
	p := &parser{input: strings.ReplaceAll(s, " ", "")}
	v, err := p.parseExpr()
	if err != nil {
		return 0, err
	}
	if p.pos != len(p.input) {
		return 0, fmt.Errorf("unexpected token at position %d", p.pos)
	}
	return v, nil
}

func (p *parser) parseExpr() (float64, error) {
	left, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for p.pos < len(p.input) {
		op := p.input[p.pos]
		if op != '+' && op != '-' {
			break
		}
		p.pos++
		right, err := p.parseTerm()
		if err != nil {
			return 0, err
		}
		if op == '+' {
			left += right
		} else {
			left -= right
		}
	}
	return left, nil
}

func (p *parser) parseTerm() (float64, error) {
	left, err := p.parseFactor()
	if err != nil {
		return 0, err
	}
	for p.pos < len(p.input) {
		op := p.input[p.pos]
		if op != '*' && op != '/' {
			break
		}
		p.pos++
		right, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		if op == '*' {
			left *= right
		} else {
			if right == 0 {
				return 0, errors.New("division by zero")
			}
			left /= right
		}
	}
	return left, nil
}

func (p *parser) parseFactor() (float64, error) {
	if p.pos >= len(p.input) {
		return 0, errors.New("unexpected end of expression")
	}
	if p.input[p.pos] == '(' {
		p.pos++
		v, err := p.parseExpr()
		if err != nil {
			return 0, err
		}
		if p.pos >= len(p.input) || p.input[p.pos] != ')' {
			return 0, errors.New("missing closing parenthesis")
		}
		p.pos++
		return v, nil
	}
	if p.input[p.pos] == '-' {
		p.pos++
		v, err := p.parseFactor()
		return -v, err
	}
	start := p.pos
	for p.pos < len(p.input) && (unicode.IsDigit(rune(p.input[p.pos])) || p.input[p.pos] == '.') {
		p.pos++
	}
	if start == p.pos {
		return 0, fmt.Errorf("expected number at position %d", p.pos)
	}
	return strconv.ParseFloat(p.input[start:p.pos], 64)
}

func trimFloat(v float64) string {
	if math.Mod(v, 1) == 0 {
		return fmt.Sprintf("%.0f", v)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}
