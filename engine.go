package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const maxAttempts = 33
const charset = "ABCDEFGHKMNPQRSTUVWXYZ123456789"
const ExecutePermissions = 0111

var TotalPages = atomic.Int32{}
var SemaLimiter int = 1         // 100K default
var BufferSize int = 168 * 1024 // 128KB default
var ThumbnailQualityScore = 90
var backgroundColor = color.RGBA{R: 40, G: 40, B: 86, A: 255}
var textColor = color.RGBA{R: 250, G: 226, B: 203, A: 255}
var DataDir string
var Cryptonyms map[string]string
var IdentifierMu sync.RWMutex
var UsedIdentifiers = map[string]bool{}
var PWD string
var Binaries = make(map[string]string)
var RawBinaries = []string{
	"pdfcpu",
	"gs",
	"pdftotext",
	"convert",
	"pdftoppm",
	"tesseract",
}
var (
	flag_b_sem_tesseract *int
	flag_b_sem_download  *int
	flag_b_sem_pdfcpu    *int
	flag_b_sem_gs        *int
	flag_b_sem_pdftotext *int
	flag_b_sem_convert   *int
	flag_b_sem_pdftoppm  *int
	flag_g_sem_png2jpg   *int
	flag_g_sem_resize    *int
	flag_g_sem_shafile   *int
	flag_g_sem_watermark *int
	flag_g_sem_darkimage *int
	flag_g_sem_filedata  *int
	flag_g_sem_shastring *int
	flag_g_sem_wjsonfile *int
)

func main() {

	binaryErr := verifyBinaries(RawBinaries)
	if binaryErr != nil {
		fmt.Printf("Error: %s\n", binaryErr)
		os.Exit(1)
	}

	ex, execErr := os.Getwd()
	if execErr != nil {
		panic(execErr)
	}

	PWD = filepath.Dir(ex)

	fmt.Println("Working Dir path: " + PWD)

	startedAt := time.Now()
	fileFlag := flag.String("file", "", "CSV file of URL + Metadata")
	dirFlag := flag.String("dir", "", "Path of the directory you want the export to be generated into.")
	semaphoreLimitFlag := flag.Int("limit", SemaLimiter, "Number of rows to concurrently process.")
	fileBufferSize := flag.Int("buffer", BufferSize, "Memory allocation for CSV buffer (min 168 * 1024 = 168KB)")

	flag_b_sem_tesseract = flag.Int("tesseract", 1, "Semaphore Limiter for `tesseract` binary.")
	flag_b_sem_download = flag.Int("download", 2, "Semaphore Limiter for downloading PDF files from URLs.")
	flag_b_sem_pdfcpu = flag.Int("pdfcpu", 17, "Semaphore Limiter for `pdfcpu` binary.")
	flag_b_sem_gs = flag.Int("gs", 17, "Semaphore Limiter for `gs` binary.")
	flag_b_sem_pdftotext = flag.Int("pdftotext", 17, "Semaphore Limiter for `pdftotext` binary.")
	flag_b_sem_convert = flag.Int("convert", 17, "Semaphore Limiter for `convert` binary.")
	flag_b_sem_pdftoppm = flag.Int("pdftoppm", 17, "Semaphore Limiter for `pdftoppm` binary.")
	flag_g_sem_png2jpg = flag.Int("png2jpg", 17, "Semaphore Limiter for converting PNG images to JPG.")
	flag_g_sem_resize = flag.Int("resize", 17, "Semaphore Limiter for resize PNG or JPG images.")
	flag_g_sem_shafile = flag.Int("shafile", 36, "Semaphore Limiter for calculating the SHA256 checksum of files.")
	flag_g_sem_watermark = flag.Int("watermark", 36, "Semaphore Limiter for adding a watermark to an image.")
	flag_g_sem_darkimage = flag.Int("darkimage", 36, "Semaphore Limiter for converting an image to dark mode.")
	flag_g_sem_filedata = flag.Int("filedata", 369, "Semaphore Limiter for writing metadata about a processed file to JSON.")
	flag_g_sem_shastring = flag.Int("shastring", 369, "Semaphore Limiter for calculating the SHA256 checksum of a string.")
	flag_g_sem_wjsonfile = flag.Int("wjsonfile", 369, "Semaphore Limiter for writing a JSON file to disk.")

	flag.Usage = func() {
		_, err := fmt.Fprintf(os.Stderr, "Usage: %s -file FILE -output-dir DIRECTORY [ -limit INT | -buffer INT ]\n", os.Args[0])
		if err != nil {
			log.Println(err)
			return
		}
		flag.PrintDefaults()
	}

	flag.Parse()

	if *fileFlag == "" || *dirFlag == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *semaphoreLimitFlag > 0 {
		SemaLimiter = *semaphoreLimitFlag
	}

	if *fileBufferSize > 0 {
		BufferSize = *fileBufferSize
	}

	if len(*dirFlag) > 0 {
		DataDir = filepath.Join(".", *dirFlag)
		fmt.Println("Tmp path: " + DataDir)
		if !IsDir(DataDir) {
			panic(fmt.Sprintf("FATAL ERROR: %v is not a directory and cannot be used for saving content...", *dirFlag))
		}
	} else {
		panic("-dir is a required flag to run this program")
	}

	logFilename := filepath.Join(".", "logs", fmt.Sprintf("engine-%04d-%02d-%02d-%02d-%02d-%02d.log",
		startedAt.Year(), startedAt.Month(), startedAt.Day(), startedAt.Hour(), startedAt.Minute(), startedAt.Second()))
	logFile, logFileErr := os.OpenFile(logFilename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if logFileErr != nil {
		log.Fatal("Failed to open log file: ", logFileErr)
	}
	log.SetOutput(logFile)

	watchdog := make(chan os.Signal, 1)
	signal.Notify(watchdog, os.Kill, syscall.SIGTERM, os.Interrupt)
	go func() {
		<-watchdog
		err := logFile.Close()
		if err != nil {
			log.Printf("failed to close the logFile due to error: %v", err)
		}
		fmt.Println("Program killed!")
		os.Exit(1)
	}()

	cryptonymFile, cryptonymFileErr := os.ReadFile(filepath.Join(".", "importable", "cryptonyms.json"))
	if cryptonymFileErr != nil {
		log.Printf("failed to parse cryptonyms.json file from the data directory due to error: %v", cryptonymFileErr)
	} else {
		cryptonymMarshalErr := json.Unmarshal(cryptonymFile, &Cryptonyms)
		if cryptonymMarshalErr != nil {
			log.Printf("failed to load the Cryptonyms due to error %v", cryptonymMarshalErr)
		}
		//log.Printf("Cryptonyms generated as: %v", Cryptonyms)
	}

	var (
		err error
		ctx = context.WithValue(context.Background(), "filename", *fileFlag)
	)

	go receiveImportedRow(ctx, ch_ImportedRow)         // runs validatePdf before sending into ch_ExtractText
	go receiveOnExtractTextCh(ctx, ch_ExtractText)     // runs extractPlainTextFromPdf before sending into ch_ExtractPages
	go receiveOnExtractPagesCh(ctx, ch_ExtractPages)   // runs extractPagesFromPdf before sending PendingPage into ch_GeneratePng
	go receiveOnGeneratePngCh(ctx, ch_GeneratePng)     // runs convertPageToPng before sending PendingPage into ch_GenerateLight
	go receiveOnGenerateLightCh(ctx, ch_GenerateLight) // runs generateLightThumbnails before sending PendingPage into ch_GenerateDark
	go receiveOnGenerateDarkCh(ctx, ch_GenerateDark)   // runs generateDarkThumbnails before sending PendingPage into ch_ConvertToJpg
	go receiveOnConvertToJpg(ctx, ch_ConvertToJpg)     // runs convertPngToJpg before sending PendingPage into ch_PerformOcr
	go receiveOnPerformOcrCh(ctx, ch_PerformOcr)       // runs performOcrOnPdf before sending PendingPage into ch_CompletedPage

	if strings.Contains(*fileFlag, ".csv") || strings.Contains(*fileFlag, ".psv") {
		err = loadCsv(ctx, *fileFlag, processRecord) // parse the file
	} else if strings.Contains(*fileFlag, ".xlsx") {
		err = loadXlsx(ctx, *fileFlag, processRecord) // parse the file
	} else {
		panic(fmt.Sprintf("unable to parse file %v", *fileFlag))
	}

	if err != nil {
		log.Printf("received an error from loadCsv/loadXlsx namely: %v", err) // a problem habbened
	}

	defer logFile.Close()

	go func() {
		PerformingWork.Wait()
		close(ch_ImportedRow)   // step 0
		close(ch_ExtractText)   // step 1
		close(ch_ExtractPages)  // step 2
		close(ch_GeneratePng)   // step 3
		close(ch_GenerateLight) // step 4
		close(ch_GenerateDark)  // step 5
		close(ch_ConvertToJpg)  // step 6
		close(ch_PerformOcr)    // step 7
		close(ch_CompletedPage) // step 8
		ch_Done <- struct{}{}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ch_Done:
			log.SetOutput(os.Stdout)
			log.Printf("done processing everything... time to end things now!")
			watchdog <- os.Kill
		case pp, ok := <-ch_CompletedPage:
			if ok {
				log.Printf("Completed processing page %v (ID: %v) from Document %v",
					pp.PageNumber, pp.Identifier, pp.RecordIdentifier)
			}
		}
	}

}
