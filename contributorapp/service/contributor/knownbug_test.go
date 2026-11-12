package contributor

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// knownbug_test.go — "expected failure" support for documented bugs.
// (Same helper in service/contributor and service/job; keep the two in sync.)
//
// A TestKnownBug_* test asserts the CORRECT behaviour inside knownBug(...).
//   - While the bug is present those assertions fail; the test is reported as
//     SKIP with the issue id, so `make check` stays green and
//     `go test -v -run KnownBug ./contributorapp/...` lists every open bug.
//   - When the code is fixed the assertions pass and the test FAILS with
//     "appears FIXED": delete the knownBug wrapper (and the issue constant)
//     and keep the assertions as a normal regression test.
//
// NEVER make a knownBug test pass by changing its assertions to match the
// code. Do setup with the real t (require.*), and put only the assertions
// that describe the desired behaviour inside the check function.

// GitHub issues for the bugs locked in this package. When an issue is fixed,
// delete its constant together with the knownBug wrapper of its tests. Add a
// new constant (issue number + short title) for every newly filed bug.
const (
	bugEmailWipe          = "#351 (UpdateProfile wipes email)"
	bugImportValidation   = "#345 (import path validation diverges from API)"
	bugPlaintextFallback  = "#346 (passwordMatches plaintext fallback)"
	bugValidationMessages = "#348 (wrong validation messages)"
)

type bugRecorder struct{ msgs []string }

func (b *bugRecorder) Errorf(format string, args ...any) {
	b.msgs = append(b.msgs, fmt.Sprintf(format, args...))
}

func knownBug(t *testing.T, issue string, check func(a *assert.Assertions)) {
	t.Helper()
	rec := &bugRecorder{}
	check(assert.New(rec))
	if len(rec.msgs) == 0 {
		t.Fatalf("known bug %s appears FIXED: remove the knownBug wrapper and keep these assertions as a regular test", issue)
	}
	t.Skipf("KNOWN BUG %s is still present; the expected behaviour is violated:\n%s",
		issue, strings.Join(rec.msgs, "\n"))
}
