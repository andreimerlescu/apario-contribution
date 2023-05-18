package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/tealeg/xlsx"
)

func loadCsv(ctx context.Context, filename string, callback CallbackFunc) error {
	file, openErr := os.Open(filename)
	if openErr != nil {
		log.Printf("cant open the file because of err: %v", openErr)
		return openErr
	}
	defer func(file *os.File) {
		closeErr := file.Close()
		if closeErr != nil {
			log.Fatalf("failed to close the file %v caused error %v", filename, closeErr)
		}
	}(file)
	bufferedReader := bufio.NewReaderSize(file, BufferSize)
	reader := csv.NewReader(bufferedReader)
	reader.Comma = '|'
	reader.FieldsPerRecord = -1
	headerFields, bufferReadErr := reader.Read()
	if bufferReadErr != nil {
		log.Printf("cant read the csv buffer because of err: %v", bufferReadErr)
		return bufferReadErr
	}
	log.Printf("headerFields = %v", strings.Join(headerFields, ","))
	row := make(chan []Column, SemaLimiter)
	totalRows, rowWg := atomic.Uint32{}, sync.WaitGroup{}
	done := make(chan struct{})
	go ReceiveRows(ctx, row, filename, callback, done)
	for {
		rowFields, readerErr := reader.Read()
		if readerErr != nil {
			log.Printf("skipping row due to error %v with data %v", readerErr, rowFields)
			break
		}
		totalRows.Add(1)
		rowWg.Add(1)
		go ProcessRow(headerFields, rowFields, &rowWg, row)
	}

	rowWg.Wait()
	close(row)
	<-done
	log.Printf("totalRows = %d", totalRows.Load())
	return nil
}

func loadXlsx(ctx context.Context, filename string, callback CallbackFunc) error {
	file, err := xlsx.OpenFile(filename)
	if err != nil {
		log.Printf("cant open the file because of err: %v", err)
		return err
	}
	sheet := file.Sheets[0]
	headerFields := make([]string, 0, len(sheet.Rows[0].Cells))
	for _, cell := range sheet.Rows[0].Cells {
		if len(cell.String()) > 0 {
			headerFields = append(headerFields, cell.String())
		}
	}
	log.Printf("headerFields = %v", strings.Join(headerFields, ","))
	row := make(chan []Column, SemaLimiter)
	totalRows, rowWg := atomic.Uint32{}, sync.WaitGroup{}
	done := make(chan struct{})
	go ReceiveRows(ctx, row, filename, callback, done)
	for _, sheetRow := range sheet.Rows[1:] {
		rowFields := make([]string, 0, len(sheetRow.Cells))
		for _, cell := range sheetRow.Cells {
			rowFields = append(rowFields, cell.String())
		}
		totalRows.Add(1)
		rowWg.Add(1)
		go ProcessRow(headerFields, rowFields, &rowWg, row)
	}
	rowWg.Wait()
	close(row)
	<-done
	log.Printf("totalRows = %d", totalRows.Load())
	return nil
}

func ProcessRow(headerFields []string, rowFields []string, rowWg *sync.WaitGroup, row chan []Column) {
	defer rowWg.Done()
	var d = map[string]string{}
	if len(headerFields) != len(rowFields) {
		if len(headerFields) < len(rowFields) {
			for i, r := range rowFields {
				if i >= len(headerFields) || len(r) == 0 {
					continue
				}
				d[headerFields[i]] = r
			}
		} else {
			for i, h := range headerFields {
				if i >= len(rowFields) || len(h) == 0 {
					continue
				}
				d[h] = rowFields[i]
			}
		}
	}
	var rowData = []Column{}
	if len(d) > 0 {
		for h, v := range d {
			rowData = append(rowData, Column{Header: h, Value: v})
		}
	} else {
		for i := 0; i < len(rowFields); i++ {
			value := rowFields[i]
			if i == 0 && len(value) == 0 {
				return
			}
			if len(headerFields) < i {
				log.Printf("skipping rowField %v due to headerFields not matching up properly", rowFields[i])
				continue
			}
			rowData = append(rowData, Column{headerFields[i], value})
		}
	}
	row <- rowData
}

func ReceiveRows(ctx context.Context, row chan []Column, filename string, callback CallbackFunc, done chan struct{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case populatedRow, ok := <-row:
			if !ok {
				done <- struct{}{}
				return
			}
			ctx := context.WithValue(ctx, CtxKey("csv_file"), filename)
			callbackErr := callback(ctx, populatedRow)
			if callbackErr != nil {
				log.Printf("failed to insert row %v with error %v", populatedRow, callbackErr)
			}
		}
	}
}
