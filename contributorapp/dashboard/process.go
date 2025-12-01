package dashboard

import (
	"encoding/csv"
	"fmt"
	"github.com/xuri/excelize/v2"
	"io"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	worker       = 4
	bufferedSize = 50
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

type CSVProcess struct{}

func (c CSVProcess) Process(file multipart.File, fileName string) (ProcessResult, error) {

	_, err := file.Seek(0, 0)
	if err != nil {
		return ProcessResult{}, fmt.Errorf("failed to go to the beginnig of the file %s: %w", fileName, err)
	}

	reader := csv.NewReader(file)

	records, rErr := reader.ReadAll()
	if rErr != nil {
		return ProcessResult{}, fmt.Errorf("failed to read file %s: %w", fileName, rErr)
	}

	if len(records) < 2 {
		return ProcessResult{}, fmt.Errorf("file is empty")
	}

	header := records[0]
	rows := records[1:]

	columnIdx := map[string]int{}
	for i, colName := range header {
		columnIdx[strings.ToLower(colName)] = i
	}

	return processRows(rows, columnIdx), nil
}

type XLSXProcess struct{}

func (x XLSXProcess) Process(file multipart.File, fileName string) (ProcessResult, error) {

	tmpFile, tmpPath, err := saveTempFile(file)
	if err != nil {
		return ProcessResult{}, fmt.Errorf("failed to create temp file: %w", err)
	}

	defer func() {
		tmpFile.Close()
		os.Remove(tmpPath)
	}()

	f, oErr := excelize.OpenReader(tmpFile)
	if oErr != nil {
		return ProcessResult{}, fmt.Errorf("failed to read xlsx file: %w", oErr)
	}

	sheet := f.GetSheetName(0)
	rows, getErr := f.GetRows(sheet)
	if getErr != nil {
		return ProcessResult{}, fmt.Errorf("failed to read rows: %w", getErr)
	}

	if len(rows) < 2 {
		return ProcessResult{}, fmt.Errorf("file is empty, file name: %s", fileName)
	}

	header := rows[0]
	dataRows := rows[1:]

	columnIndex := map[string]int{}
	for i, col := range header {
		columnIndex[strings.ToLower(col)] = i
	}

	return processRows(dataRows, columnIndex), nil
}

func saveTempFile(src multipart.File) (*os.File, string, error) {
	tmp, err := os.CreateTemp("", "upload-*.xlsx")
	if err != nil {
		return nil, "", err
	}

	tmpPath := tmp.Name()

	_, err = io.Copy(tmp, src)
	if err != nil {
		return nil, "", err
	}

	_, err = tmp.Seek(0, 0)
	if err != nil {
		return nil, "", err
	}

	return tmp, tmpPath, nil
}

func processRows(rows [][]string, colIdx map[string]int) ProcessResult {
	result := newProcessResult()

	result.Total = len(rows)

	rowChan := make(chan rowInfo, bufferedSize)
	successChan := make(chan ContributorRecord, bufferedSize)
	failChan := make(chan FailRecord, bufferedSize)

	var wg sync.WaitGroup

	for i := 1; i <= worker; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := range rowChan {
				idStr := r.data[colIdx[GithubID.String()]]
				id, err := strconv.Atoi(idStr)
				if err != nil {
					failChan <- FailRecord{
						RecordNumber: r.rowNumber,
						Reason:       fmt.Sprintf("invalid github id %s: %v", idStr, err),
						RawData:      r.data,
					}

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
			ri := rowInfo{rowNumber: i + 1, data: r}

			rowChan <- ri
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
