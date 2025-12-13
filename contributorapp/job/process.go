package job

import (
	"encoding/csv"
	"fmt"
	"github.com/xuri/excelize/v2"
	"os"
	"strconv"
	"strings"
	"sync"
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

func newProcessResult() ProcessResult {
	return ProcessResult{
		Total:          0,
		Success:        0,
		SuccessRecords: make([]ContributorRecord, 0),
		FailRecords:    make([]FailRecord, 0),
	}
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

	colIdx := map[string]int{}
	for i, col := range header {
		colIdx[strings.ToLower(col)] = i
	}

	var rows [][]string
	for {
		rec, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return ProcessResult{}, fmt.Errorf("failed read row: %w", err)
		}
		rows = append(rows, rec)
	}

	return processRows(rows, colIdx, c.workers, c.buffer), nil
}

func (x XLSXProcess) Process(file *os.File) (ProcessResult, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return ProcessResult{}, fmt.Errorf("failed open xlsx: %w", err)
	}
	sheet := f.GetSheetName(0)
	rows, err := f.Rows(sheet)
	if err != nil {
		return ProcessResult{}, fmt.Errorf("failed get rows: %w", err)
	}

	headerRow, err := rows.Columns()
	if err != nil {
		return ProcessResult{}, fmt.Errorf("failed read header: %w", err)
	}

	colIdx := map[string]int{}
	for i, col := range headerRow {
		colIdx[strings.ToLower(col)] = i
	}

	var dataRows [][]string
	for rows.Next() {
		row, _ := rows.Columns()
		dataRows = append(dataRows, row)
	}

	return processRows(dataRows, colIdx, x.workers, x.buffer), nil
}

func processRows(rows [][]string, colIdx map[string]int, workers, buffer int) ProcessResult {
	result := ProcessResult{
		Total:          len(rows),
		SuccessRecords: make([]ContributorRecord, 0),
		FailRecords:    make([]FailRecord, 0),
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
				id, err := strconv.Atoi(r.data[colIdx[GithubID.String()]])
				if err != nil {
					failChan <- FailRecord{RecordNumber: r.rowNumber, Reason: fmt.Sprintf("invalid github id: %v", err), RawData: r.data}
					continue
				}
				var cr ContributorRecord
				setContributorRecord(id, &cr, r.data, colIdx)
				cr.RowNumber = r.rowNumber
				successChan <- cr
			}
		}()
	}

	go func() {
		for i, r := range rows {
			rowChan <- rowInfo{rowNumber: i + 1, data: r}
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

func setContributorRecord(id int, c *ContributorRecord, r []string, columnIdx map[string]int) {
	c.GithubID = int64(id)
	c.GithubUsername = safeGet(r, columnIdx[GithubUsername.String()])
	c.DisplayName = safeGet(r, columnIdx[DisplayName.String()])
	c.ProfileImage = safeGet(r, columnIdx[ProfileImage.String()])
	c.Bio = safeGet(r, columnIdx[Bio.String()])
	c.PrivacyMode = safeGet(r, columnIdx[PrivacyMode.String()])
}

func safeGet(r []string, idx int) string {
	if idx < 0 || idx >= len(r) {
		return ""
	}
	return r[idx]
}
