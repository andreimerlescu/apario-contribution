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
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func analyze_StartOnFullText(ctx context.Context, pp PendingPage) {
	defer func() {
		wg_active_tasks.Done()
		ch_AnalyzeCryptonyms <- pp
	}()
}

func analyzeCryptonyms(ctx context.Context, pp PendingPage) {
	defer func() {
		wg_active_tasks.Done()
		ch_AnalyzeLocations <- pp
	}()
}

func analyzeLocations(ctx context.Context, pp PendingPage) {
	defer func() {
		wg_active_tasks.Done()
		ch_AnalyzeGematria <- pp
	}()

	for {
		if a_b_locations_loaded.Load() {
			break
		}
		select {
		case <-time.After(9 * time.Second):
			log.Printf("waiting for locations to finish loading before running analyzeLocations(%v)", pp.OCRTextPath)
			continue
		case <-ctx.Done():
			return
		}
	}

	done := make(chan struct{})

	var fileLocations []*Location

	go func() {
		defer func() {
			done <- struct{}{}
			close(done)
		}()

		file, fileErr := os.Open(pp.OCRTextPath)
		if fileErr != nil {
			log.Printf("Error opening file %q: %v\n", pp.OCRTextPath, fileErr)
			return
		}

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				line := scanner.Text()
				words := strings.Fields(line)
				for _, word := range words {
					if location, ok := m_location_countries[word]; ok {
						fileLocations = append(fileLocations, location)
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Println(err)
		}

		file.Close()

		return
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			output := fmt.Sprintf("Locations in OCR file %v", pp.OCRTextPath)
			for _, l := range fileLocations {
				output += fmt.Sprintf("-> city `%v` in country `%v` state `%v`", l.City, l.Country, l.State)
			}
			pp.Locations = fileLocations
			log.Println(output)
			return
		}
	}
}

func analyzeGematria(ctx context.Context, pp PendingPage) {
	defer func() {
		wg_active_tasks.Done()
		ch_AnalyzeDictionary <- pp
	}()

	for {
		if a_b_dictionary_loaded.Load() {
			break
		}
		select {
		case <-time.After(9 * time.Second):
			log.Printf("waiting for word dictionary to finish loading before running analyzeGematria(%v)", pp.OCRTextPath)
			continue
		case <-ctx.Done():
			return
		}
	}

	done := make(chan struct{})

	var fileResults = map[string][]WordResult{}

	go func() {
		defer func() {
			done <- struct{}{}
			close(done)
		}()

		file, fileErr := os.Open(pp.OCRTextPath)
		if fileErr != nil {
			log.Printf("Error opening file %q: %v\n", pp.OCRTextPath, fileErr)
			return
		}
		defer func() {
			file.Close()
		}()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				line := scanner.Text()
				words := strings.Fields(line)
				for _, word := range words {
					for language, dictionary := range m_language_dictionary {
						if _, ok := dictionary[word]; ok {
							wr := WordResult{
								Word:     word,
								Language: language,
								Gematria: calculateGemAnalysis(word),
							}
							_, found := fileResults[language]
							if !found {
								fileResults[language] = []WordResult{}
							}
							fileResults[language] = append(fileResults[language], wr)
						}
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Println(err)
		}

		return
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			output := fmt.Sprintf("Words in OCR file %v", pp.OCRTextPath)
			var languages = map[string]int{}
			var selectedLanguage string
			var totalWords int
			for language, results := range fileResults {
				languages[language] = len(results)
			}
			for language, count := range languages {
				if len(selectedLanguage) == 0 || count > totalWords {
					selectedLanguage = language
					totalWords = count
				}
			}
			pp.Language = selectedLanguage
			pp.Words = fileResults[selectedLanguage]
			for _, wr := range pp.Words {
				output += fmt.Sprintf("-> %v (%v) = %v", wr.Word, wr.Language, wr.Gematria)
			}
			log.Println(output)
			return
		}
	}

}

func analyzeWordIndexer(ctx context.Context, pp PendingPage) {
	defer func() {
		wg_active_tasks.Done()
		ch_CompletedPage <- pp
	}()

	for {
		if a_b_dictionary_loaded.Load() {
			break
		}
		select {
		case <-time.After(9 * time.Second):
			log.Printf("waiting for word dictionary to finish loading before running analyzeWordIndexer(%v)", pp.OCRTextPath)
			continue
		case <-ctx.Done():
			return
		}
	}

	sm_pages.Store(pp.Identifier, pp)
	err := WritePendingPageToJson(pp)
	if err != nil {
		log.Printf("failed to write pending page %v to %v because of error %v", pp.Identifier, pp.ManifestPath, err)
	}

	return
}
