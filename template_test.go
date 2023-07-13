package gomplate

import (
	"testing"
	"time"

	"github.com/flanksource/gomplate/v3/k8s"
	"github.com/stretchr/testify/assert"
)

func TestGomplate(t *testing.T) {
	tests := []struct {
		env      map[string]interface{}
		template string
		out      string
	}{
		{map[string]interface{}{"hello": "world"}, "{{ .hello }}", "world"},
		{map[string]interface{}{"age": 75 * time.Second}, "{{ .age | humanDuration  }}", "1m15s"},
		{map[string]interface{}{"healthySvc": k8s.GetUnstructured(k8s.TestHealthy)}, "{{ (.healthySvc | isHealthy) }}", "true"},
		{map[string]interface{}{"healthySvc": k8s.GetUnstructured(k8s.TestLuaStatus)}, "{{ (.healthySvc | getStatus) }}", "Degraded: found less than two generators, Merge requires two or more"},
		{map[string]interface{}{"healthySvc": k8s.GetUnstructured(k8s.TestHealthy)}, "{{ (.healthySvc | getHealth).Status  }}", "Healthy"},

		{map[string]interface{}{"size": 123456}, "{{ .size | humanSize }}", "120.6K"},
		{map[string]interface{}{"v": "1.2.3-beta.1+c0ff33"}, "{{  (.v | semver).Prerelease  }}", "beta.1"},
		{map[string]interface{}{"old": "1.2.3", "new": "1.2.3"}, "{{  .old | semverCompare .new }}", "true"},
		{map[string]interface{}{"old": "1.2.3", "new": "1.2.4"}, "{{  .old | semverCompare .new }}", "false"},
	}
	for _, tc := range tests {

		t.Run(tc.template, func(t *testing.T) {
			out, err := RunTemplate(tc.env, Template{
				Template: tc.template,
			})
			assert.ErrorIs(t, err, nil)
			assert.Equal(t, tc.out, out)
		})

	}

}
