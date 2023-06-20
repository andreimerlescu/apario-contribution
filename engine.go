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
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	`os/exec`
	"os/signal"
	"path/filepath"
	`runtime`
	"strings"
	"syscall"
	`time`
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	for _, arg := range os.Args {
		if arg == "help" {
			fmt.Println(config.Usage())
			os.Exit(0)
		}
		if arg == "show" {
			for _, innerArg := range os.Args {
				if innerArg == "w" || innerArg == "c" {
					license, err := os.ReadFile(filepath.Join(".", "LICENSE"))
					if err != nil {
						fmt.Printf("Cannot find the license file to load to comply with the GNU-3 license terms. This program was modified outside of its intended runtime use.")
						os.Exit(1)
					} else {
						fmt.Printf("%v\n", string(license))
						os.Exit(1)
					}
				}
			}
		}
	}

	configErr := config.Parse(filepath.Join(".", "config.yaml"))
	if configErr != nil {
		log.Fatalf("failed to parse config.yaml due to err: %v", configErr)
	}

	binaryErr := verifyBinaries(sl_required_binaries)
	if binaryErr != nil {
		fmt.Printf("Error: %s\n", binaryErr)
		os.Exit(1)
	}

	ex, execErr := os.Getwd()
	if execErr != nil {
		panic(execErr)
	}

	dir_current_directory = filepath.Dir(ex)
	fmt.Sprintf("Current Working Directory: %s\n", dir_current_directory)

	if *flag_s_file == "" || *flag_s_directory == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *flag_i_sem_limiter > 0 {
		channel_buffer_size = *flag_i_sem_limiter
	}

	if *flag_i_buffer > 0 {
		reader_buffer_bytes = *flag_i_buffer
	}

	if len(*flag_s_directory) > 0 {
		dir_data_directory = filepath.Join(".", *flag_s_directory)
		fmt.Println("Tmp path: " + dir_data_directory)
		if !IsDir(dir_data_directory) {
			panic(fmt.Sprintf("FATAL ERROR: %v is not a directory and cannot be used for saving content...", *flag_s_directory))
		}
	} else {
		panic("-dir is a required flag to run this program")
	}

	logFile, logFileErr := os.OpenFile(*flag_g_log_file, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
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
		cancel()

		wg_active_tasks.PreventAdd()

		ch_ImportedRow.Close()       // step 01
		ch_ExtractText.Close()       // step 02
		ch_ExtractPages.Close()      // step 03
		ch_GeneratePng.Close()       // step 04
		ch_GenerateLight.Close()     // step 05
		ch_GenerateDark.Close()      // step 06
		ch_ConvertToJpg.Close()      // step 07
		ch_PerformOcr.Close()        // step 08
		ch_AnalyzeText.Close()       // step 09
		ch_AnalyzeCryptonyms.Close() // step 10
		ch_AnalyzeLocations.Close()  // step 11
		ch_AnalyzeGematria.Close()   // step 12
		ch_AnalyzeDictionary.Close() // step 13
		ch_CompletedPage.Close()     // step 14
		ch_CompiledDocument.Close()  // step 15

		fmt.Printf("Completed running in %d", time.Since(startedAt))

		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("tasklist", "/FI", "IMAGENAME eq apario-contribution.exe")
		default:
			cmd = exec.Command("pgrep", "apario-contribution")
		}

		output, err := cmd.Output()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		pids := parsePIDs(string(output))

		for _, pid := range pids {
			terminatePID(pid)
		}

		os.Exit(0)
	}()

	a_b_ch_ImportRow_closed.Store(false)
	a_b_ch_ExtractText_closed.Store(false)
	a_b_ch_ExtractPages_closed.Store(false)
	a_b_ch_GeneratePng_closed.Store(false)
	a_b_ch_GenerateLight_closed.Store(false)
	a_b_ch_GenerateDark_closed.Store(false)
	a_b_ch_ConvertToJpg_closed.Store(false)
	a_b_ch_PerformOcr_closed.Store(false)
	a_b_ch_AnalyzeText_closed.Store(false)
	a_b_ch_AnalyzeCryptonyms_closed.Store(false)
	a_b_ch_AnalyzeLocations_closed.Store(false)
	a_b_ch_AnalyzeGematria_closed.Store(false)
	a_b_ch_AnalyzeDictionary_closed.Store(false)
	a_b_ch_CompletedPage_closed.Store(false)
	a_b_ch_CompiledDocument_closed.Store(false)

	a_b_dictionary_loaded.Store(false)
	go populateDictionary()

	cryptonymFile, cryptonymFileErr := os.ReadFile(filepath.Join(".", "importable", "cryptonyms.json"))
	if cryptonymFileErr != nil {
		log.Printf("failed to parse cryptonyms.json file from the data directory due to error: %v", cryptonymFileErr)
	} else {
		cryptonymMarshalErr := json.Unmarshal(cryptonymFile, &m_cryptonyms)
		if cryptonymMarshalErr != nil {
			log.Printf("failed to load the m_cryptonyms due to error %v", cryptonymMarshalErr)
		}
		out := ""
		var cryptonyms []string
		for cryptonym, _ := range m_cryptonyms {
			cryptonyms = append(cryptonyms, cryptonym)
		}
		out = strings.Join(cryptonyms, ",")
		log.Printf("Cryptonyms to search for: %v", out)
	}

	ctx = context.WithValue(ctx, CtxKey("filename"), *flag_s_file)

	go receiveImportedRow(ctx, ch_ImportedRow.Chan())             // step 01 - runs validatePdf before sending into ch_ExtractText
	go receiveOnExtractTextCh(ctx, ch_ExtractText.Chan())         // step 02 - runs extractPlainTextFromPdf before sending into ch_ExtractPages
	go receiveOnExtractPagesCh(ctx, ch_ExtractPages.Chan())       // step 03 - runs extractPagesFromPdf before sending PendingPage into ch_GeneratePng
	go receiveOnGeneratePngCh(ctx, ch_GeneratePng.Chan())         // step 04 - runs convertPageToPng before sending PendingPage into ch_GenerateLight
	go receiveOnGenerateLightCh(ctx, ch_GenerateLight.Chan())     // step 05 - runs generateLightThumbnails before sending PendingPage into ch_GenerateDark
	go receiveOnGenerateDarkCh(ctx, ch_GenerateDark.Chan())       // step 06 - runs generateDarkThumbnails before sending PendingPage into ch_ConvertToJpg
	go receiveOnConvertToJpg(ctx, ch_ConvertToJpg.Chan())         // step 07 - runs convertPngToJpg before sending PendingPage into ch_PerformOcr
	go receiveOnPerformOcrCh(ctx, ch_PerformOcr.Chan())           // step 08 - runs performOcrOnPdf before sending PendingPage into ch_AnalyzeText
	go receiveFullTextToAnalyze(ctx, ch_AnalyzeText.Chan())       // step 09 - runs analyze_StartOnFullText before sending PendingPage into ch_AnalyzeCryptonyms
	go receiveAnalyzeCryptonym(ctx, ch_AnalyzeCryptonyms.Chan())  // step 10 - runs analyzeCryptonyms before sending PendingPage into ch_AnalyzeLocations
	go receiveAnalyzeLocations(ctx, ch_AnalyzeLocations.Chan())   // step 11 - runs analyzeLocations before sending PendingPage into ch_AnalyzeGematria
	go receiveAnalyzeGematria(ctx, ch_AnalyzeGematria.Chan())     // step 12 - runs analyzeGematria before sending PendingPage into ch_AnalyzeDictionary
	go receiveAnalyzeDictionary(ctx, ch_AnalyzeDictionary.Chan()) // step 13 - runs analyzeWordIndexer before sending PendingPage into ch_CompletedPage
	go receiveCompletedPendingPage(ctx, ch_CompletedPage.Chan())  // step 14 - compiles a final result of a Document before sending it into ch_CompiledDocument
	go receiveCompiledDocument(ctx, ch_CompiledDocument.Chan())   // step 15 - compiles the SQL insert statements for the Document

	go func() {
		wg_active_tasks.Add(1)
		defer wg_active_tasks.Done()
		locationsCsvErr := loadCsv(ctx, filepath.Join(".", "private", "locations.csv"), processLocation)
		if locationsCsvErr != nil {
			log.Printf("received an error from loadCsv/loadXlsx namely: %v", locationsCsvErr) // a problem habbened
			return
		}

		a_b_locations_loaded.Store(true)

	}()

	var importErr error
	if strings.Contains(*flag_s_file, ".csv") || strings.Contains(*flag_s_file, ".psv") {
		importErr = loadCsv(ctx, *flag_s_file, processRecord) // parse the file
	} else if strings.Contains(*flag_s_file, ".xlsx") {
		importErr = loadXlsx(ctx, *flag_s_file, processRecord) // parse the file
	} else {
		panic(fmt.Sprintf("unable to parse file %v", *flag_s_file))
	}

	if importErr != nil {
		log.Printf("received an error from loadCsv/loadXlsx namely: %v", importErr) // a problem habbened
	}

	defer logFile.Close()

	wg_active_tasks.Wait()
	ch_Done <- struct{}{}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ch_Done:
			log.SetOutput(os.Stdout)
			log.Printf("done processing everything... time to end things now!")
			watchdog <- os.Kill
		case id, ok := <-ch_CompiledDocument.Chan():
			if ok {
				d, ok := id.(Document)
				if !ok {
					log.Printf("cannot typecast the final result for %d as a .(Document)", d.Identifier)
				}
				log.Printf("Completed processing document %v", d.Identifier)
			}
		}
	}

}
