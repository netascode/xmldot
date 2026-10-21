// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// withLimit sets a security limit for the duration of a test or function and
// returns a function that restores the previous value.
// Usage: defer withLimit(&MaxDocumentSize, 1024)()
func withLimit(limit *int, value int) func() {
	old := *limit
	*limit = value
	return func() { *limit = old }
}

// recoverLimit converts a limit violation raised by exceedLimit into an error, for
// tests that call internal functions directly.
// Usage: defer recoverLimit(&err). Other panics are re-raised unchanged.
func recoverLimit(err *error) {
	if r := recover(); r != nil {
		lp, ok := r.(limitPanic)
		if !ok {
			panic(r)
		}
		*err = lp.err
	}
}

// requireLimitError fails the test unless err is a *LimitError for the named limit.
func requireLimitError(t *testing.T, err error, limit string) {
	t.Helper()
	if !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("expected ErrLimitExceeded for %s, got %v", limit, err)
	}
	var limitErr *LimitError
	if !errors.As(err, &limitErr) {
		t.Fatalf("expected *LimitError for %s, got %T", limit, err)
	}
	if limitErr.Limit != limit {
		t.Fatalf("expected limit %s, got %s", limit, limitErr.Limit)
	}
}

func TestLimitError(t *testing.T) {
	err := error(&LimitError{Limit: "MaxDocumentSize", Value: 11, Max: 10})
	if !errors.Is(err, ErrLimitExceeded) {
		t.Error("LimitError should wrap ErrLimitExceeded")
	}
	want := "security limit exceeded: MaxDocumentSize is 10, observed 11"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestLimits_GetFailsInsteadOfTruncating(t *testing.T) {
	items := func(n int) string {
		var sb strings.Builder
		sb.WriteString("<root>")
		for i := 0; i < n; i++ {
			fmt.Fprintf(&sb, "<item><name>n%d</name></item>", i)
		}
		sb.WriteString("</root>")
		return sb.String()
	}

	tests := []struct {
		name  string
		limit *int
		value int
		xml   string
		path  string
		want  string
	}{
		{"document size", &MaxDocumentSize, 100, items(10), "root.item", "MaxDocumentSize"},
		{"array count", &MaxArrayResults, 5, items(10), "root.item.#", "MaxArrayResults"},
		{"array index", &MaxArrayResults, 5, items(10), "root.item.7", "MaxArrayResults"},
		{"fragment roots", &MaxArrayResults, 5, strings.Repeat("<a>1</a>", 10), "a.#", "MaxArrayResults"},
		{"field extraction children", &MaxArrayResults, 5, "<root><item><name>a</name><name>b</name><name>c</name></item><item><name>d</name><name>e</name><name>f</name></item></root>", "root.item.#.name", "MaxArrayResults"},
		{"field extraction", &MaxArrayResults, 5, items(10), "root.item.#.name", "MaxArrayResults"},
		{"filter", &MaxArrayResults, 5, items(10), `root.item.#(name%"n*")#`, "MaxArrayResults"},
		{"single-level wildcard", &MaxArrayResults, 5, items(10), "root.*", "MaxArrayResults"},
		{"recursive wildcard results", &MaxWildcardResults, 5, items(10), "root.**.name", "MaxWildcardResults"},
		{"attributes", &MaxAttributes, 2, `<root><a x="1" y="2" z="3">v</a></root>`, "root.a", "MaxAttributes"},
		{"element name token size", &MaxTokenSize, 10, `<root><` + strings.Repeat("e", 20) + `>v</` + strings.Repeat("e", 20) + `><a>1</a></root>`, "root.a", "MaxTokenSize"},
		{"token size", &MaxTokenSize, 10, `<root><a x="` + strings.Repeat("v", 20) + `">v</a></root>`, "root.a", "MaxTokenSize"},
		{"path segments", &MaxPathSegments, 2, items(1), "root.item.name", "MaxPathSegments"},
		{"filter expression", &MaxFilterExpressionLength, 5, items(1), `root.item.#(name=="n0")`, "MaxFilterExpressionLength"},
		{"pattern iterations", &MaxPatternIterations, 10, "<root><item><v>" + strings.Repeat("a", 200) + "</v></item></root>", `root.item.#(v!%"*` + strings.Repeat("a", 50) + `b")`, "MaxPatternIterations"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The query succeeds with the default limits
			if r := Get(tt.xml, tt.path); r.Err != nil || !r.Exists() {
				t.Fatalf("query should succeed with default limits: exists=%v err=%v", r.Exists(), r.Err)
			}

			defer withLimit(tt.limit, tt.value)()
			r := Get(tt.xml, tt.path)
			if r.Exists() {
				t.Errorf("expected Null result, got %v", r.Type)
			}
			requireLimitError(t, r.Err, tt.want)
		})
	}
}

func TestLimits_EarlyStop(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("<root>")
	for i := 0; i < 10; i++ {
		fmt.Fprintf(&sb, "<item><k>%d</k></item>", i)
	}
	sb.WriteString("</root>")
	xml := sb.String()
	fragment := strings.Repeat("<a>1</a>", 10)

	defer withLimit(&MaxArrayResults, 5)()

	// Indexed access and first-match filters only need the siblings up to the answer
	for path, want := range map[string]string{
		"root.item.0.k":          "0",
		"root.item.4.k":          "4",
		"root.item.#(k==3).k":    "3",
		"root.item.#(k>0).k":     "1",
		"root.item.#(k%\"9\").k": "9",
	} {
		if r := Get(xml, path); r.Err != nil || r.String() != want {
			t.Errorf("%s: got %q, err %v, want %q", path, r.String(), r.Err, want)
		}
		if r := GetWithOptions(xml, path, &Options{CaseSensitive: false}); r.Err != nil || r.String() != want {
			t.Errorf("%s (options): got %q, err %v, want %q", path, r.String(), r.Err, want)
		}
	}
	if r := Get(fragment, "a.1"); r.Err != nil || r.String() != "1" {
		t.Errorf("fragment index: got %q, err %v", r.String(), r.Err)
	}

	// All-match filters count only matching siblings
	if r := Get(xml, "root.item.#(k<3)#.k"); r.Err != nil || len(r.Array()) != 3 {
		t.Errorf("all-match filter: got %v, err %v", r.Array(), r.Err)
	}
	requireLimitError(t, Get(xml, "root.item.#(k>0)#.k").Err, "MaxArrayResults")
	requireLimitError(t, GetWithOptions(xml, "root.item.#(k>0)#.k", &Options{CaseSensitive: false}).Err, "MaxArrayResults")
}

func TestLimits_PatternBudgetScalesWithValue(t *testing.T) {
	// Linear-time matches complete regardless of the value length
	defer withLimit(&MaxPatternIterations, 10)()
	xml := "<root><v>" + strings.Repeat("a", 100000) + "end</v></root>"
	for _, pattern := range []string{"a*", "*", "*end", "a*e?d", "*a*end"} {
		if r := Get(xml, `root.#(v%"`+pattern+`")`); r.Err != nil || !r.Exists() {
			t.Errorf("pattern %q: exists=%v err=%v", pattern, r.Exists(), r.Err)
		}
	}
}

func TestLimits_RecursiveOperationBudget(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("<root>")
	for i := 0; i < 20000; i++ {
		sb.WriteString("<x><y><z>1</z></y></x>")
	}
	sb.WriteString("<x><y><target>found</target></y></x></root>")
	xml := sb.String()

	// The budget scales with the document size, so large documents are fully searched
	if got := Get(xml, "root.**.target").String(); got != "found" {
		t.Fatalf("recursive search should find target, got %q", got)
	}

	// A small document cannot exceed the minimum budget
	defer withLimit(&MaxRecursiveOperations, 10)()
	if got := Get("<r><a><b>v</b></a></r>", "r.**.b").String(); got != "v" {
		t.Fatalf("recursive search on small document should succeed, got %q", got)
	}

	// Chained recursive searches share one budget per query: each "**" alone stays
	// within the document size, but searching below every nested match does not
	nested := "<r>" + strings.Repeat("<a>", 60) + "v" + strings.Repeat("</a>", 60) + "</r>"
	if r := Get(nested, "r.**.a"); r.Err != nil {
		t.Fatalf("single recursive search should succeed: %v", r.Err)
	}
	requireLimitError(t, Get(nested, "r.**.a.**.a").Err, "MaxRecursiveOperations")
	requireLimitError(t, GetWithOptions(nested, "r.**.a.**.a", &Options{CaseSensitive: false}).Err, "MaxRecursiveOperations")
}

func TestLimits_NestingDepth(t *testing.T) {
	xml := "<r>" + strings.Repeat("<n>", 20) + "v" + strings.Repeat("</n>", 20) + "</r>"
	defer withLimit(&MaxNestingDepth, 10)()

	requireLimitError(t, Get(xml, "r.**.missing").Err, "MaxNestingDepth")

	_, err := Set(xml, "r.x", "1")
	requireLimitError(t, err, "MaxNestingDepth")
}

func TestLimits_ErrPropagatesThroughResult(t *testing.T) {
	defer withLimit(&MaxArrayResults, 1)()
	r := Get(`<root><a>1</a><a>2</a></root>`, "root.a.#")
	requireLimitError(t, r.Get("x").Err, "MaxArrayResults")
	requireLimitError(t, r.GetWithOptions("x", &Options{CaseSensitive: false}).Err, "MaxArrayResults")
	for _, mr := range r.GetMany("x", "y") {
		requireLimitError(t, mr.Err, "MaxArrayResults")
	}
}

func TestLimits_MapReturnsNil(t *testing.T) {
	root := Get(`<root><a>1</a><b>2</b><c>3</c></root>`, "root")
	if m := root.Map(); m == nil || len(m) != 3 {
		t.Fatalf("Map should return 3 children, got %v", m)
	}
	if m := Get(`<root></root>`, "root").Map(); m == nil {
		t.Fatal("Map of element without children should be empty, not nil")
	}

	defer withLimit(&MaxArrayResults, 2)()
	if m := root.Map(); m != nil {
		t.Errorf("Map should return nil when MaxArrayResults is exceeded, got %v", m)
	}
	if m := root.MapWithOptions(&Options{CaseSensitive: false}); m != nil {
		t.Errorf("MapWithOptions should return nil when MaxArrayResults is exceeded, got %v", m)
	}
}

func TestLimits_SetReturnsLimitError(t *testing.T) {
	xml := `<root><item>a</item><item>b</item><item>c</item></root>`

	tests := []struct {
		name  string
		limit *int
		value int
		fn    func() (string, error)
		want  string
	}{
		{"document size", &MaxDocumentSize, 10, func() (string, error) { return Set(xml, "root.x", "1") }, "MaxDocumentSize"},
		{"value size", &MaxValueSize, 3, func() (string, error) { return Set(xml, "root.x", "value") }, "MaxValueSize"},
		{"result size", &MaxDocumentSize, len(xml) + 5, func() (string, error) { return Set(xml, "root.x", "value") }, "MaxDocumentSize"},
		{"delete document size", &MaxDocumentSize, 10, func() (string, error) { return Delete(xml, "root.item") }, "MaxDocumentSize"},
		{"path segments", &MaxPathSegments, 1, func() (string, error) { return Set(xml, "root.x", "1") }, "MaxPathSegments"},
		{"set many", &MaxDocumentSize, 10, func() (string, error) { return SetMany(xml, []string{"root.x"}, []interface{}{"1"}) }, "MaxDocumentSize"},
		{"delete many", &MaxDocumentSize, 10, func() (string, error) { return DeleteMany(xml, "root.item") }, "MaxDocumentSize"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.fn(); err != nil {
				t.Fatalf("operation should succeed with default limits: %v", err)
			}

			defer withLimit(tt.limit, tt.value)()
			out, err := tt.fn()
			requireLimitError(t, err, tt.want)
			if out != xml {
				t.Errorf("expected original document on error, got %q", out)
			}
		})
	}
}

func TestLimits_ValidateReportsLimit(t *testing.T) {
	defer withLimit(&MaxAttributes, 1)()
	verr := ValidateWithError(`<root a="1" b="2"/>`)
	if verr == nil {
		t.Fatal("expected validation error")
	}
	requireLimitError(t, verr, "MaxAttributes")

	if verr := ValidateWithError(`<root><a></root>`); verr == nil || errors.Is(verr, ErrLimitExceeded) {
		t.Errorf("malformed XML should not be reported as a limit error, got %v", verr)
	}
}

func TestLimits_OtherPanicsPropagate(t *testing.T) {
	if err := RegisterModifier("panics", NewModifierFunc("panics", panicModifier)); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = UnregisterModifier("panics") }()

	if r := panicValue(func() { Get(`<root>v</root>`, "root|@panics") }); r != "boom" {
		t.Errorf("expected modifier panic to propagate, got %v", r)
	}
}

// panicModifier is a modifier that always panics with "boom".
func panicModifier(Result) Result {
	panic("boom")
}

// panicValue runs fn and returns the value it panicked with, or nil.
func panicValue(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// parseFilterConditionChecked calls parseFilterCondition, converting limit
// violations into errors as the public API does.
func parseFilterConditionChecked(expr string) (filter *Filter, err error) {
	defer recoverLimit(&err)
	return parseFilterCondition(expr)
}

// parsePathChecked calls parsePath, returning nil segments on limit violations
// as the public API does.
func parsePathChecked(path string) (segments []PathSegment) {
	var err error
	defer func() {
		if err != nil {
			segments = nil
		}
	}()
	defer recoverLimit(&err)
	return parsePath(path)
}

// splitNamespaceChecked calls splitNamespace, returning ("", name) on limit violations.
func splitNamespaceChecked(name string) (prefix, localName string) {
	var err error
	defer func() {
		if err != nil {
			prefix, localName = "", name
		}
	}()
	defer recoverLimit(&err)
	return splitNamespace(name)
}

func TestLimits_ReviewRegressions(t *testing.T) {
	t.Run("ValidateError Unwrap on nil", func(t *testing.T) {
		if errors.Is(ValidateWithError("<r/>"), ErrLimitExceeded) {
			t.Error("valid document must not report a limit error")
		}
	})

	t.Run("long namespace prefix in document", func(t *testing.T) {
		prefix := strings.Repeat("p", MaxNamespacePrefixLength+1)
		doc := "<root><" + prefix + ":x>a</" + prefix + ":x><y>b</y></root>"
		if r := Get(doc, "root.y"); r.Err != nil || r.String() != "b" {
			t.Errorf("Get: got %q, err %v", r.String(), r.Err)
		}
		if _, err := Set(doc, "root.y", "c"); err != nil {
			t.Errorf("Set: %v", err)
		}
		// The limit still applies to prefixes in paths
		requireLimitError(t, Get(doc, "root."+prefix+":x").Err, "MaxNamespacePrefixLength")
	})

	t.Run("field name length", func(t *testing.T) {
		requireLimitError(t, Get("<r><i><a>1</a></i></r>", "r.i.#."+strings.Repeat("a", MaxFieldNameLength+1)).Err, "MaxFieldNameLength")
	})

	t.Run("Map on failed result", func(t *testing.T) {
		defer withLimit(&MaxArrayResults, 1)()
		r := Get("<r><a>1</a><a>2</a></r>", "r.a.#")
		if m := r.Map(); m != nil {
			t.Errorf("Map of failed result should be nil, got %v", m)
		}
		if m := r.MapWithOptions(&Options{CaseSensitive: false}); m != nil {
			t.Errorf("MapWithOptions of failed result should be nil, got %v", m)
		}
	})

	t.Run("Result JSON omits Err", func(t *testing.T) {
		b, err := json.Marshal(Result{Type: Null, Err: ErrLimitExceeded})
		if err != nil || strings.Contains(string(b), "Err") {
			t.Errorf("unexpected JSON %s (err %v)", b, err)
		}
	})

	t.Run("malformed attribute after limit is malformed", func(t *testing.T) {
		defer withLimit(&MaxAttributes, 3)()
		for _, doc := range []string{
			`<y a0="v" a1="v" a2="v" ?>`,
			`<y a0="v" a1="v" a2="v" 1bad="x">`,
			`<y a0="v" a1="v" a2="v" b=unquoted>`,
		} {
			verr := ValidateWithError(doc)
			if verr == nil || errors.Is(verr, ErrLimitExceeded) {
				t.Errorf("%s: expected malformed error, got %v", doc, verr)
			}
			if _, err := Set(doc, "y.z", "1"); !errors.Is(err, ErrMalformedXML) {
				t.Errorf("%s: Set expected ErrMalformedXML, got %v", doc, err)
			}
		}
		requireLimitError(t, ValidateWithError(`<y a0="v" a1="v" a2="v" a3="v"/>`), "MaxAttributes")
	})

	t.Run("processing instruction token size agrees with validation", func(t *testing.T) {
		defer withLimit(&MaxTokenSize, 10)()
		// The body "ppppppppp?" is exactly MaxTokenSize bytes
		doc := "<?" + strings.Repeat("p", 9) + "?><r>v</r>"
		if !Valid(doc) {
			t.Fatal("document should be valid")
		}
		over := "<?" + strings.Repeat("p", 10) + "?><r>v</r>"
		requireLimitError(t, ValidateWithError(over), "MaxTokenSize")
		requireLimitError(t, Get(over, "r").Err, "MaxTokenSize")
		if r := Get(doc, "r"); r.Err != nil || r.String() != "v" {
			t.Errorf("Get: got %q, err %v", r.String(), r.Err)
		}
		if _, err := Set(doc, "r", "w"); err != nil {
			t.Errorf("Set: %v", err)
		}
	})

	t.Run("Set result must satisfy limits", func(t *testing.T) {
		func() {
			defer withLimit(&MaxAttributes, 2)()
			_, err := Set(`<r a="1" b="2"/>`, "r.@c", "3")
			requireLimitError(t, err, "MaxAttributes")
		}()
		func() {
			defer withLimit(&MaxTokenSize, 10)()
			_, err := Set(`<r/>`, "r.@a", strings.Repeat("v", 20))
			requireLimitError(t, err, "MaxTokenSize")
		}()
		func() {
			defer withLimit(&MaxNestingDepth, 3)()
			_, err := SetRaw(`<r><a></a></r>`, "r.a.b", "<c><d>1</d></c>")
			requireLimitError(t, err, "MaxNestingDepth")
		}()
	})

	t.Run("limit exactly at boundary is accepted", func(t *testing.T) {
		doc := `<r><a>1</a></r>`
		func() {
			defer withLimit(&MaxDocumentSize, len(doc))()
			if r := Get(doc, "r.a"); r.Err != nil {
				t.Errorf("document at MaxDocumentSize: %v", r.Err)
			}
		}()
		func() {
			defer withLimit(&MaxValueSize, 5)()
			if _, err := Set(doc, "r.a", "12345"); err != nil {
				t.Errorf("value at MaxValueSize: %v", err)
			}
			_, err := Set(doc, "r.a.-1", "123456")
			requireLimitError(t, err, "MaxValueSize")
		}()
		func() {
			defer withLimit(&MaxWildcardResults, 2)()
			xml := "<r><a>1</a><a>2</a></r>"
			if r := Get(xml, "r.**.a"); r.Err != nil || len(r.Array()) != 2 {
				t.Errorf("results at MaxWildcardResults: %v %v", r.Array(), r.Err)
			}
			requireLimitError(t, Get("<r><a>1</a><a>2</a><a>3</a></r>", "r.**.a").Err, "MaxWildcardResults")
		}()
	})

	t.Run("Delete and GetWithOptions recover limits", func(t *testing.T) {
		defer withLimit(&MaxPathSegments, 1)()
		_, err := Delete("<r><a>1</a></r>", "r.a")
		requireLimitError(t, err, "MaxPathSegments")
		requireLimitError(t, GetWithOptions("<r><a>1</a></r>", "r.a", &Options{CaseSensitive: false}).Err, "MaxPathSegments")
		func() {
			defer withLimit(&MaxDocumentSize, 5)()
			requireLimitError(t, GetWithOptions("<r><a>1</a></r>", "r.a", &Options{CaseSensitive: false}).Err, "MaxDocumentSize")
			requireLimitError(t, Get("<r><a>1</a></r>", "r").Get("a").Err, "MaxDocumentSize")
		}()
	})

	t.Run("markup inside comments and CDATA is not parsed", func(t *testing.T) {
		for name, inner := range map[string]string{
			"CDATA":   "<![CDATA[" + strings.Repeat("<br>", MaxNestingDepth+50) + `<x y="` + "]]>",
			"comment": "<!-- " + strings.Repeat("<br>", MaxNestingDepth+50) + `<x y="` + " -->",
			"closing": "<![CDATA[</a>]]>",
		} {
			doc := "<r><a>" + inner + "</a><b>ok</b></r>"
			if !Valid(doc) {
				t.Fatalf("%s: document should be valid", name)
			}
			if r := Get(doc, "r.b"); r.Err != nil || r.String() != "ok" {
				t.Errorf("%s: Get r.b = %q, err %v", name, r.String(), r.Err)
			}
			if r := Get(doc, "r.a"); r.Err != nil || r.Raw != inner {
				t.Errorf("%s: Get r.a Raw = %q, err %v", name, r.Raw, r.Err)
			}
			if _, err := Set(doc, "r.b", "new"); err != nil {
				t.Errorf("%s: Set: %v", name, err)
			}
		}
	})

	t.Run("Map document size", func(t *testing.T) {
		root := Get("<r><a>1</a><b>2</b></r>", "r")
		defer withLimit(&MaxDocumentSize, 5)()
		if m := root.Map(); m != nil {
			t.Errorf("Map over MaxDocumentSize should be nil, got %v", m)
		}
	})
}
