package grammar

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	ginkgo "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("CEL.g4", func() {
	ginkgo.It("stays byte-identical to the copy cel-go ships", func() {
		// The whole point of generating from the grammar is that a cel-go bump
		// which changes the lexer fails here instead of silently desyncing the
		// editor from the parser.
		upstream, err := upstreamCELGrammarPath()
		if err != nil {
			ginkgo.Skip("cel-go module source not available: " + err.Error())
		}
		want, err := os.ReadFile(upstream)
		Expect(err).ToNot(HaveOccurred())
		Expect(CELGrammarSource()).To(Equal(string(want)),
			"vendored CEL.g4 is stale -- copy %s over genmonarch/grammar/CEL.g4 and re-run `make monarch`", upstream)
	})

	ginkgo.Context("parsed into a lexical vocabulary", func() {
		var g *CELGrammar

		ginkgo.BeforeEach(func() {
			var err error
			g, err = ParseCEL()
			Expect(err).ToNot(HaveOccurred())
		})

		ginkgo.It("extracts word keywords separately from operators", func() {
			Expect(g.Keywords).To(ConsistOf("in", "true", "false", "null"))
		})

		ginkgo.It("orders operators longest-first so `<=` is never shadowed by `<`", func() {
			Expect(g.Operators).To(ContainElements("==", "!=", "<=", ">=", "&&", "||", "?", ":", "%"))
			Expect(indexOf(g.Operators, "<=")).To(BeNumerically("<", indexOf(g.Operators, "<")))
			Expect(indexOf(g.Operators, ">=")).To(BeNumerically("<", indexOf(g.Operators, ">")))
		})

		ginkgo.It("excludes inline-only fragments from the operator set", func() {
			// `fragment BACKSLASH : '\\'` is a building block of the escape
			// sequences, not a CEL operator.
			Expect(g.Operators).ToNot(ContainElement(`\`))
		})

		ginkgo.It("matches the longest string form rather than stopping at the shortest", func() {
			// Unanchored, as a Monarch tokenizer consumes it.
			for _, tc := range []struct{ input, want string }{
				{`"""hello"""`, `"""hello"""`},
				{`'''hello'''`, `'''hello'''`},
				{`r"""a\db"""`, `r"""a\db"""`},
				{`"plain"`, `"plain"`},
			} {
				Expect(firstMatch(g.Patterns["STRING"], tc.input)).To(Equal(tc.want))
			}
		})

		ginkgo.It("matches a hex integer whole rather than stopping after the leading 0", func() {
			Expect(firstMatch(g.Patterns["NUM_INT"], "0x1f")).To(Equal("0x1f"))
			Expect(firstMatch(g.Patterns["NUM_UINT"], "0x1fu")).To(Equal("0x1fu"))
		})

		ginkgo.It("translates the simple token rules", func() {
			Expect(g.Patterns["COMMENT"]).To(Equal(`//[^\n]*`))
			Expect(g.Patterns["IDENTIFIER"]).To(Equal(`[A-Za-z_][A-Za-z0-9_]*`))
			Expect(g.Patterns["WHITESPACE"]).To(Equal(`[\t \r\n\f]+`))
		})

		ginkgo.It("recovers back-tick escaped identifiers", func() {
			// Easy to miss by hand -- `k8s.labels` may be written `` `k8s.labels` ``.
			Expect(g.Patterns).To(HaveKey("ESC_IDENTIFIER"))
			Expect(matches(g.Patterns["ESC_IDENTIFIER"], "`some.escaped-id/v1`")).To(BeTrue())
		})

		ginkgo.It("emits patterns that are all valid regular expressions", func() {
			for name, pattern := range g.Patterns {
				_, err := regexp.Compile(pattern)
				Expect(err).ToNot(HaveOccurred(), "rule %s produced invalid regex %q", name, pattern)
			}
		})

		ginkgo.DescribeTable("the STRING rule covers every CEL literal form",
			func(literal string) {
				Expect(matches(g.Patterns["STRING"], literal)).To(BeTrue())
			},
			ginkgo.Entry("double quoted", `"hello"`),
			ginkgo.Entry("single quoted", `'hello'`),
			ginkgo.Entry("triple double quoted", `"""hello"""`),
			ginkgo.Entry("triple single quoted", `'''hello'''`),
			ginkgo.Entry("raw", `r"a\db"`),
			ginkgo.Entry("raw upper", `R'a\db'`),
			ginkgo.Entry("raw triple", `r"""a\db"""`),
			ginkgo.Entry("hex escape", `"\x41"`),
			ginkgo.Entry("unicode escape", `"A"`),
			ginkgo.Entry("long unicode escape", `"\U0001F600"`),
			ginkgo.Entry("octal escape", `"\101"`),
		)

		ginkgo.DescribeTable("the numeric rules separate int, uint and float",
			func(rule, literal string) {
				Expect(matches(g.Patterns[rule], literal)).To(BeTrue())
			},
			ginkgo.Entry("decimal int", "NUM_INT", "123"),
			ginkgo.Entry("hex int", "NUM_INT", "0x1f"),
			ginkgo.Entry("uint suffix", "NUM_UINT", "123u"),
			ginkgo.Entry("hex uint", "NUM_UINT", "0x1fU"),
			ginkgo.Entry("float", "NUM_FLOAT", "1.5"),
			ginkgo.Entry("float exponent", "NUM_FLOAT", "1e-3"),
			ginkgo.Entry("leading dot float", "NUM_FLOAT", ".5e3"),
		)

		ginkgo.It("marks bytes literals distinctly from strings", func() {
			Expect(matches(g.Patterns["BYTES"], `b"abc"`)).To(BeTrue())
			Expect(matches(g.Patterns["BYTES"], `B'abc'`)).To(BeTrue())
		})
	})
})

// firstMatch returns what pattern consumes from the start of s, the way a
// Monarch tokenizer applies it -- anchored at the start only, so an alternation
// ordered wrongly shows up as a short match instead of being hidden by
// backtracking.
func firstMatch(pattern, s string) string {
	re, err := regexp.Compile("^(?:" + pattern + ")")
	if err != nil {
		return "<invalid regex: " + err.Error() + ">"
	}
	return re.FindString(s)
}

// matches reports whether pattern matches the whole of s.
func matches(pattern, s string) bool {
	re, err := regexp.Compile("^(?:" + pattern + ")$")
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

func indexOf(haystack []string, needle string) int {
	for i, s := range haystack {
		if s == needle {
			return i
		}
	}
	return -1
}

// upstreamCELGrammarPath locates CEL.g4 inside the resolved cel-go module.
func upstreamCELGrammarPath() (string, error) {
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/google/cel-go").Output()
	if err != nil {
		return "", err
	}
	dir := strings.TrimSpace(string(out))
	if dir == "" {
		return "", os.ErrNotExist
	}
	return filepath.Join(dir, "parser", "gen", "CEL.g4"), nil
}
