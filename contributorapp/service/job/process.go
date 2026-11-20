package job

import (
	"archive/zip"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"github.com/xuri/excelize/v2"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ProcessResult struct {
	Total          int
	Success        int
	Fail           int
	SuccessRecords []ContributorRecord
	FailRecords    []FailRecord
}

type rowInfo struct {
	rowNumber int
	data      []string
}

type CSVProcess struct {
	workers int
	buffer  int
}

type XLSXProcess struct {
	workers int
	buffer  int
}

func (c CSVProcess) Process(file *os.File) (ProcessResult, error) {
	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return ProcessResult{}, fmt.Errorf("failed read header: %w", err)
	}

	var rows [][]string
	for {
		rec, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return ProcessResult{}, fmt.Errorf("failed read row: %w", err)
		}
		if len(rec) == 0 {
			continue
		}
		rows = append(rows, rec)
	}

	return processRowsWithHeader(rows, header, c.workers, c.buffer), nil
}

func (x XLSXProcess) Process(file *os.File) (ProcessResult, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return ProcessResult{}, fmt.Errorf("failed open xlsx: %w", err)
	}
	defer f.Close()

	// excelize's row iterator reports no error when a worksheet's XML ends
	// early, so a truncated sheet would import its partial last row (padding
	// hides the missing cells) and silently drop the rest. Check the XML is
	// well-formed ourselves before trusting the iterator.
	if err := validateWorksheetXML(file); err != nil {
		return ProcessResult{}, fmt.Errorf("failed read row: %w", err)
	}

	sheet := f.GetSheetName(0)
	rowsIter, err := f.Rows(sheet)
	if err != nil {
		return ProcessResult{}, fmt.Errorf("failed get rows: %w", err)
	}
	defer func() { _ = rowsIter.Close() }()

	var rows [][]string
	var header []string
	first := true
	for rowsIter.Next() {
		row, err := rowsIter.Columns()
		if err != nil {
			// A decode failure must abort the whole file, exactly like the CSV
			// path ("failed read row"). Discarding this error used to turn a
			// partial row into a silent import with empty trailing fields (#352).
			return ProcessResult{}, fmt.Errorf("failed read row: %w", err)
		}
		if first {
			header = row
			first = false
			continue
		}
		if len(row) == 0 {
			continue
		}
		rows = append(rows, row)
	}
	if err := rowsIter.Error(); err != nil {
		return ProcessResult{}, fmt.Errorf("failed read row: %w", err)
	}

	return processRowsWithHeader(rows, header, x.workers, x.buffer), nil
}

// validateWorksheetXML reports an error if any worksheet part of the XLSX
// file is not well-formed XML (e.g. truncated). It uses ReadAt only, so the
// file offset is left untouched.
func validateWorksheetXML(file *os.File) error {
	fi, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat xlsx: %w", err)
	}
	zr, err := zip.NewReader(file, fi.Size())
	if err != nil {
		return fmt.Errorf("open xlsx zip: %w", err)
	}
	for _, zf := range zr.File {
		if !strings.HasPrefix(zf.Name, "xl/worksheets/") || !strings.HasSuffix(zf.Name, ".xml") {
			continue
		}
		if err := checkWellFormedXML(zf); err != nil {
			return fmt.Errorf("malformed worksheet %s: %w", zf.Name, err)
		}
	}
	return nil
}

func checkWellFormedXML(zf *zip.File) error {
	rc, err := zf.Open()
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()
	dec := xml.NewDecoder(rc)
	for {
		if _, err := dec.Token(); err != nil {
			if err == io.EOF {
				// Token returns a syntax error, not io.EOF, if any
				// element is still open, so EOF here means well-formed.
				return nil
			}
			return err
		}
	}
}

func processRowsWithHeader(rows [][]string, header []string, workers, buffer int) ProcessResult {
	result := ProcessResult{
		Total:          len(rows),
		SuccessRecords: []ContributorRecord{},
		FailRecords:    []FailRecord{},
	}

	// #352d: never run with zero workers (e.g. process_count missing from the
	// config). No worker would read rowChan, so Process would return
	// Success=0/Fail=0 with no error and every row would be dropped silently.
	// Fall back to a single worker: slower, but no row is ever lost.
	if workers < 1 {
		workers = 1
	}

	colIdx := map[string]int{}
	for i, col := range header {
		colName := strings.ToLower(strings.TrimSpace(col))
		colIdx[colName] = i
	}

	requiredCols := []ColumnName{GithubID, GithubUsername, PrivacyMode, Email}
	for _, c := range requiredCols {
		if _, ok := colIdx[strings.ToLower(c.String())]; !ok {
			failRecords := []FailRecord{
				{
					RecordNumber: 0,
					Reason:       fmt.Sprintf("missing required column: %s", c.String()),
					LastError:    "header missing required column",
					ErrType:      ErrTypeValidation,
				},
			}
			// #352c: keep Fail in sync with FailRecords on every return path,
			// exactly like the normal path below (Fail = len(FailRecords)).
			return ProcessResult{
				Total:       len(rows),
				Fail:        len(failRecords),
				FailRecords: failRecords,
			}
		}
	}

	rowChan := make(chan rowInfo, buffer)
	successChan := make(chan ContributorRecord, buffer)
	failChan := make(chan FailRecord, buffer)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := range rowChan {
				// #352b: excelize drops trailing empty cells, so an XLSX row can
				// be shorter than the header. Pad it with "" (lossless: trailing
				// empties are exactly what was dropped). Only a row LONGER than
				// the header carries data the header cannot name -> fail it.
				data := r.data
				if len(data) > len(header) {
					failChan <- FailRecord{
						RecordNumber: r.rowNumber,
						Reason:       "row length more than header",
						RawData:      cloneRow(data),
						LastError:    "row too long",
						ErrType:      ErrTypeValidation,
					}
					continue
				}
				if len(data) < len(header) {
					padded := make([]string, len(header))
					copy(padded, data)
					data = padded
				}

				idStr := cellValue(data, colIdx, strings.ToLower(GithubID.String()))
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					failChan <- FailRecord{
						RecordNumber: r.rowNumber,
						Reason:       fmt.Sprintf("invalid github id: %v", err),
						RawData:      cloneRow(r.data),
						LastError:    err.Error(),
						ErrType:      ErrTypeValidation,
					}
					continue
				}

				cr := ContributorRecord{
					RowNumber:      r.rowNumber,
					GithubID:       id,
					GithubUsername: cellValue(data, colIdx, strings.ToLower(GithubUsername.String())),
					DisplayName:    cellValue(data, colIdx, strings.ToLower(DisplayName.String())),
					ProfileImage:   cellValue(data, colIdx, strings.ToLower(ProfileImage.String())),
					Bio:            cellValue(data, colIdx, strings.ToLower(Bio.String())),
					PrivacyMode:    cellValue(data, colIdx, strings.ToLower(PrivacyMode.String())),
					Email:          cellValue(data, colIdx, strings.ToLower(Email.String())),
					CreatedAt:      time.Now(),
				}

				successChan <- cr
			}
		}()
	}

	go func() {
		for i, r := range rows {
			rowChan <- rowInfo{
				rowNumber: i + 1,
				data:      cloneRow(r),
			}
		}
		close(rowChan)
	}()

	go func() {
		wg.Wait()
		close(successChan)
		close(failChan)
	}()

	for successChan != nil || failChan != nil {
		select {
		case s, ok := <-successChan:
			if !ok {
				successChan = nil
				continue
			}
			result.SuccessRecords = append(result.SuccessRecords, s)
		case f, ok := <-failChan:
			if !ok {
				failChan = nil
				continue
			}
			result.FailRecords = append(result.FailRecords, f)
		}
	}

	result.Success = len(result.SuccessRecords)
	result.Fail = len(result.FailRecords)
	return result
}

func safeGet(r []string, idx int) string {
	if idx < 0 || idx >= len(r) {
		return ""
	}
	return r[idx]
}

// cellValue returns the value of the named column in row, or "" when the
// column is absent from the header. A missing colIdx key must not be confused
// with index 0: the map zero value would otherwise read the first column
// (github_id) into every absent optional field (#352a).
func cellValue(row []string, colIdx map[string]int, name string) string {
	idx, ok := colIdx[name]
	if !ok {
		return ""
	}
	return safeGet(row, idx)
}

func cloneRow(r []string) []string {
	cp := make([]string, len(r))
	copy(cp, r)
	return cp
}
