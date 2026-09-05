// Command generator exports the OpenAPI Changes Model from the oasdiff
// reference implementation: the vocabulary, the severity law, every named
// change with its claims, and the coverage of the full edit space.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"slices"
	"sort"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/checker/coverage"
	"github.com/oasdiff/oasdiff/checker/localizations"
	"github.com/oasdiff/oasdiff/checker/rules"
	"gopkg.in/yaml.v3"
)

type Model struct {
	Model         string          `yaml:"model"`
	Version       string          `yaml:"version"`
	GeneratedFrom string          `yaml:"generated_from"`
	Vocabulary    Vocabulary      `yaml:"vocabulary"`
	SeverityLaw   SeverityLaw     `yaml:"severity_law"`
	Transitions   []Transition    `yaml:"transitions"`
	Changes       []Change        `yaml:"changes"`
	Coverage      []coverage.Edit `yaml:"coverage"`
}

type Vocabulary struct {
	Locations  string            `yaml:"locations"`
	Actions    map[string]string `yaml:"actions"`
	Directions map[string]string `yaml:"directions"`
	Areas      map[string]string `yaml:"areas"`
	Kinds      map[string]string `yaml:"kinds"`
	Effects    map[string]string `yaml:"effects"`
	Guards     map[string]string `yaml:"guards"`
	Levels     map[string]string `yaml:"levels"`
	Statuses   map[string]string `yaml:"statuses"`
	Categories map[string]string `yaml:"categories"`
}

type SeverityLaw struct {
	Description string        `yaml:"description"`
	Guards      []GuardRule   `yaml:"guards_apply_first"`
	Verdicts    []VerdictRule `yaml:"verdicts"`
}

type GuardRule struct {
	Guard  string `yaml:"guard"`
	Effect string `yaml:"then"`
}

type VerdictRule struct {
	Effect    string `yaml:"effect"`
	Direction string `yaml:"direction,omitempty"`
	Level     string `yaml:"level"`
}

type Transition struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	ClaimedKinds []string `yaml:"claimed_kinds"`
	ReportedBy   []string `yaml:"reported_by"`
}

type Change struct {
	Id          string   `yaml:"id"`
	Level       string   `yaml:"level"`
	Direction   string   `yaml:"direction"`
	Area        string   `yaml:"area"`
	Kind        string   `yaml:"kind"`
	Effect      string   `yaml:"effect"`
	Guards      []string `yaml:"guards,omitempty"`
	Claims      []string `yaml:"claims"`
	Description string   `yaml:"description"`
	Message     string   `yaml:"message"`
}

func oasdiffVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/oasdiff/oasdiff" {
				return dep.Version
			}
		}
	}
	return "unknown"
}

func main() {
	out := flag.String("out", "../openapi-changes-model.yaml", "output path")
	flag.Parse()

	localizer := localizations.New(localizations.LangEn, "")
	metadata := checker.GetAllRules().Metadata()

	changes := make([]Change, 0, len(metadata))
	for _, r := range metadata {
		guards := make([]string, 0, len(r.Guards))
		for _, g := range r.Guards {
			guards = append(guards, string(g))
		}
		changes = append(changes, Change{
			Id:          r.Id,
			Level:       r.Level.String(),
			Direction:   r.Direction.String(),
			Area:        r.Area.String(),
			Kind:        r.Kind.String(),
			Effect:      r.Effect.String(),
			Guards:      guards,
			Claims:      r.Locations,
			Description: localizer.Get("messages." + r.Id + "-description"),
			Message:     localizer.Get("messages." + r.Id),
		})
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Id < changes[j].Id })

	transitions := make([]Transition, 0)
	for _, tr := range checker.GetTransitions() {
		kinds := make([]string, 0, len(tr.Claims))
		for _, k := range tr.Claims {
			kinds = append(kinds, k.String())
		}
		reportedBy := slices.Clone(tr.ReportedBy)
		sort.Strings(reportedBy)
		transitions = append(transitions, Transition{
			Name:         tr.Name,
			Description:  tr.Description,
			ClaimedKinds: kinds,
			ReportedBy:   reportedBy,
		})
	}
	sort.Slice(transitions, func(i, j int) bool { return transitions[i].Name < transitions[j].Name })

	model := Model{
		Model:         "OpenAPI Changes Model",
		Version:       "0.1.0-draft",
		GeneratedFrom: "oasdiff " + oasdiffVersion(),
		Vocabulary: Vocabulary{
			Locations: "A location is a path through the OpenAPI object model, dot-separated, with * standing " +
				"in for a name the API author chooses (a path, a method, a media type, a property name) and x-* for a specification " +
				"extension: paths.*.*.requestBody.content.*.schema.maxLength names the maxLength keyword of any " +
				"request body schema. A claim is location:action[,action...], the edits a change covers; a claim " +
				"pattern may use ** to cover a location family.",
			Actions: map[string]string{
				"add":      "a member is added to a collection (a property, an enum value, a response status)",
				"remove":   "a member is removed from a collection",
				"set":      "a field appears where it was absent",
				"unset":    "a field disappears",
				"change":   "a field's value is replaced by an incomparable value",
				"increase": "an ordered field's value grows",
				"decrease": "an ordered field's value shrinks",
			},
			Directions: map[string]string{
				"request":  "the change concerns what clients send",
				"response": "the change concerns what clients receive",
				"none":     "the change concerns neither side of the wire (metadata, lifecycle)",
			},
			Areas: map[string]string{
				"schema":      "a schema and its keywords, wherever the schema appears",
				"parameters":  "operation and path parameters",
				"requestBody": "the request body object and its media types",
				"responses":   "the responses map, response objects, and their media types",
				"paths":       "paths, operations, and operation metadata",
				"headers":     "response headers",
				"security":    "security schemes, requirements, and scopes",
				"tags":        "tags and their metadata",
				"components":  "the components section (compared where referenced)",
			},
			Kinds: map[string]string{
				"existence":    "an element is added or removed",
				"requiredness": "required, optional, or nullable state",
				"mutability":   "read-only or write-only state",
				"type":         "data type or format",
				"constraints":  "bounds such as min/max, length, items, pattern",
				"values":       "enum, const, and default values",
				"structure":    "composition and applicator keywords: allOf, anyOf, oneOf, discriminator, if/then/else, contains",
				"lifecycle":    "deprecation, sunset, and stability",
			},
			Effects: map[string]string{
				"narrows":      "the new contract rejects payloads the previous contract accepted",
				"widens":       "the new contract accepts payloads the previous contract rejected",
				"incomparable": "the change both rejects previously valid payloads and accepts previously invalid ones",
				"unknown":      "the specification does not carry enough information to decide the effect",
				"none":         "the change cannot affect which payloads are valid",
				"violation":    "the change breaks a declared lifecycle contract (deprecation, sunset, stability) rather than the wire contract",
			},
			Guards: map[string]string{
				"read-only":   "the changed property is readOnly, so it never appears in requests; request-side effects are nullified",
				"write-only":  "the changed property is writeOnly, so it never appears in responses; response-side effects are nullified",
				"sanctioned":  "the removed element was deprecated and its sunset period was honored, so the removal follows the deprecation contract",
				"non-success": "the affected response status is a non-success status; the responses map does not promise the server returns only the statuses it lists",
				"has-default": "the changed element declares a default value (declared for audit; does not change the verdict)",
				"negotiated":  "the element is one the client selects or relies on (a status, media type, or header); its availability is judged with request polarity",
			},
			Levels: map[string]string{
				"error":   "a consumer that conformed to the old contract can stop conforming or fail",
				"warning": "plausibly breaking, but the specification cannot decide; the finding says what is missing",
				"info":    "provably safe for every consumer that conformed to the old contract",
			},
			Statuses: map[string]string{
				"covered":      "one or more named changes claim the edit",
				"waived":       "no change covers the edit and a written reason says why; the category refines it",
				"non-contract": "the edit cannot affect which payloads are valid (descriptions, examples, extensions)",
				"uncovered":    "no change and no waiver; the reference implementation fails its build in this state, so the listing normally contains none",
			},
			Categories: map[string]string{
				"open":              "a missing change, with its reason and a suggested id",
				"resolved-at-usage": "component definitions are compared at their referencing operations, which have their own rows",
				"covered-as":        "the same document edit is reported under another action",
			},
		},
		SeverityLaw: SeverityLaw{
			Description: "A change's level is derived from its effect, its direction, and its guards. " +
				"Guards apply first, each nullifying or requalifying the effect on the side it speaks about; " +
				"then the effect and direction decide: narrowing breaks request consumers, widening breaks " +
				"response consumers, an incomparable change breaks both, and an unknown one is a warning. " +
				"When a change cannot be proven safe it is reported as breaking.",
			Guards: []GuardRule{
				{"read-only", "effect becomes none when direction is request"},
				{"write-only", "effect becomes none when direction is response"},
				{"non-success", "effect becomes none"},
				{"sanctioned", "effect becomes none"},
				{"negotiated", "direction becomes request"},
			},
			Verdicts: []VerdictRule{
				{Effect: "narrows", Direction: "request", Level: "error"},
				{Effect: "narrows", Direction: "response", Level: "info"},
				{Effect: "widens", Direction: "request", Level: "info"},
				{Effect: "widens", Direction: "response", Level: "error"},
				{Effect: "incomparable", Level: "error"},
				{Effect: "violation", Level: "error"},
				{Effect: "unknown", Level: "warning"},
				{Effect: "none", Level: "info"},
			},
		},
		Transitions: transitions,
		Changes:     changes,
		Coverage:    coverage.Analyze(metadata),
	}

	data, err := yaml.Marshal(model)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	header := "# The OpenAPI Changes Model. Generated from the oasdiff reference implementation;\n" +
		"# do not edit by hand. See README.md for how to propose a change.\n"
	if err := os.WriteFile(*out, append([]byte(header), data...), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s: %d changes, %d edits\n", *out, len(changes), len(model.Coverage))
}

var _ = rules.DeriveLevel // the law encoded above is the one this implementation runs
