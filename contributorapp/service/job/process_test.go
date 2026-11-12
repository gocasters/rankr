package job

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

// process_test.go — Issue #274 (TD-23), scenarios C1–C13 from test-plan-274.html.
//
// Characterization tests for the CSV parser (CSVProcess.Process and the shared
// processRowsWithHeader). They lock the CURRENT behaviour of the import path:
// what maps to a success record, what aborts the whole file, what becomes a
// per-row failure, and how header validation behaves. The "whole file fails
// on a malformed row" rule (C6/C7) is the agreed behaviour. Known BUGS are not
// locked as expected values: they are TestKnownBug_* tests (see
// knownbug_test.go) that assert the correct behaviour and SKIP while the bug
// is present.

const fullHeader = "github_id,github_username,display_name,profile_image,bio,privacy_mode,email\n"

func newTestCSVProcess(workers, buffer int) CSVProcess {
	return CSVProcess{workers: workers, buffer: buffer}
}

func writeTempCSV(t *testing.T, content string) *os.File {
	t.Helper()
	p := filepath.Join(t.TempDir(), "input.csv")
	require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	f, err := os.Open(p)
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	return f
}

// recordByRow finds a success record by its 1-based data row number, so tests
// stay order-independent (workers process rows concurrently).
func recordByRow(t *testing.T, res ProcessResult, row int) ContributorRecord {
	t.Helper()
	for _, r := range res.SuccessRecords {
		if r.RowNumber == row {
			return r
		}
	}
	t.Fatalf("no success record with RowNumber=%d (got %d records)", row, len(res.SuccessRecords))
	return ContributorRecord{}
}

// C1 (P0): a well-formed file with the full 7-column header parses every row
// and maps each column to its field. Empty optional cells must stay "".
func TestCSV_ValidFile_AllSuccess(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "valid.csv"))
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	res, err := newTestCSVProcess(4, 2).Process(f)
	require.NoError(t, err)

	assert.Equal(t, 3, res.Total)
	assert.Equal(t, 3, res.Success)
	assert.Equal(t, 0, res.Fail)
	require.Len(t, res.SuccessRecords, 3)
	assert.Empty(t, res.FailRecords)

	// row 1: every field present
	r1 := recordByRow(t, res, 1)
	assert.Equal(t, int64(101), r1.GithubID)
	assert.Equal(t, "alice01", r1.GithubUsername)
	assert.Equal(t, "Alice Smith", r1.DisplayName)
	assert.Equal(t, "https://img.example/alice.png", r1.ProfileImage)
	assert.Equal(t, "Bio of alice", r1.Bio)
	assert.Equal(t, "real", r1.PrivacyMode)
	assert.Equal(t, "alice@example.com", r1.Email)
	assert.False(t, r1.CreatedAt.IsZero())

	// row 2: quoted display name + empty optional cells stay empty
	r2 := recordByRow(t, res, 2)
	assert.Equal(t, int64(102), r2.GithubID)
	assert.Equal(t, "bob002", r2.GithubUsername)
	assert.Equal(t, "Doe, John", r2.DisplayName)
	assert.Empty(t, r2.ProfileImage)
	assert.Empty(t, r2.Bio)
	assert.Equal(t, "anonymous", r2.PrivacyMode)
	assert.Equal(t, "bob@example.com", r2.Email)

	// row 3: empty profile_image cell maps to ""
	r3 := recordByRow(t, res, 3)
	assert.Equal(t, int64(103), r3.GithubID)
	assert.Equal(t, "carol03", r3.GithubUsername)
	assert.Equal(t, "carol03", r3.DisplayName)
	assert.Empty(t, r3.ProfileImage)
	assert.Equal(t, "Likes Go", r3.Bio)
	assert.Equal(t, "real", r3.PrivacyMode)
	assert.Equal(t, "carol@example.com", r3.Email)
}

// C13 KNOWN BUG #352 (bugParserColumns): when an OPTIONAL column is absent from the
// header, colIdx has no entry for it, the map lookup returns 0, and the field
// receives the value of the FIRST column (github_id). Any file that omits
// display_name/profile_image/bio (legal: only 4 columns are required) imports
// every row with these fields corrupted. Expected: absent column => "".
func TestKnownBug_CSV_MissingOptionalColumn_FallsBackToGithubID(t *testing.T) {
	f := writeTempCSV(t,
		"github_id,github_username,privacy_mode,email\n"+
			"12345,alice,real,a@b.com\n")

	res, err := newTestCSVProcess(2, 2).Process(f)
	require.NoError(t, err)
	require.Equal(t, 1, res.Success)
	require.Len(t, res.SuccessRecords, 1)

	r := res.SuccessRecords[0]
	// required columns are correct
	assert.Equal(t, int64(12345), r.GithubID)
	assert.Equal(t, "alice", r.GithubUsername)
	assert.Equal(t, "real", r.PrivacyMode)
	assert.Equal(t, "a@b.com", r.Email)

	knownBug(t, bugParserColumns, func(a *assert.Assertions) {
		a.Empty(r.DisplayName, "absent display_name column must give \"\", not the github_id value")
		a.Empty(r.ProfileImage, "absent profile_image column must give \"\", not the github_id value")
		a.Empty(r.Bio, "absent bio column must give \"\", not the github_id value")
	})
}

// C2 (P1): header names are matched case-insensitively and trimmed of
// surrounding whitespace (V6) — required-column checks and column lookup both
// go through ToLower(TrimSpace(...)).
func TestCSV_HeaderCaseAndWhitespaceInsensitive(t *testing.T) {
	f := writeTempCSV(t,
		" GitHub_ID ,github_Username,DISPLAY_NAME,profile_image,BIO,Privacy_Mode,EMAIL \n"+
			"101,alice01,Alice,,,real,alice@example.com\n")

	res, err := newTestCSVProcess(2, 2).Process(f)
	require.NoError(t, err)

	require.Equal(t, 1, res.Success, "weird-cased header must still be accepted")
	assert.Equal(t, 0, res.Fail)
	r := res.SuccessRecords[0]
	assert.Equal(t, int64(101), r.GithubID)
	assert.Equal(t, "alice01", r.GithubUsername)
	assert.Equal(t, "real", r.PrivacyMode)
	assert.Equal(t, "alice@example.com", r.Email)
}

// C3 (P0): removing any single required column aborts header validation with
// exactly one FailRecord: RecordNumber=0 (file-level, not row-level), reason
// "missing required column: X", ErrType=validation.
// (The Fail counter on this path is a known bug — see
// TestKnownBug_CSV_MissingRequiredColumn_FailCounter.)
func TestCSV_MissingRequiredColumns(t *testing.T) {
	required := []string{"github_id", "github_username", "privacy_mode", "email"}

	for _, missing := range required {
		t.Run(missing, func(t *testing.T) {
			// build the full header minus one column
			cols := strings.Split(strings.TrimSuffix(fullHeader, "\n"), ",")
			var kept []string
			for _, c := range cols {
				if c != missing {
					kept = append(kept, c)
				}
			}

			// one valid data row against the reduced header
			values := map[string]string{
				"github_id":       "101",
				"github_username": "alice01",
				"display_name":    "Alice",
				"profile_image":   "",
				"bio":             "",
				"privacy_mode":    "real",
				"email":           "alice@example.com",
			}
			var row []string
			for _, c := range kept {
				row = append(row, values[c])
			}

			content := strings.Join(kept, ",") + "\n" + strings.Join(row, ",") + "\n"
			f := writeTempCSV(t, content)

			res, err := newTestCSVProcess(2, 2).Process(f)
			require.NoError(t, err, "missing column is reported per-file, not as an error")

			require.Len(t, res.FailRecords, 1, "exactly one file-level failure")
			fr := res.FailRecords[0]
			assert.Equal(t, 0, fr.RecordNumber, "file-level failure has RecordNumber 0")
			assert.Equal(t, fmt.Sprintf("missing required column: %s", missing), fr.Reason)
			assert.Equal(t, "header missing required column", fr.LastError)
			assert.Equal(t, ErrTypeValidation, fr.ErrType)

			assert.Empty(t, res.SuccessRecords)
			assert.Equal(t, 0, res.Success)
		})
	}
}

// KNOWN BUG #352 (bugParserColumns): the missing-required-column early return
// leaves ProcessResult.Fail at 0 although FailRecords has one entry (the
// normal path sets Fail = len(FailRecords)). Expected: the counters agree.
func TestKnownBug_CSV_MissingRequiredColumn_FailCounter(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "missing_col.csv"))
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	res, err := newTestCSVProcess(2, 2).Process(f)
	require.NoError(t, err)
	require.Len(t, res.FailRecords, 1)

	knownBug(t, bugParserColumns, func(a *assert.Assertions) {
		a.Equal(len(res.FailRecords), res.Fail, "ProcessResult.Fail must equal len(FailRecords)")
	})
}

// C3b (P0): the checked-in fixture missing github_id produces the same
// file-level failure — covers the testdata/ path used by the other tests too.
func TestCSV_MissingRequiredColumn_Fixture(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "missing_col.csv"))
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	res, err := newTestCSVProcess(2, 2).Process(f)
	require.NoError(t, err)

	require.Len(t, res.FailRecords, 1)
	assert.Equal(t, "missing required column: github_id", res.FailRecords[0].Reason)
	assert.Equal(t, 0, res.FailRecords[0].RecordNumber)
	assert.Empty(t, res.SuccessRecords)
}

// C4 (P0): one unparseable github_id fails ONLY that row (V5). The rest of
// the file still imports, the failing row keeps its raw data for retry, and
// the error type is validation.
func TestCSV_InvalidGithubID_PerRow(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "bad_id.csv"))
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	res, err := newTestCSVProcess(4, 2).Process(f)
	require.NoError(t, err)

	assert.Equal(t, 2, res.Total)
	assert.Equal(t, 1, res.Success)
	assert.Equal(t, 1, res.Fail)
	require.Len(t, res.FailRecords, 1)

	fr := res.FailRecords[0]
	assert.Equal(t, 2, fr.RecordNumber, "the bad row is data row 2")
	assert.Contains(t, fr.Reason, "invalid github id")
	assert.Equal(t, ErrTypeValidation, fr.ErrType)
	assert.NotEmpty(t, fr.LastError)
	// raw data preserved for a possible retry/inspection
	require.Len(t, fr.RawData, 7)
	assert.Equal(t, "notanumber", fr.RawData[0])
	assert.Equal(t, "bob002", fr.RawData[1])

	// the good row before the bad one is unaffected
	r := recordByRow(t, res, 1)
	assert.Equal(t, int64(101), r.GithubID)
}

// C5 (P1): a github_id outside int64 fails only its own row.
func TestCSV_OverflowGithubID_PerRow(t *testing.T) {
	f := writeTempCSV(t, fullHeader+
		"99999999999999999999,ovf001,Overflow,,,real,ovf@example.com\n"+
		"101,ok001,Ok,,,real,ok@example.com\n")

	res, err := newTestCSVProcess(2, 2).Process(f)
	require.NoError(t, err)

	assert.Equal(t, 2, res.Total)
	assert.Equal(t, 1, res.Success)
	assert.Equal(t, 1, res.Fail)
	require.Len(t, res.FailRecords, 1)
	assert.Equal(t, 1, res.FailRecords[0].RecordNumber)
	assert.Contains(t, res.FailRecords[0].Reason, "invalid github id")
	assert.Equal(t, ErrTypeValidation, res.FailRecords[0].ErrType)
	assert.Equal(t, int64(101), recordByRow(t, res, 2).GithubID)
}

// KNOWN BUG #345 (bugImportValidation): the parser accepts github_id <= 0 (ParseInt
// has no lower bound) although the create/update API rejects negative ids.
// Expected: a non-positive github_id fails its row as a validation error.
func TestKnownBug_CSV_NegativeGithubID_Accepted(t *testing.T) {
	f := writeTempCSV(t, fullHeader+
		"-5,neg001,Negative,,,real,neg@example.com\n"+
		"101,ok001,Ok,,,real,ok@example.com\n")

	res, err := newTestCSVProcess(2, 2).Process(f)
	require.NoError(t, err)

	knownBug(t, bugImportValidation, func(a *assert.Assertions) {
		a.Equal(1, res.Success, "only the valid row may succeed")
		if a.Len(res.FailRecords, 1, "the negative github_id row must fail") {
			a.Equal(1, res.FailRecords[0].RecordNumber)
			a.Equal(ErrTypeValidation, res.FailRecords[0].ErrType)
		}
	})
}

// C6 (P0): V3 lock (Q3-a decision) — a row shorter than the header aborts the
// WHOLE file with "failed read row" (encoding/csv enforces FieldsPerRecord
// from the header). This is the current "all-or-nothing" structural behaviour;
// the per-row length check inside the workers is dead code for CSV input.
func TestCSV_ShortRow_FailsWholeFile(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "short_row.csv"))
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	res, err := newTestCSVProcess(4, 2).Process(f)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed read row")

	// all-or-nothing: nothing at all is processed
	assert.Equal(t, ProcessResult{}, res, "zero result on structural error")
}

// C7 (P1): a row LONGER than the header fails the whole file the same way
// (csv.ErrFieldCount applies to both directions).
func TestCSV_LongRow_FailsWholeFile(t *testing.T) {
	f := writeTempCSV(t, fullHeader+
		"101,alice01,Alice,,,real,alice@example.com,EXTRA\n")

	res, err := newTestCSVProcess(2, 2).Process(f)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed read row")
	assert.Equal(t, ProcessResult{}, res)
}

// C8 (P1): a zero-byte file fails at the header read with EOF.
func TestCSV_EmptyFile_Error(t *testing.T) {
	f := writeTempCSV(t, "")

	res, err := newTestCSVProcess(2, 2).Process(f)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed read header")
	assert.Contains(t, err.Error(), "EOF")
	assert.Equal(t, ProcessResult{}, res)
}

// C9 (P1): header-only file is valid: zero rows, no error, and both record
// slices are non-nil and empty (callers iterate them without nil checks).
func TestCSV_HeaderOnly_ZeroRows(t *testing.T) {
	f := writeTempCSV(t, fullHeader)

	res, err := newTestCSVProcess(2, 2).Process(f)
	require.NoError(t, err)

	assert.Equal(t, 0, res.Total)
	assert.Equal(t, 0, res.Success)
	assert.Equal(t, 0, res.Fail)
	require.NotNil(t, res.SuccessRecords)
	require.NotNil(t, res.FailRecords)
	assert.Empty(t, res.SuccessRecords)
	assert.Empty(t, res.FailRecords)
}

// C10 (P2): quoted fields containing the delimiter parse correctly —
// encoding/csv standard behaviour the import path relies on.
func TestCSV_QuotedFields(t *testing.T) {
	f := writeTempCSV(t, fullHeader+
		`101,alice01,"Doe, John",,"Line1
Line2",real,alice@example.com`+"\n")

	res, err := newTestCSVProcess(2, 2).Process(f)
	require.NoError(t, err)
	require.Equal(t, 1, res.Success)

	r := res.SuccessRecords[0]
	assert.Equal(t, "Doe, John", r.DisplayName, "quoted comma must stay inside one field")
	assert.Contains(t, r.Bio, "\n", "quoted newline must stay inside one field")
}

// C11 (P0): 500 rows through 8 workers — success+fail must equal total and
// every data row number 1..500 must appear exactly once (order is NOT
// guaranteed across workers, so this asserts on the set, not the slice).
// Run under -race by make check this also proves the worker pool is race-free.
func TestCSV_ConcurrentManyRows_RaceSafe(t *testing.T) {
	var b strings.Builder
	b.WriteString(fullHeader)
	for i := 1; i <= 500; i++ {
		fmt.Fprintf(&b, "%d,user%03d,Name%d,,,real,user%d@example.com\n", i, i, i, i)
	}
	f := writeTempCSV(t, b.String())

	res, err := newTestCSVProcess(8, 4).Process(f)
	require.NoError(t, err)

	assert.Equal(t, 500, res.Total)
	assert.Equal(t, 500, res.Success)
	assert.Equal(t, 0, res.Fail)

	seen := make(map[int]int, 500)
	for _, r := range res.SuccessRecords {
		seen[r.RowNumber]++
	}
	require.Len(t, seen, 500, "every row number 1..500 must be present")
	for i := 1; i <= 500; i++ {
		assert.Equal(t, 1, seen[i], "row %d must appear exactly once", i)
	}
}

// C12 (P1): row numbers count DATA rows from 1 (header excluded), so a user
// can trace a failure back to the exact line of their file.
func TestCSV_RowNumbers_MatchInput(t *testing.T) {
	f := writeTempCSV(t, fullHeader+
		"101,row001,First,,,real,a@example.com\n"+
		"bad,row002,Second,,,real,b@example.com\n"+
		"103,row003,Third,,,real,c@example.com\n")

	res, err := newTestCSVProcess(4, 2).Process(f)
	require.NoError(t, err)

	require.Len(t, res.FailRecords, 1)
	assert.Equal(t, 2, res.FailRecords[0].RecordNumber, "bad row is the 2nd data row (file line 3)")

	// good rows keep their input order numbers: 1 and 3 (not 2 — that failed)
	rowNumbers := map[int]bool{}
	for _, r := range res.SuccessRecords {
		rowNumbers[r.RowNumber] = true
	}
	assert.Equal(t, map[int]bool{1: true, 3: true}, rowNumbers)
}

// KNOWN BUG #352 (bugParserColumns): excelize's Rows().Columns() drops trailing
// empty cells, so an XLSX row whose LAST optional cells are empty is shorter
// than the header and fails with "row length less than header". The same data
// as CSV imports fine. Expected: trailing empty cells mean "" values.
func TestKnownBug_XLSX_TrailingEmptyCells_RowRejected(t *testing.T) {
	x := excelize.NewFile()
	sheet := x.GetSheetName(0)
	require.NoError(t, x.SetSheetRow(sheet, "A1", &[]string{"github_id", "github_username", "privacy_mode", "email", "bio"}))
	require.NoError(t, x.SetSheetRow(sheet, "A2", &[]string{"101", "alice01", "real", "alice@example.com", ""}))
	require.NoError(t, x.SetSheetRow(sheet, "A3", &[]string{"102", "bob002", "real", "bob@example.com", "has a bio"}))
	p := filepath.Join(t.TempDir(), "input.xlsx")
	require.NoError(t, x.SaveAs(p))
	require.NoError(t, x.Close())

	f, err := os.Open(p)
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })

	res, err := XLSXProcess{workers: 2, buffer: 2}.Process(f)
	require.NoError(t, err)
	require.Equal(t, 2, res.Total)

	knownBug(t, bugParserColumns, func(a *assert.Assertions) {
		a.Equal(2, res.Success, "a row with an empty last cell must import like the CSV equivalent")
		a.Empty(res.FailRecords, "no row may fail only because its trailing cells are empty")
	})
}

// KNOWN BUG #352 (bugParserColumns): with workers <= 0 (e.g. process_count missing
// from config) no worker reads the rows: Process returns Success=0, Fail=0,
// no error, and the job ends "failed" with every row silently lost.
// Expected: either an error, or every row accounted for.
// (buffer >= rows here so the producer goroutine does not block forever.)
func TestKnownBug_CSV_ZeroWorkers_DropsRowsSilently(t *testing.T) {
	f := writeTempCSV(t, fullHeader+
		"101,alice01,Alice,,,real,alice@example.com\n"+
		"102,bob002,Bob,,,real,bob@example.com\n")

	res, err := newTestCSVProcess(0, 10).Process(f)

	knownBug(t, bugParserColumns, func(a *assert.Assertions) {
		if err == nil {
			a.Equal(res.Total, res.Success+res.Fail, "every row must end up as success or fail, or Process must return an error")
		}
	})
}
