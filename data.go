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

var (
	b_sem_tesseract  = sema.New(*flag_b_sem_tesseract)
	b_sem_download   = sema.New(*flag_b_sem_download)
	b_sem_pdfcpu     = sema.New(*flag_b_sem_pdfcpu)
	b_sem_gs         = sema.New(*flag_b_sem_gs)
	b_sem_pdftotext  = sema.New(*flag_b_sem_pdftotext)
	b_sem_convert    = sema.New(*flag_b_sem_convert)
	b_sem_pdftoppm   = sema.New(*flag_b_sem_pdftoppm)
	g_sem_png2jpg    = sema.New(*flag_g_sem_png2jpg)
	g_sem_resize     = sema.New(*flag_g_sem_resize)
	g_sem_shafile    = sema.New(*flag_g_sem_shafile)
	g_sema_watermark = sema.New(*flag_g_sem_watermark)
	g_sem_darkimage  = sema.New(*flag_g_sem_darkimage)
	g_sem_filedata   = sema.New(*flag_g_sem_filedata)
	g_sem_shastring  = sema.New(*flag_g_sem_shastring)
	g_sem_wjsonfile  = sema.New(*flag_g_sem_wjsonfile)
)

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
