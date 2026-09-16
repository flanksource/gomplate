package gomplate

import (
	"fmt"
	"sync"
	"testing"
	"time"

	_ "github.com/flanksource/gomplate/v3/js"
	_ "github.com/robertkrimen/otto/underscore"
)

func TestCacheKeyConsistency(t *testing.T) {
	var hello = func() any {
		return "world"
	}

	var foo = func() any {
		return "bar"
	}

	{
		tt := Template{
			Expression: "{{.name}}{{.age}}",
			Functions: map[string]any{
				"hello": hello,
				"Hello": foo,
				"foo":   foo,
				"Foo":   foo,
			},
		}

		expectedCacheKey := tt.cacheKey(map[string]any{"age": 19, "name": "james"})
		for i := 0; i < 10; i++ {
			key := tt.cacheKey(map[string]any{"age": 19, "name": "james"})
			if key != expectedCacheKey {
				t.Errorf("cache key mismatch: %s != %s", key, expectedCacheKey)
			}
		}
	}

	{
		tt := Template{
			Template:   "{{.name}}{{.age}}",
			LeftDelim:  "{{",
			RightDelim: "}}",
			Functions: map[string]any{
				"hello": hello,
				"Hello": foo,
				"foo":   foo,
				"Foo":   foo,
			},
		}

		expectCacheKey := tt.cacheKey(map[string]any{"age": 19, "name": "james"})
		for i := 0; i < 10; i++ {
			key := tt.cacheKey(map[string]any{"age": 19, "name": "james"})
			if key != expectCacheKey {
				t.Errorf("cache key mismatch: %s != %s", key, expectCacheKey)
			}
		}
	}
}

func TestExplicitCacheKey(t *testing.T) {
	tt := Template{
		Expression: "name + age",
		CacheKey:   "user-defined-key",
	}

	if got := tt.cacheKey(map[string]any{"foo": 1}); got != "user-defined-key" {
		t.Errorf("expected explicit CacheKey to be used, got %q", got)
	}

	if got := tt.cacheKey(map[string]any{"bar": 2}); got != "user-defined-key" {
		t.Errorf("explicit CacheKey must be stable across env shapes, got %q", got)
	}

	if !tt.IsCacheable() {
		t.Errorf("template with explicit CacheKey must be cacheable")
	}

	withFuncs := Template{
		Expression: "name",
		Functions:  map[string]any{"hello": func() any { return "world" }},
	}
	if withFuncs.IsCacheable() {
		t.Errorf("template with Functions and no CacheKey must not be cacheable")
	}
	withFuncs.CacheKey = "stable"
	if !withFuncs.IsCacheable() {
		t.Errorf("template with Functions but explicit CacheKey must be cacheable")
	}
}

func TestCacheTime(t *testing.T) {
	// No expiration: cache entry should have zero expiration time.
	{
		tpl := Template{
			Expression: "1 + 1",
			CacheKey:   "cachetime-noexp",
			CacheTime:  -1,
		}
		if _, err := RunExpression(nil, tpl); err != nil {
			t.Fatalf("eval: %v", err)
		}
		_, exp, ok := celExpressionCache.GetWithExpiration(tpl.celCacheKey(nil, currentNativeTypes().generation))
		if !ok {
			t.Fatalf("entry not cached")
		}
		if !exp.IsZero() {
			t.Errorf("expected no-expiration entry, got expiry %v", exp)
		}
	}

	// Explicit short TTL: entry expiration should be close to now+CacheTime.
	{
		tpl := Template{
			Expression: "1 + 1",
			CacheKey:   "cachetime-short",
			CacheTime:  50 * time.Millisecond,
		}
		before := time.Now()
		if _, err := RunExpression(nil, tpl); err != nil {
			t.Fatalf("eval: %v", err)
		}
		_, exp, ok := celExpressionCache.GetWithExpiration(tpl.celCacheKey(nil, currentNativeTypes().generation))
		if !ok {
			t.Fatalf("entry not cached")
		}
		diff := exp.Sub(before)
		if diff < 40*time.Millisecond || diff > 200*time.Millisecond {
			t.Errorf("expected expiry ~50ms after set, got %v", diff)
		}
	}

	// Zero CacheTime: should fall back to the cache's default TTL (~1h).
	{
		tpl := Template{
			Expression: "1 + 1",
			CacheKey:   "cachetime-default",
		}
		before := time.Now()
		if _, err := RunExpression(nil, tpl); err != nil {
			t.Fatalf("eval: %v", err)
		}
		_, exp, ok := celExpressionCache.GetWithExpiration(tpl.celCacheKey(nil, currentNativeTypes().generation))
		if !ok {
			t.Fatalf("entry not cached")
		}
		diff := exp.Sub(before)
		if diff < 30*time.Minute || diff > 90*time.Minute {
			t.Errorf("expected ~1h default TTL, got %v", diff)
		}
	}

}

func TestRunTemplate_ValueFunctions(t *testing.T) {
	out, err := RunTemplate(map[string]any{"msg": "world"}, Template{
		Template:       "hello {{ msg }}",
		ValueFunctions: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello world" {
		t.Errorf("got %q, want %q", out, "hello world")
	}
}

func TestRunTemplate_ValueFunctionsCoexistWithDotAccess(t *testing.T) {
	out, err := RunTemplate(map[string]any{"msg": "world"}, Template{
		Template:       "{{ msg }}-{{ .msg }}",
		ValueFunctions: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "world-world" {
		t.Errorf("got %q, want %q", out, "world-world")
	}
}

func TestRunTemplate_ValueFunctionsOffKeepsDotOnly(t *testing.T) {
	out, err := RunTemplate(map[string]any{"msg": "world"}, Template{
		Template: "{{ .msg }}",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "world" {
		t.Errorf("got %q, want %q", out, "world")
	}

	if _, err := RunTemplate(map[string]any{"msg": "world"}, Template{
		Template: "{{ msg }}",
	}); err == nil {
		t.Error("expected error calling bare {{ msg }} without ValueFunctions, got nil")
	}
}

func TestRunTemplate_ValueFunctionsDoNotMutateCallerFuncs(t *testing.T) {
	callerFuncs := map[string]any{}
	_, err := RunTemplate(map[string]any{"msg": "world"}, Template{
		Template:       "{{ msg }}",
		Functions:      callerFuncs,
		ValueFunctions: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(callerFuncs) != 0 {
		t.Errorf("caller Functions map was mutated: %v", callerFuncs)
	}
}

func TestRunTemplate_DelimSetsMultiPass(t *testing.T) {
	out, err := RunTemplate(map[string]any{"msg": "world"}, Template{
		Template: "hello $(msg)",
		DelimSets: []Delims{
			{Left: "{{", Right: "}}"},
			{Left: "$(", Right: ")"},
		},
		ValueFunctions: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello world" {
		t.Errorf("got %q, want %q", out, "hello world")
	}
}

func TestRunTemplate_DelimSetsFeedsOutputForward(t *testing.T) {
	// First pass replaces $(inner) with "msg", producing "{{ msg }}";
	// second pass then resolves the {{ msg }}.
	out, err := RunTemplate(
		map[string]any{"inner": "msg", "msg": "world"},
		Template{
			Template: "{{ $(inner) }}",
			DelimSets: []Delims{
				{Left: "$(", Right: ")"},
				{Left: "{{", Right: "}}"},
			},
			ValueFunctions: true,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "world" {
		t.Errorf("got %q, want %q", out, "world")
	}
}

func TestRunTemplate_HeaderOverridesDelimSets(t *testing.T) {
	// Header sets [[ ]] so DelimSets should be ignored and $(...) left literal.
	out, err := RunTemplate(map[string]any{"msg": "world"}, Template{
		Template: "# gotemplate: left-delim=[[ right-delim=]]\n[[ .msg ]] $(msg)",
		DelimSets: []Delims{
			{Left: "{{", Right: "}}"},
			{Left: "$(", Right: ")"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "world $(msg)" {
		t.Errorf("got %q, want %q", out, "world $(msg)")
	}
}

func TestRunTemplate_DelimSetsPreferredOverSingleDelimPair(t *testing.T) {
	// Both DelimSets and LeftDelim/RightDelim set — DelimSets wins.
	out, err := RunTemplate(map[string]any{"msg": "world"}, Template{
		Template:   "$(msg)",
		LeftDelim:  "[[",
		RightDelim: "]]",
		DelimSets: []Delims{
			{Left: "$(", Right: ")"},
		},
		ValueFunctions: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "world" {
		t.Errorf("got %q, want %q", out, "world")
	}
}

func TestRunTemplate_SingleDelimPairStillWorks(t *testing.T) {
	out, err := RunTemplate(map[string]any{"msg": "world"}, Template{
		Template:   "[[ .msg ]]",
		LeftDelim:  "[[",
		RightDelim: "]]",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "world" {
		t.Errorf("got %q, want %q", out, "world")
	}
}

func TestRunExpressionReusesProgramAcrossDifferentData(t *testing.T) {
	tpl := Template{
		Expression: "name + age",
		CacheKey:   "reuse-test",
	}

	out1, err := RunExpression(map[string]any{"name": "alice-", "age": "30"}, tpl)
	if err != nil {
		t.Fatalf("first eval: %v", err)
	}
	if out1 != "alice-30" {
		t.Errorf("first eval result: got %v", out1)
	}

	out2, err := RunExpression(map[string]any{"name": "bob-", "age": "42"}, tpl)
	if err != nil {
		t.Fatalf("second eval: %v", err)
	}
	if out2 != "bob-42" {
		t.Errorf("second eval result: got %v (cached program should not leak first-call data)", out2)
	}
}

// The go template path must ignore CacheKey. A parsed go template holds the
// Functions it was parsed with, and CacheKey used to bypass both the check that
// keeps such templates out of the cache and the key that keeps distinct parses
// apart. The tests below cover each way that used to go wrong.

func TestGoTemplate_CacheKeyDoesNotReuseValueFunctions(t *testing.T) {
	template := Template{
		Template:       "node://$(tags.cluster)/$(.name)",
		CacheKey:       "gotemplate-value-functions",
		CacheTime:      time.Hour,
		ValueFunctions: true,
		DelimSets:      []Delims{{Left: "$(", Right: ")"}},
	}

	first, err := RunTemplate(map[string]any{
		"name": "node-a",
		"tags": map[string]string{"cluster": "production"},
	}, template)
	if err != nil {
		t.Fatalf("first render: %v", err)
	}
	if first != "node://production/node-a" {
		t.Errorf("first render: got %q, want %q", first, "node://production/node-a")
	}

	second, err := RunTemplate(map[string]any{
		"name": "node-b",
		"tags": map[string]string{"cluster": "staging"},
	}, template)
	if err != nil {
		t.Fatalf("second render: %v", err)
	}
	if second != "node://staging/node-b" {
		t.Errorf("second render used the first render's value functions: got %q, want %q",
			second, "node://staging/node-b")
	}
}

func TestGoTemplate_CacheKeyDoesNotReuseCustomFunctions(t *testing.T) {
	forUser := func(user string) Template {
		return Template{
			Template:  "user is {{ whoami }}",
			CacheKey:  "gotemplate-custom-functions",
			CacheTime: time.Hour,
			Functions: map[string]any{
				"whoami": func() any { return user },
			},
		}
	}

	first, err := RunTemplate(map[string]any{}, forUser("alice"))
	if err != nil {
		t.Fatalf("first render: %v", err)
	}
	if first != "user is alice" {
		t.Errorf("first render: got %q, want %q", first, "user is alice")
	}

	second, err := RunTemplate(map[string]any{}, forUser("bob"))
	if err != nil {
		t.Fatalf("second render: %v", err)
	}
	if second != "user is bob" {
		t.Errorf("second render used the first render's functions: got %q, want %q",
			second, "user is bob")
	}
}

func TestGoTemplate_CacheKeyDoesNotCollideAcrossDelimiterPasses(t *testing.T) {
	// Each pass parses different text with different delimiters. Sharing one
	// CacheKey made the second pass reuse the first pass's parsed template, so
	// the $( ) placeholder was never resolved.
	out, err := RunTemplate(map[string]any{}, Template{
		Template:  "A={{ 1 }} B=$( 2 )",
		CacheKey:  "gotemplate-delimiter-passes",
		CacheTime: time.Hour,
		DelimSets: []Delims{
			{Left: "{{", Right: "}}"},
			{Left: "$(", Right: ")"},
		},
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if out != "A=1 B=2" {
		t.Errorf("second delimiter pass did not run: got %q, want %q", out, "A=1 B=2")
	}
}

func TestGoTemplate_CacheKeyDoesNotCollideAcrossTemplates(t *testing.T) {
	first, err := RunTemplate(map[string]any{}, Template{
		Template:  "I am template A",
		CacheKey:  "gotemplate-shared-label",
		CacheTime: time.Hour,
	})
	if err != nil {
		t.Fatalf("first render: %v", err)
	}
	if first != "I am template A" {
		t.Errorf("first render: got %q, want %q", first, "I am template A")
	}

	second, err := RunTemplate(map[string]any{}, Template{
		Template:  "I am template B",
		CacheKey:  "gotemplate-shared-label",
		CacheTime: time.Hour,
	})
	if err != nil {
		t.Fatalf("second render: %v", err)
	}
	if second != "I am template B" {
		t.Errorf("templates sharing a CacheKey collided: got %q, want %q",
			second, "I am template B")
	}
}

func TestGoTemplate_CacheKeyIsSafeForConcurrentRenders(t *testing.T) {
	template := Template{
		Template:       "$(.name)=$(tags.cluster)",
		CacheKey:       "gotemplate-concurrent",
		CacheTime:      time.Hour,
		ValueFunctions: true,
		DelimSets:      []Delims{{Left: "$(", Right: ")"}},
	}

	const workers = 8
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("node-%d", i)
			cluster := fmt.Sprintf("cluster-%d", i)
			want := name + "=" + cluster
			for j := 0; j < 100; j++ {
				got, err := RunTemplate(map[string]any{
					"name": name,
					"tags": map[string]string{"cluster": cluster},
				}, template)
				if err != nil {
					errs <- err
					return
				}
				if got != want {
					errs <- fmt.Errorf("got %q, want %q", got, want)
					return
				}
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

// isGoTemplateCacheable must not honour CacheKey, unlike IsCacheable.
func TestIsGoTemplateCacheable(t *testing.T) {
	withFuncs := Template{
		Template:  "{{ hello }}",
		Functions: map[string]any{"hello": func() any { return "world" }},
		CacheKey:  "stable",
	}
	if !withFuncs.IsCacheable() {
		t.Error("IsCacheable must still honour CacheKey for cel expressions")
	}
	if withFuncs.isGoTemplateCacheable() {
		t.Error("a go template with Functions must not be cached, even with a CacheKey")
	}

	plain := Template{Template: "{{ .name }}", CacheKey: "stable"}
	if !plain.isGoTemplateCacheable() {
		t.Error("a go template without Functions must be cacheable")
	}
}
