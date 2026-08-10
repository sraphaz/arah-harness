// Package core is arah-core: typed domain model and Execution Control invariants.
// No I/O, no PowerShell, no GitHub — only domain rules (kern hexagonal style).
package core

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type State string

const (
	StateIntake    State = "intake"
	StateRouted    State = "routed"
	StateExecuting State = "executing"
	StateVerifying State = "verifying"
	StateDone      State = "done"
	StateBlocked   State = "blocked"
)

type WorkClass string

const (
	WorkTrivial        WorkClass = "trivial"
	WorkStandard       WorkClass = "standard"
	WorkArchitectural  WorkClass = "architectural"
	WorkRelease        WorkClass = "release"
)

type IntentType string

const (
	IntentAnalysis   IntentType = "analysis"
	IntentExecution  IntentType = "execution"
	IntentReview     IntentType = "review"
	IntentPlanning   IntentType = "planning"
)

// DomainError carries a stable machine code (envelope-friendly).
type DomainError struct {
	Code        string
	Message     string
	Details     map[string]any
	Remediation []string
}

func (e *DomainError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

func errf(code, msg string, details map[string]any, rem ...string) *DomainError {
	return &DomainError{Code: code, Message: msg, Details: details, Remediation: rem}
}

var validTransitions = map[State][]State{
	StateIntake:    {StateRouted, StateBlocked},
	StateRouted:    {StateExecuting, StateBlocked},
	StateExecuting: {StateVerifying, StateBlocked, StateDone},
	StateVerifying: {StateDone, StateBlocked, StateExecuting},
	StateDone:      {},
	StateBlocked:   {},
}

func TransitionAllowed(from, to State) bool {
	for _, s := range validTransitions[from] {
		if s == to {
			return true
		}
	}
	return false
}

func IsTerminal(s State) bool {
	return s == StateDone || s == StateBlocked
}

// Contract is the Execution Control work instance.
type Contract struct {
	Version          string            `yaml:"version" json:"version"`
	TaskID           string            `yaml:"task_id" json:"task_id"`
	Objective        string            `yaml:"objective" json:"objective"`
	WorkClass        WorkClass         `yaml:"work_class" json:"work_class"`
	IntentType       IntentType        `yaml:"intent_type" json:"intent_type"`
	State            State             `yaml:"state" json:"state"`
	PrimaryExecutor  string            `yaml:"primary_executor" json:"primary_executor"`
	ChoreographyRule string            `yaml:"choreography_rule,omitempty" json:"choreography_rule,omitempty"`
	Participants     Participants      `yaml:"participants" json:"participants"`
	Scope            Scope             `yaml:"scope" json:"scope"`
	Execution        Execution         `yaml:"execution" json:"execution"`
	Limits           Limits            `yaml:"limits" json:"limits"`
	Counters         Counters          `yaml:"counters" json:"counters"`
	Policy           map[string]any    `yaml:"policy" json:"policy"`
	Result           Result            `yaml:"result" json:"result"`
	History          []HistoryEntry    `yaml:"history" json:"history"`
}

type Participants struct {
	Consultants  []string `yaml:"consultants" json:"consultants"`
	Reviewers    []string `yaml:"reviewers" json:"reviewers"`
	Subordinates []string `yaml:"subordinates" json:"subordinates"`
}

type Scope struct {
	Area           string   `yaml:"area" json:"area"`
	AllowedPaths   []string `yaml:"allowed_paths" json:"allowed_paths"`
	ForbiddenPaths []string `yaml:"forbidden_paths" json:"forbidden_paths"`
}

type Execution struct {
	ExpectedOutputs      []string `yaml:"expected_outputs" json:"expected_outputs"`
	VerificationCommands []string `yaml:"verification_commands" json:"verification_commands"`
	CompletionEvidence   []string `yaml:"completion_evidence" json:"completion_evidence"`
}

type Limits struct {
	MaxHandoffs         int `yaml:"max_handoffs" json:"max_handoffs"`
	MaxConsultations    int `yaml:"max_consultations" json:"max_consultations"`
	MaxAnalysisCycles   int `yaml:"max_analysis_cycles" json:"max_analysis_cycles"`
}

type Counters struct {
	Handoffs        int `yaml:"handoffs" json:"handoffs"`
	Consultations   int `yaml:"consultations" json:"consultations"`
	AnalysisCycles  int `yaml:"analysis_cycles" json:"analysis_cycles"`
}

type Result struct {
	ChangedFiles     []string `yaml:"changed_files" json:"changed_files"`
	CommandsExecuted []string `yaml:"commands_executed" json:"commands_executed"`
	Evidence         []string `yaml:"evidence" json:"evidence"`
	BlockingReason   *string  `yaml:"blocking_reason" json:"blocking_reason"`
}

type HistoryEntry struct {
	At   string `yaml:"at" json:"at"`
	From string `yaml:"from" json:"from"`
	To   string `yaml:"to" json:"to"`
	Note string `yaml:"note,omitempty" json:"note,omitempty"`
}

// ResolvedRouting is the choreography outcome used to build a contract.
type ResolvedRouting struct {
	PrimaryExecutor  string
	ChoreographyRule string
	Consultants      []string
	Reviewers        []string
	Subordinates     []string
	AllowedPaths     []string
}

// WorkClassPolicy returns default limits/policy flags for a class.
func WorkClassPolicy(wc WorkClass) (Limits, map[string]any) {
	switch wc {
	case WorkTrivial:
		return Limits{0, 0, 0}, map[string]any{"spec_required": false}
	case WorkArchitectural:
		return Limits{2, 2, 1}, map[string]any{"full_spec_required": true}
	case WorkRelease:
		return Limits{1, 2, 1}, map[string]any{
			"release_approval_required": true,
			"human_gate_required":       true,
		}
	default:
		return Limits{1, 1, 1}, map[string]any{"lightweight_spec": true}
	}
}

func NewTaskID() string {
	stamp := time.Now().UTC().Format("20060102-150405.000")
	stamp = strings.ReplaceAll(stamp, ".", "")
	return fmt.Sprintf("task-%s-%s", stamp, shortRand())
}

func shortRand() string {
	return fmt.Sprintf("%06x", time.Now().UnixNano()&0xffffff)
}

func NewContract(objective, area string, wc WorkClass, intent IntentType, routing ResolvedRouting) (*Contract, error) {
	if strings.TrimSpace(objective) == "" {
		return nil, errf("EXECUTION.OBJECTIVE_REQUIRED", "objective is required", nil, "Pass a non-empty objective")
	}
	if routing.PrimaryExecutor == "" {
		return nil, errf("EXECUTION.EXACTLY_ONE_PRIMARY_EXECUTOR_REQUIRED", "no eligible primary_executor", map[string]any{"area": area})
	}
	if routing.PrimaryExecutor == "orchestrator" {
		return nil, errf("EXECUTION.ORCHESTRATOR_CANNOT_BE_PRIMARY", "orchestrator cannot be primary_executor", nil)
	}
	limits, policy := WorkClassPolicy(wc)
	consultants := append([]string{}, routing.Consultants...)
	if limits.MaxConsultations == 0 {
		consultants = nil
	}
	c := &Contract{
		Version:          "1.0",
		TaskID:           NewTaskID(),
		Objective:        objective,
		WorkClass:        wc,
		IntentType:       intent,
		State:            StateIntake,
		PrimaryExecutor:  routing.PrimaryExecutor,
		ChoreographyRule: routing.ChoreographyRule,
		Participants: Participants{
			Consultants:  consultants,
			Reviewers:    append([]string{}, routing.Reviewers...),
			Subordinates: append([]string{}, routing.Subordinates...),
		},
		Scope: Scope{
			Area:         area,
			AllowedPaths: append([]string{}, routing.AllowedPaths...),
		},
		Execution: Execution{},
		Limits:    limits,
		Counters:  Counters{},
		Policy:    policy,
		Result:    Result{},
		History:   nil,
	}
	c.addHistory("none", string(StateIntake), "contract created")
	return c, nil
}

func (c *Contract) addHistory(from, to, note string) {
	c.History = append(c.History, HistoryEntry{
		At:   time.Now().UTC().Format(time.RFC3339Nano),
		From: from,
		To:   to,
		Note: note,
	})
}

// Transition applies a validated state change.
func (c *Contract) Transition(to State, note string) error {
	from := c.State
	if from == to {
		return nil
	}
	if IsTerminal(from) {
		return errf("EXECUTION.TERMINAL_STATE_IMMUTABLE",
			fmt.Sprintf("task is terminal (%s)", from),
			map[string]any{"task_id": c.TaskID, "state": string(from)})
	}
	if from == StateExecuting && to == StateRouted {
		return errf("EXECUTION.REROUTE_AFTER_EXECUTING_FORBIDDEN",
			"after executing, task cannot return to routed",
			map[string]any{"task_id": c.TaskID, "from": string(from), "to": string(to)})
	}
	if !TransitionAllowed(from, to) {
		return errf("EXECUTION.INVALID_STATE_TRANSITION",
			fmt.Sprintf("invalid_state_transition:%s->%s", from, to),
			map[string]any{"task_id": c.TaskID, "from": string(from), "to": string(to)})
	}
	if (to == StateRouted || to == StateExecuting || to == StateVerifying || to == StateDone || to == StateBlocked) &&
		strings.TrimSpace(c.PrimaryExecutor) == "" {
		return errf("EXECUTION.PRIMARY_EXECUTOR_REQUIRED",
			"primary_executor required for this state",
			map[string]any{"task_id": c.TaskID, "state": string(to)})
	}
	c.addHistory(string(from), string(to), note)
	c.State = to
	return nil
}

var concreteEvidenceRe = regexp.MustCompile(`(?i)(updated|created|removed|deleted|wrote|patched|test(s)?\s+(pass|ok|green)|executed|file:|path:|\.(ts|tsx|js|go|ps1|yaml|yml|md)\b)`)
var analysisOnlyRe = regexp.MustCompile(`(?i)^(analy[sz]e|análise|parecer|review only)`)

// HasConcreteEvidence mirrors Test-EcpConcreteEvidence.
func (c *Contract) HasConcreteEvidence(extra []string) bool {
	all := append([]string{}, extra...)
	all = append(all, c.Execution.CompletionEvidence...)
	all = append(all, c.Result.Evidence...)
	all = append(all, c.Result.ChangedFiles...)
	all = append(all, c.Result.CommandsExecuted...)
	if c.IntentType != IntentExecution {
		return len(filterNonEmpty(all)) > 0
	}
	concrete := false
	for _, e := range all {
		s := strings.TrimSpace(e)
		if s == "" {
			continue
		}
		if analysisOnlyRe.MatchString(s) {
			continue
		}
		if concreteEvidenceRe.MatchString(s) || strings.ContainsAny(s, "/\\.") {
			concrete = true
			break
		}
	}
	return concrete
}

func filterNonEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}

// Start moves intake → routed → executing (orchestration start).
func (c *Contract) Start() error {
	if c.State == StateIntake {
		if err := c.Transition(StateRouted, "choreography resolved"); err != nil {
			return err
		}
	}
	if c.State == StateRouted {
		return c.Transition(StateExecuting, "primary executor activated")
	}
	return nil
}

// Complete validates evidence and transitions to done.
func (c *Contract) Complete(evidence []string) error {
	if IsTerminal(c.State) {
		return errf("EXECUTION.TERMINAL_STATE_IMMUTABLE",
			fmt.Sprintf("task is terminal (%s)", c.State),
			map[string]any{"task_id": c.TaskID})
	}
	if len(filterNonEmpty(evidence)) == 0 {
		return errf("EXECUTION.COMPLETION_EVIDENCE_REQUIRED",
			"A tarefa não pode ser concluída sem evidência.",
			map[string]any{"task_id": c.TaskID},
			"Informe arquivos alterados",
			"Registre os testes executados")
	}
	if !c.HasConcreteEvidence(evidence) {
		return errf("EXECUTION.COMPLETION_EVIDENCE_REQUIRED",
			"evidence is not concrete enough for intent_type=execution",
			map[string]any{"task_id": c.TaskID},
			"Include file paths or test results")
	}
	if c.State == StateExecuting {
		if err := c.Transition(StateVerifying, "verification started"); err != nil {
			return err
		}
	}
	if err := c.Transition(StateDone, "completion evidence accepted"); err != nil {
		return err
	}
	c.Execution.CompletionEvidence = append(c.Execution.CompletionEvidence, evidence...)
	c.Result.Evidence = append(c.Result.Evidence, evidence...)
	return nil
}

// Block records a concrete blocking reason.
func (c *Contract) Block(reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return errf("EXECUTION.BLOCKING_REASON_REQUIRED",
			"blocked exige causa concreta",
			map[string]any{"task_id": c.TaskID},
			"Informe -reason com a dependência ou decisão faltante")
	}
	if err := c.Transition(StateBlocked, reason); err != nil {
		return err
	}
	c.Result.BlockingReason = &reason
	return nil
}
