/*
Project Apario is the World's Truth Repository that was invented and started by Andrei Merlescu in 2020.
Copyright (C) 2023  Andrei Merlescu

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package main

import (
	"context"
	"fmt"
	"image/color"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	`github.com/andreimerlescu/configurable`

	"go-vue-sql-apario/sema"
)

const (
	c_retry_attempts     = 33
	c_identifier_charset = "ABCDEFGHKMNPQRSTUVWXYZ123456789"
	c_dir_permissions    = 0111
)

var (
	startedAt = time.Now().UTC()
	config    = configurable.New()

	// Integers
	channel_buffer_size    int = 1          // Buffered Channel's Size
	reader_buffer_bytes    int = 128 * 1024 // 128KB default buffer for reading CSV, XLSX, and PSV files into memory
	jpeg_compression_ratio     = 90         // Progressive JPEG Quality (valid options are 1-100)

	// Colors
	color_background = color.RGBA{R: 40, G: 40, B: 86, A: 255}    // navy blue
	color_text       = color.RGBA{R: 250, G: 226, B: 203, A: 255} // sky yellow

	// Strings
	dir_data_directory    string
	dir_current_directory string

	// Maps
	m_cryptonyms          = make(map[string]string)
	m_location_cities     = make(map[string]*Location)
	m_location_countries  = make(map[string]*Location)
	m_location_states     = make(map[string]*Location)
	m_used_identifiers    = make(map[string]bool)
	m_required_binaries   = make(map[string]string)
	m_language_dictionary = make(map[string]map[string]struct{})
	m_gcm_jewish          = make(GemCodeMap)
	m_gcm_english         = make(GemCodeMap)
	m_gcm_simple          = make(GemCodeMap)
	m_months              = map[string]time.Month{
		"jan": time.January, "january": time.January, "01": time.January, "1": time.January,
		"feb": time.February, "february": time.February, "02": time.February, "2": time.February,
		"mar": time.March, "march": time.March, "03": time.March, "3": time.March,
		"apr": time.April, "april": time.April, "04": time.April, "4": time.April,
		"may": time.May, "05": time.May, "5": time.May,
		"jun": time.June, "june": time.June, "06": time.June, "6": time.June,
		"jul": time.July, "july": time.July, "07": time.July, "7": time.July,
		"aug": time.August, "august": time.August, "08": time.August, "8": time.August,
		"sep": time.September, "september": time.September, "09": time.September, "9": time.September,
		"oct": time.October, "october": time.October, "10": time.October,
		"nov": time.November, "november": time.November, "11": time.November,
		"dec": time.December, "december": time.December, "12": time.December,
	}

	// Synchronization
	mu_identifier         = sync.RWMutex{}
	mu_location_countries = sync.RWMutex{}
	mu_location_states    = sync.RWMutex{}
	mu_location_cities    = sync.RWMutex{}
	wg_active_tasks       = sync.WaitGroup{}

	// Command Line Flags
	flag_s_file             = config.NewString("file", "", "CSV file of URL + Metadata")
	flag_s_directory        = config.NewString("dir", "", "Path of the directory you want the export to be generated into.")
	flag_i_sem_limiter      = config.NewInt("limit", channel_buffer_size, "Number of rows to concurrently process.")
	flag_i_buffer           = config.NewInt("buffer", reader_buffer_bytes, "Memory allocation for CSV buffer (min 168 * 1024 = 168KB)")
	flag_b_sem_tesseract    = config.NewInt("tesseract", 1, "Semaphore Limiter for `tesseract` binary.")
	flag_b_sem_download     = config.NewInt("download", 2, "Semaphore Limiter for downloading PDF files from URLs.")
	flag_b_sem_pdfcpu       = config.NewInt("pdfcpu", 17, "Semaphore Limiter for `pdfcpu` binary.")
	flag_b_sem_gs           = config.NewInt("gs", 17, "Semaphore Limiter for `gs` binary.")
	flag_b_sem_pdftotext    = config.NewInt("pdftotext", 17, "Semaphore Limiter for `pdftotext` binary.")
	flag_b_sem_convert      = config.NewInt("convert", 17, "Semaphore Limiter for `convert` binary.")
	flag_b_sem_pdftoppm     = config.NewInt("pdftoppm", 17, "Semaphore Limiter for `pdftoppm` binary.")
	flag_g_sem_png2jpg      = config.NewInt("png2jpg", 17, "Semaphore Limiter for converting PNG images to JPG.")
	flag_g_sem_resize       = config.NewInt("resize", 17, "Semaphore Limiter for resize PNG or JPG images.")
	flag_g_sem_shafile      = config.NewInt("shafile", 36, "Semaphore Limiter for calculating the SHA256 checksum of files.")
	flag_g_sem_watermark    = config.NewInt("watermark", 36, "Semaphore Limiter for adding a watermark to an image.")
	flag_g_sem_darkimage    = config.NewInt("darkimage", 36, "Semaphore Limiter for converting an image to dark mode.")
	flag_g_sem_filedata     = config.NewInt("filedata", 369, "Semaphore Limiter for writing metadata about a processed file to JSON.")
	flag_g_sem_shastring    = config.NewInt("shastring", 369, "Semaphore Limiter for calculating the SHA256 checksum of a string.")
	flag_g_sem_wjsonfile    = config.NewInt("wjsonfile", 369, "Semaphore Limiter for writing a JSON file to disk.")
	flag_g_jpg_quality      = config.NewInt("jpeg-quality", 71, "Quality percentage (as int 1-100) for compressing PNG images into JPEG files.")
	flag_g_progressive_jpeg = config.NewBool("progressive", true, "Convert compressed JPEG images into progressive images.")
	flag_g_log_file         = config.NewString("log", filepath.Join(".", "logs", fmt.Sprintf("engine-%04d-%02d-%02d-%02d-%02d-%02d.log", startedAt.Year(), startedAt.Month(), startedAt.Day(), startedAt.Hour(), startedAt.Minute(), startedAt.Second())), "File to save logs to. Default is logs/engine-YYYY-MM-DD-HH-MM-SS.log")

	// Binary Dependencies
	sl_required_binaries = []string{
		"pdfcpu",
		"gs",
		"pdftotext",
		"convert",
		"pdftoppm",
		"tesseract",
	}

	// Atomics
	a_b_dictionary_loaded = atomic.Bool{}
	a_b_gematria_loaded   = atomic.Bool{}
	a_b_locations_loaded  = atomic.Bool{}
	a_i_total_pages       = atomic.Int64{}

	// Concurrent Maps
	sm_page_directories sync.Map
	sm_documents        sync.Map
	sm_pages            sync.Map

	// Semaphores
	sem_tesseract  = sema.New(*flag_b_sem_tesseract)
	sem_download   = sema.New(*flag_b_sem_download)
	sem_pdfcpu     = sema.New(*flag_b_sem_pdfcpu)
	sem_gs         = sema.New(*flag_b_sem_gs)
	sem_pdftotext  = sema.New(*flag_b_sem_pdftotext)
	sem_convert    = sema.New(*flag_b_sem_convert)
	sem_pdftoppm   = sema.New(*flag_b_sem_pdftoppm)
	sem_png2jpg    = sema.New(*flag_g_sem_png2jpg)
	sem_resize     = sema.New(*flag_g_sem_resize)
	sem_shafile    = sema.New(*flag_g_sem_shafile)
	sema_watermark = sema.New(*flag_g_sem_watermark)
	sem_darkimage  = sema.New(*flag_g_sem_darkimage)
	sem_filedata   = sema.New(*flag_g_sem_filedata)
	sem_shastring  = sema.New(*flag_g_sem_shastring)
	sem_wjsonfile  = sema.New(*flag_g_sem_wjsonfile)

	// Channels
	ch_ImportedRow       = make(chan ResultData, channel_buffer_size)
	ch_ExtractText       = make(chan ResultData, channel_buffer_size)
	ch_ExtractPages      = make(chan ResultData, channel_buffer_size)
	ch_GeneratePng       = make(chan PendingPage, channel_buffer_size)
	ch_GenerateLight     = make(chan PendingPage, channel_buffer_size)
	ch_GenerateDark      = make(chan PendingPage, channel_buffer_size)
	ch_ConvertToJpg      = make(chan PendingPage, channel_buffer_size)
	ch_PerformOcr        = make(chan PendingPage, channel_buffer_size)
	ch_AnalyzeText       = make(chan PendingPage, channel_buffer_size)
	ch_AnalyzeCryptonyms = make(chan PendingPage, channel_buffer_size)
	ch_AnalyzeGematria   = make(chan PendingPage, channel_buffer_size)
	ch_AnalyzeLocations  = make(chan PendingPage, channel_buffer_size)
	ch_AnalyzeDictionary = make(chan PendingPage, channel_buffer_size)
	ch_CompletedPage     = make(chan PendingPage, channel_buffer_size)
	ch_CompiledDocument  = make(chan Document, channel_buffer_size)
	ch_Done              = make(chan struct{}, 1)
)

type Document struct {
	Identifier          string         `json:"identifier"`
	URL                 string         `json:"url"`
	Pages               map[int64]Page `json:"pages"`
	TotalPages          int64          `json:"total_pages"`
	CoverPageIdentifier string         `json:"cover_page_identifier"`
	Collection          Collection     `json:"collection"`
}

type Page struct {
	Identifier         string            `json:"identifier"`
	DocumentIdentifier string            `json:"document_identifier"`
	PageNumber         int64             `json:"page_number"`
	Metadata           map[string]string `json:"metadata"`
	FullTextGematria   GemScore          `json:"full_text_gematria"`
	FullText           string            `json:"full_text"`
	Locations          []*Location       `json:"locations"`
}

type Location struct {
	Continent   string  `json:"continent"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	City        string  `json:"city"`
	State       string  `json:"state"`
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
}

type Collection struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
}

type ResultData struct {
	Identifier        string            `json:"identifier"`
	URL               string            `json:"url"`
	DataDir           string            `json:"data_dir"`
	PDFPath           string            `json:"pdf_path"`
	PDFChecksum       string            `json:"pdf_checksum"`
	OCRTextPath       string            `json:"ocr_text_path"`
	ExtractedTextPath string            `json:"extracted_text_path"`
	RecordPath        string            `json:"record_path"`
	TotalPages        int64             `json:"total_pages"`
	Metadata          map[string]string `json:"metadata"`
}

type JPEG struct {
	Light Images `json:"light"`
	Dark  Images `json:"dark"`
}

type PNG struct {
	Light Images `json:"light"`
	Dark  Images `json:"dark"`
}

type PendingPage struct {
	Identifier       string              `json:"identifier"`
	RecordIdentifier string              `json:"record_identifier"`
	PageNumber       int                 `json:"page_number"`
	PDFPath          string              `json:"pdf_path"`
	PagesDir         string              `json:"pages_dir"`
	OCRTextPath      string              `json:"ocr_text_path"`
	ManifestPath     string              `json:"manifest_path"`
	Language         string              `json:"language"`
	Words            []WordResult        `json:"words"`
	Cryptonyms       []string            `json:"cryptonyms"`
	Dates            []time.Time         `json:"dates"`
	Locations        []*Location         `json:"locations"`
	Gematrias        map[string]Gematria `json:"gematrias"`
	JPEG             JPEG                `json:"jpeg"`
	PNG              PNG                 `json:"png"`
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

type GemCodeMap map[string]uint

type GemScore struct {
	Jewish  uint
	English uint
	Simple  uint
}

type Gematria struct {
	Word  string   `json:"word"`
	Score GemScore `json:"score"`
}

type WordResult struct {
	Word     string   `json:"word"`
	Language string   `json:"language"`
	Gematria Gematria `json:"gematria"`
	Quantity int      `json:"quantity"`
}
