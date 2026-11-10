package releasepolicy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

func TestReleasePolicy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Release policy")
}

var _ = Describe("major version release policy", func() {
	root := filepath.Join("..", "..")
	It("keeps maintenance releases in v3 and recognizes breaking releases", func() {
		content, err := os.ReadFile(filepath.Join(root, ".releaserc"))
		Expect(err).NotTo(HaveOccurred())
		var config struct {
			Branches []struct {
				Name  string
				Range string
			}
			Plugins []any
		}
		Expect(yaml.Unmarshal(content, &config)).To(Succeed())
		Expect(config.Branches).To(ContainElement(struct {
			Name  string
			Range string
		}{"3.x", "3.x"}))
		analyzer := config.Plugins[0].([]any)[1].(map[string]any)
		Expect(analyzer["releaseRules"]).To(ContainElement(map[string]any{"breaking": true, "release": "major"}))
	})
	It("disables routine maintenance updates while retaining reviewed security updates", func() {
		content, err := os.ReadFile(filepath.Join(root, "renovate.json"))
		Expect(err).NotTo(HaveOccurred())
		var config map[string]any
		Expect(json.Unmarshal(content, &config)).To(Succeed())
		Expect(config["baseBranchPatterns"]).To(ConsistOf("main", "3.x"))
		Expect(config["osvVulnerabilityAlerts"]).To(Equal(true))
		Expect(config["packageRules"]).To(ContainElement(SatisfyAll(
			HaveKeyWithValue("matchBaseBranches", []any{"3.x"}),
			HaveKeyWithValue("matchPackageNames", []any{"*"}),
			HaveKeyWithValue("enabled", false),
		)))
		alerts := config["vulnerabilityAlerts"].(map[string]any)
		Expect(alerts).To(HaveKeyWithValue("enabled", true))
		Expect(alerts).To(HaveKeyWithValue("automerge", false))
		Expect(alerts).To(HaveKeyWithValue("vulnerabilityFixStrategy", "lowest"))
	})
})
