package templates

import (
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

// Drift guard: the domain owns the canonical starter markers used for
// boilerplate-only detection and must not read these templates. This test
// pins that freshly generated blank starter content is still recognized as
// boilerplate-only, so a template wording change fails here and forces an
// intentional domain marker and spec update instead of a silent validation
// behavior change.
func TestBlankStarterContentIsRecognizedAsBoilerplateOnlyByDomain(t *testing.T) {
	content := NewDefaultBlankChangeContent()

	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		starterContent, err := content.ContentFor(requiredFile)
		if err != nil {
			t.Fatalf("ContentFor(%q) error = %v", requiredFile, err)
		}

		findings := domain.ValidateChangeFileContents([]domain.ChangeFileContent{
			{
				FileName:     requiredFile,
				RelativePath: "openspec/changes/example/" + requiredFile,
				Content:      starterContent,
			},
		})

		boilerplateFound := false
		for _, finding := range findings {
			if finding.Code == domain.ValidationFindingCodeBoilerplateOnlyContent {
				boilerplateFound = true
			}
			if finding.Severity == domain.ValidationFindingSeverityError {
				t.Fatalf("blank starter %q produced error finding %v, want warnings only", requiredFile, finding)
			}
		}
		if !boilerplateFound {
			t.Fatalf("blank starter %q not recognized as boilerplate-only; update the domain starter markers intentionally (findings: %v)", requiredFile, findings)
		}
	}
}
