package job

import (
	"encoding/csv"
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

	sheet := f.GetSheetName(0)
	rowsIter, err := f.Rows(sheet)
	if err != nil {
		return ProcessResult{}, fmt.Errorf("failed get rows: %w", err)
	}

	var rows [][]string
	var header []string
	first := true
	for rowsIter.Next() {
		row, _ := rowsIter.Columns()
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

	return processRowsWithHeader(rows, header, x.workers, x.buffer), nil
}

func processRowsWithHeader(rows [][]string, header []string, workers, buffer int) ProcessResult {
	result := ProcessResult{
		Total:          len(rows),
		SuccessRecords: []ContributorRecord{},
		FailRecords:    []FailRecord{},
	}

	colIdx := map[string]int{}
	for i, col := range header {
		colName := strings.ToLower(strings.TrimSpace(col))
		colIdx[colName] = i
	}

	requiredCols := []ColumnName{GithubID, GithubUsername, PrivacyMode, Email}
	for _, c := range requiredCols {
		if _, ok := colIdx[strings.ToLower(c.String())]; !ok {
			return ProcessResult{
				Total: len(rows),
				FailRecords: []FailRecord{
					{
						RecordNumber: 0,
						Reason:       fmt.Sprintf("missing required column: %s", c.String()),
						LastError:    "header missing required column",
						ErrType:      ErrTypeValidation,
					},
				},
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
				if len(r.data) < len(header) {
					failChan <- FailRecord{
						RecordNumber: r.rowNumber,
						Reason:       "row length less than header",
						RawData:      cloneRow(r.data),
						LastError:    "row too short",
						ErrType:      ErrTypeValidation,
					}
					continue
				}

				idStr := safeGet(r.data, colIdx[strings.ToLower(GithubID.String())])
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
					GithubUsername: safeGet(r.data, colIdx[strings.ToLower(GithubUsername.String())]),
					DisplayName:    safeGet(r.data, colIdx[strings.ToLower(DisplayName.String())]),
					ProfileImage:   safeGet(r.data, colIdx[strings.ToLower(ProfileImage.String())]),
					Bio:            safeGet(r.data, colIdx[strings.ToLower(Bio.String())]),
					PrivacyMode:    safeGet(r.data, colIdx[strings.ToLower(PrivacyMode.String())]),
					Email:          safeGet(r.data, colIdx[strings.ToLower(Email.String())]),
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

func cloneRow(r []string) []string {
	cp := make([]string, len(r))
	copy(cp, r)
	return cp
}
