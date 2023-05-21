package main

import (
	"context"
	"sync"

	"go-vue-sql-apario/sema"
)

var ch_ImportedRow = make(chan ResultData, SemaLimiter)
var TempDirs sync.Map

var db_documents sync.Map
var db_pages sync.Map

var ch_Done = make(chan struct{}, 1)
var PerformingWork = sync.WaitGroup{}

var OCRSemaphore = sema.New(17)
var DownloadSemaphore = sema.New(3)

var ch_ExtractText = make(chan ResultData, SemaLimiter)
var ch_ExtractPages = make(chan ResultData, SemaLimiter)
var ch_GeneratePng = make(chan PendingPage, SemaLimiter)
var ch_GenerateLight = make(chan PendingPage, SemaLimiter)
var ch_GenerateDark = make(chan PendingPage, SemaLimiter)
var ch_ConvertToJpg = make(chan PendingPage, SemaLimiter)
var ch_PerformOcr = make(chan PendingPage, SemaLimiter)
var ch_CompletedPage = make(chan PendingPage, SemaLimiter)

type ResultData struct {
	Identifier        string            `json:"identifier"`
	URL               string            `json:"url"`
	DataDir           string            `json:"data_dir"`
	PDFPath           string            `json:"pdf_path"`
	PDFChecksum       string            `json:"pdf_checksum"`
	OCRTextPath       string            `json:"ocr_text_path"`
	ExtractedTextPath string            `json:"extracted_text_path"`
	RecordPath        string            `json:"record_path"`
	TotalPages        int               `json:"total_pages"`
	Metadata          map[string]string `json:"metadata"`
}

type PendingPage struct {
	Identifier       string `json:"identifier"`
	RecordIdentifier string `json:"record_identifier"`
	PageNumber       int    `json:"page_number"`
	PDFPath          string `json:"pdf_path"`
	PagesDir         string `json:"pages_dir"`
	OCRTextPath      string `json:"ocr_text_path"`
	Light            Images `json:"light"`
	Dark             Images `json:"dark"`
}

type Images struct {
	Original string `json:"original"`
	Large    string `json:"large"`
	Medium   string `json:"medium"`
	Small    string `json:"small"`
	Social   string `json:"social"`
}

type Column struct {
	Header string
	Value  string
}

type Qbit struct {
	seq   [3]byte
	count int
}

type CtxKey string
type CallbackFunc func(ctx context.Context, row []Column) error
