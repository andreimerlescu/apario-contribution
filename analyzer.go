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
	`sync`
	"time"
)

func analyze_StartOnFullText(ctx context.Context, pp PendingPage) {
	defer func() {
		pp_save(pp)
		wg_active_tasks.Done()
		if ch_AnalyzeCryptonyms.CanWrite() {
			err := ch_AnalyzeCryptonyms.Write(pp)
			if err != nil {
				log.Printf("cant write to the ch_AnalyzeCryptonyms channel due to error %v", err)
				return
			}
		}
	}()
	file, fileErr := os.ReadFile(pp.OCRTextPath)
	if fileErr != nil {
		log.Printf("Error opening file %q: %v\n", pp.OCRTextPath, fileErr)
		return
	}
	pp.Dates = extractDates(string(file))
}

func analyzeCryptonyms(ctx context.Context, pp PendingPage) {
	defer func() {
		pp_save(pp)
		wg_active_tasks.Done()
		if ch_AnalyzeLocations.CanWrite() {
			err := ch_AnalyzeLocations.Write(pp)
			if err != nil {
				log.Printf("cannot write to the ch_AnalyzeLocations channel due to error %v", err)
				return
			}
		}
	}()

	var result []string
	file, fileErr := os.ReadFile(pp.OCRTextPath)
	if fileErr != nil {
		log.Printf("Error opening file %q: %v\n", pp.OCRTextPath, fileErr)
		return
	}
	for key := range m_cryptonyms {
		if strings.Contains(string(file), key) {
			result = append(result, key)
		}
	}
	pp.Cryptonyms = result
}

func analyzeLocations(ctx context.Context, pp PendingPage) {
	defer func() {
		pp_save(pp)
		wg_active_tasks.Done()
		if ch_AnalyzeGematria.CanWrite() {
			err := ch_AnalyzeGematria.Write(pp)
			if err != nil {
				log.Printf("cant write to the ch_AnalyzeGematria channel due to error %v", err)
				return
			}
		}
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
						log.Printf("found a country location %v inside pp.PDFPath %v", word, pp.PDFPath)
					}
					if location, ok := m_location_states[word]; ok {
						fileLocations = append(fileLocations, location)
						log.Printf("found a state location %v inside pp.PDFPath %v", word, pp.PDFPath)
					}
					if location, ok := m_location_cities[word]; ok {
						fileLocations = append(fileLocations, location)
						log.Printf("found a city location %v inside pp.PDFPath %v", word, pp.PDFPath)
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
		pp_save(pp)
		wg_active_tasks.Done()
		if ch_AnalyzeDictionary.CanWrite() {
			err := ch_AnalyzeDictionary.Write(pp)
			if err != nil {
				log.Printf("cant write to the ch_AnalyzeDictionary channel due to error %v", err)
				return
			}
		}
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
								Gematria: Gematria{word, NewGemScore(word)},
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
			var twMu = sync.Mutex{}
			for language, results := range fileResults {
				languages[language] = len(results)
			}
			for language, count := range languages {
				if len(selectedLanguage) == 0 || count > totalWords {
					twMu.Lock()
					selectedLanguage = language
					totalWords = count
					twMu.Unlock()
				}
			}
			pp.Language = selectedLanguage
			pp.Words = fileResults[selectedLanguage]
			var deDupWords = map[string]WordResult{}
			for _, wr := range pp.Words {
				if _, exists := deDupWords[wr.Word]; !exists {
					deDupWords[wr.Word] = wr
				}
			}
			var cleanedWords []WordResult
			for _, wr := range deDupWords {
				cleanedWords = append(cleanedWords, wr)
			}
			pp.Words = cleanedWords
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
		pp_save(pp)
		wg_active_tasks.Done()
		if ch_CompletedPage.CanWrite() {
			err := ch_CompletedPage.Write(pp)
			if err != nil {
				log.Printf("cant write to the ch_CompletedPage channel due to error %v", err)
				return
			}
		}
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

	return
}
