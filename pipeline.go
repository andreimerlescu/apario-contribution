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
	"bytes"
	`context`
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func validatePdf(ctx context.Context, record ResultData) (ResultData, error) {
	wg_active_tasks.Add(1)
	defer wg_active_tasks.Done()
	log.Printf("started validatePdf(%v) = %v", record.Identifier, record.PDFPath)

	_, rjsonErr := os.Stat(record.RecordPath)
	if os.IsNotExist(rjsonErr) {
		/*
			pdfcpu validate REPLACE_WITH_FILE_PATH | grep 'validation ok'
		*/
		cmd0_validate_pdf := exec.Command(m_required_binaries["pdfcpu"], "validate", record.PDFPath)
		var cmd0_validate_pdf_stdout bytes.Buffer
		var cmd0_validate_pdf_stderr bytes.Buffer
		cmd0_validate_pdf.Stdout = &cmd0_validate_pdf_stdout
		cmd0_validate_pdf.Stderr = &cmd0_validate_pdf_stderr
		sem_pdfcpu.Acquire()
		cmd0_validate_pdf_err := cmd0_validate_pdf.Run()
		sem_pdfcpu.Release()

		if cmd0_validate_pdf_err != nil {
			return record, fmt.Errorf("Failed to execute `pdfcpu validate %v` due to error: %s\n", record.PDFPath, cmd0_validate_pdf_err)
		}

		if !strings.Contains(cmd0_validate_pdf_stdout.String(), "validation ok") {
			return record, fmt.Errorf("failed to validate the pdf %v\n\tSTDOUT = %v", record.PDFPath, cmd0_validate_pdf_stdout.String())
		}
		/*
			gs -q -sDEVICE=pdfwrite -dCompatibilityLevel=1.7 -o REPLACE_WITH_FILE_PATH REPLACE_WITH_FILE_PATH
		*/
		cmd1_convert_pdf := exec.Command(m_required_binaries["gs"], "-q -sDEVICE=pdfwrite -dCompatibilityLevel=1.7 -o", record.PDFPath, record.PDFPath)
		var cmd1_convert_pdf_stdout bytes.Buffer
		var cmd1_convert_pdf_stderr bytes.Buffer
		cmd1_convert_pdf.Stdout = &cmd1_convert_pdf_stdout
		cmd1_convert_pdf.Stderr = &cmd1_convert_pdf_stderr
		sem_gs.Acquire()
		cmd1_convert_pdf_err := cmd1_convert_pdf.Run()
		sem_gs.Release()
		if cmd1_convert_pdf_err != nil {
			return record, fmt.Errorf("Failed to execute command `gs -q -sDEVICE=pdfwrite -dCompatibilityLevel=1.7 -o %v %v` due to error: %s\n", record.PDFPath, record.PDFPath, cmd1_convert_pdf_err)
		}

		/*
			pdfcpu optimize REPLACE_WITH_FILE_PATH
		*/
		cmd2_optimize_pdf := exec.Command(m_required_binaries["pdfcpu"], "optimize", record.PDFPath)
		var cmd2_optimize_pdf_stdout bytes.Buffer
		var cmd2_optimize_pdf_stderr bytes.Buffer
		cmd2_optimize_pdf.Stdout = &cmd2_optimize_pdf_stdout
		cmd2_optimize_pdf.Stderr = &cmd2_optimize_pdf_stderr
		sem_pdfcpu.Acquire()
		cmd2_optimize_pdf_err := cmd2_optimize_pdf.Run()
		sem_pdfcpu.Release()
		if cmd2_optimize_pdf_err != nil {
			return record, fmt.Errorf("Failed to execute command `pdfcpu optimize %v` due to error: %s\n", record.PDFPath, cmd2_optimize_pdf_err)
		}
	}

	return record, nil
}

func extractPlainTextFromPdf(ctx context.Context, record ResultData) {
	wg_active_tasks.Add(1)
	defer wg_active_tasks.Done()
	defer func() {
		log.Printf("finished extracting the text from the PDF %v, now sending rd into ch_ExtractPages", filepath.Base(record.PDFPath))
		ch_ExtractPages <- record
	}()
	log.Printf("started extractPlainTextFromPdf(%v) = %v", record.Identifier, record.PDFPath)
	if ok, err := fileHasData(record.ExtractedTextPath); !ok || err != nil {
		/*
			pdftotext REPLACE_WITH_FILE_PATH REPLACE_WITH_TEXT_OUTPUT_FILE_PATH
		*/
		cmd4_extract_text_pdf := exec.Command(m_required_binaries["pdftotext"], record.PDFPath, record.ExtractedTextPath)
		var cmd4_extract_text_pdf_stdout bytes.Buffer
		var cmd4_extract_text_pdf_stderr bytes.Buffer
		cmd4_extract_text_pdf.Stdout = &cmd4_extract_text_pdf_stdout
		cmd4_extract_text_pdf.Stderr = &cmd4_extract_text_pdf_stderr
		sem_pdftotext.Acquire()
		cmd4_extract_text_pdf_err := cmd4_extract_text_pdf.Run()
		sem_pdftotext.Release()
		if cmd4_extract_text_pdf_err != nil {
			log.Printf("Failed to execute command `pdftotext %v %v` due to error: %s\n", record.PDFPath, record.ExtractedTextPath, cmd4_extract_text_pdf_err)
			return
		}
	}
}

func extractPagesFromPdf(ctx context.Context, record ResultData) {
	wg_active_tasks.Add(1)
	defer wg_active_tasks.Done()
	log.Printf("started extractPagesFromPdf(%v) = %v", record.Identifier, record.PDFPath)
	/*
		pdfcpu extract -mode page REPLACE_WITH_FILE_PATH REPLACE_WITH_OUTPUT_DIRECTORY
	*/
	pagesDir := filepath.Join(record.DataDir, "pages")
	sm_page_directories.Store(record.Identifier, pagesDir)
	_, pagesDirExistsErr := os.Stat(pagesDir)
	performPagesExtract := false
	if os.IsNotExist(pagesDirExistsErr) {
		performPagesExtract = true
	} else {
		ok, err := DirHasPDFs(pagesDir)
		if err == nil && ok {
			performPagesExtract = true
		}
	}
	if performPagesExtract {
		pagesDirErr := os.MkdirAll(pagesDir, 0755)
		if pagesDirErr != nil {
			log.Printf("failed to create directory %v due to error %v", pagesDir, pagesDirErr)
			return
		}
		cmd5_extract_pages_in_pdf := exec.Command(m_required_binaries["pdfcpu"], "extract", "-mode", "page", record.PDFPath, pagesDir)
		var cmd5_extract_pages_in_pdf_stdout bytes.Buffer
		var cmd5_extract_pages_in_pdf_stderr bytes.Buffer
		cmd5_extract_pages_in_pdf.Stdout = &cmd5_extract_pages_in_pdf_stdout
		cmd5_extract_pages_in_pdf.Stderr = &cmd5_extract_pages_in_pdf_stderr
		sem_pdfcpu.Acquire()
		cmd5_extract_pages_in_pdf_err := cmd5_extract_pages_in_pdf.Run()
		sem_pdfcpu.Release()
		if cmd5_extract_pages_in_pdf_err != nil {
			log.Printf("Failed to execute command `pdfcpu extract -mode page %v %v` due to error: %s\n", record.PDFPath, pagesDir, cmd5_extract_pages_in_pdf_err)
			return
		}
	} else {
		log.Printf("not performing `pdfcpu extrace -mode page %v %v` because the directory %v already has PDFs inside it", record.PDFPath, pagesDir, pagesDir)
	}

	pagesDirWalkErr := filepath.Walk(pagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing a path %q: %v\n", path, err)
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".pdf") {
			nameParts := strings.Split(info.Name(), "_page_")
			if len(nameParts) < 2 {
				return fmt.Errorf("incorrect filename provided as %v", info.Name())
			}
			pgNoStr := strings.ReplaceAll(nameParts[1], ".pdf", "")
			pgNo, pgNoErr := strconv.Atoi(pgNoStr)
			if pgNoErr != nil {
				return fmt.Errorf("failed to extract the pgNo from the PDF filename %v", info.Name())
			}
			identifier := NewIdentifier(9)
			pp := PendingPage{
				Identifier:       identifier,
				RecordIdentifier: record.Identifier,
				PageNumber:       pgNo,
				PagesDir:         pagesDir,
				PDFPath:          path,
				OCRTextPath:      filepath.Join(pagesDir, fmt.Sprintf("ocr.%06d.txt", pgNo)),
				Light: Images{
					Original: filepath.Join(pagesDir, fmt.Sprintf("page.light.%06d.original.png", pgNo)),
					Large:    filepath.Join(pagesDir, fmt.Sprintf("page.light.%06d.large.png", pgNo)),
					Medium:   filepath.Join(pagesDir, fmt.Sprintf("page.light.%06d.medium.png", pgNo)),
					Small:    filepath.Join(pagesDir, fmt.Sprintf("page.light.%06d.small.png", pgNo)),
					Social:   filepath.Join(pagesDir, fmt.Sprintf("page.light.%06d.social.png", pgNo)),
				},
				Dark: Images{
					Original: filepath.Join(pagesDir, fmt.Sprintf("page.dark.%06d.original.png", pgNo)),
					Large:    filepath.Join(pagesDir, fmt.Sprintf("page.dark.%06d.large.png", pgNo)),
					Medium:   filepath.Join(pagesDir, fmt.Sprintf("page.dark.%06d.medium.png", pgNo)),
					Small:    filepath.Join(pagesDir, fmt.Sprintf("page.dark.%06d.small.png", pgNo)),
					Social:   filepath.Join(pagesDir, fmt.Sprintf("page.dark.%06d.social.png", pgNo)),
				},
			}
			sm_pages.Store(pp.Identifier, pp)
			err := WritePendingPageToJson(pp, filepath.Join(pagesDir, fmt.Sprintf("manifest.%06d.json", pgNo)))
			if err != nil {
				return err
			}
			log.Printf("sending page %d (ID %v) from record %v URL %v into the ch_GeneratingPng", pgNo, identifier, record.Identifier, record.URL)
			ch_GeneratePng <- pp
		}

		return nil
	})

	if pagesDirWalkErr != nil {
		log.Printf("Error walking the path ./pages: %v\n", pagesDirWalkErr)
		return
	}

	return
}

func convertPageToPng(ctx context.Context, pp PendingPage) {
	wg_active_tasks.Add(1)
	defer wg_active_tasks.Done()
	log.Printf("started convertPageToPng(%v.%v) = %v", pp.RecordIdentifier, pp.Identifier, pp.PDFPath)
	/*
		pdf_to_png: "pdftoppm REPLACE_WITH_PNG_OPTS REPLACE_WITH_FILE_PATH REPLACE_WITH_PNG_PATH",
	*/
	_, loErr := os.Stat(pp.Light.Original)
	if os.IsNotExist(loErr) {
		originalFilename := strings.ReplaceAll(pp.Light.Original, `.png`, ``)
		cmd := exec.Command(m_required_binaries["pdftoppm"],
			`-r`, `369`, `-png`, `-freetype`, `yes`, `-aa`, `yes`, `-aaVector`, `yes`, `-thinlinemode`, `solid`,
			pp.PDFPath, originalFilename)
		var cmd_stdout bytes.Buffer
		var cmd_stderr bytes.Buffer
		cmd.Stdout = &cmd_stdout
		cmd.Stderr = &cmd_stderr
		sem_pdftoppm.Acquire()
		cmd_err := cmd.Run()
		sem_pdftoppm.Release()
		if cmd_err != nil {
			log.Printf("failed to convert page %v to png %v due to error: %s\n", filepath.Base(pp.PDFPath), pp.Light.Original, cmd_err)
			return
		}

		pngRenameErr := os.Rename(fmt.Sprintf("%v-1.png", originalFilename), fmt.Sprintf("%v.png", originalFilename))
		if pngRenameErr != nil {
			log.Printf("failed to rename the jpg %v due to error: %v", originalFilename, pngRenameErr)
			return
		}
	}

	log.Printf("completed convertPageToPng now sending %v (%v.%v) -> ch_GenerateLight ", pp.PDFPath, pp.RecordIdentifier, pp.Identifier)
	ch_GenerateLight <- pp
	return
}

func generateLightThumbnails(ctx context.Context, pp PendingPage) {
	wg_active_tasks.Add(1)
	defer wg_active_tasks.Done()
	defer func() {
		log.Printf("completed generateLightThumbnails now sending %v (%v.%v) -> ch_GenerateDark ", pp.PDFPath, pp.RecordIdentifier, pp.Identifier)
		ch_GenerateDark <- pp
	}()
	log.Printf("started generateLightThumbnails(%v.%v) = %v", pp.RecordIdentifier, pp.Identifier, pp.PDFPath)

	original, err := os.Open(pp.Light.Original)
	if err != nil {
		log.Printf("failed to open pp.OriginalPath(%v) due to error %v", pp.Light.Original, err)
		return
	}
	// create the large thumbnail from the JPG
	_, llgErr := os.Stat(pp.Light.Large)
	if os.IsNotExist(llgErr) {
		lgResizeErr := resizePng(original, 999, pp.Light.Large)
		if lgResizeErr != nil {
			log.Printf("failed to resize jpg %v due to error %v", pp.Light.Large, lgResizeErr)
			return
		}
	}

	// create the medium thumbnail from the JPG
	_, lmdErr := os.Stat(pp.Light.Medium)
	if os.IsNotExist(lmdErr) {
		mdResizeErr := resizePng(original, 666, pp.Light.Medium)
		if mdResizeErr != nil {
			log.Printf("failed to resize jpg %v due to error %v", pp.Light.Medium, mdResizeErr)
			return
		}
	}

	// create the small thumbnail from the JPG
	_, lsmErr := os.Stat(pp.Light.Small)
	if os.IsNotExist(lsmErr) {
		smResizeErr := resizePng(original, 333, pp.Light.Small)
		if smResizeErr != nil {
			log.Printf("failed to resize jpg %v due to error %v", pp.Light.Small, smResizeErr)
			return
		}
	}

}

func generateDarkThumbnails(ctx context.Context, pp PendingPage) {
	wg_active_tasks.Add(1)
	defer wg_active_tasks.Done()
	defer func() {
		log.Printf("completed generateDarkThumbnails now sending %v (%v.%v) -> ch_PerformOcr ", pp.PDFPath, pp.RecordIdentifier, pp.Identifier)
		ch_PerformOcr <- pp
	}()
	log.Printf("started generateDarkThumbnails(%v.%v) = %v", pp.RecordIdentifier, pp.Identifier, pp.PDFPath)
	// task: the pp.Light.Original into pp.Dark.Original

	_, ppdoErr := os.Stat(pp.Dark.Original)
	if os.IsNotExist(ppdoErr) {
		// convert REPLACE_WITH_OUTPUT_PNG_PAGE_FILENAME -channel rgba -matte -fill 'rgba(250,226,203,1)' -fuzz 45% -opaque 'rgba(76,76,76,1)' -flatten REPLACE_WITH_OUTPUT_PNG_DARK_PAGE_FILENAME
		cmdA := exec.Command(m_required_binaries["convert"], pp.Light.Original, "-channel", "rgba", "-matte", "-fill", `rgba(250,226,203,1)`, "-fuzz", "45%", "-opaque", `rgba(76,76,76,1)`, "-flatten", pp.Dark.Original)
		var cmdA_stdout bytes.Buffer
		var cmdA_stderr bytes.Buffer
		cmdA.Stdout = &cmdA_stdout
		cmdA.Stderr = &cmdA_stderr
		sem_convert.Acquire()
		cmdA_err := cmdA.Run()
		sem_convert.Release()
		if cmdA_err != nil {
			log.Printf("failed to convert %v into %v due to error: %s\n", pp.Light.Original, pp.Dark.Original, cmdA_err)
			return
		}

		// convert REPLACE_WITH_OUTPUT_PNG_DARK_PAGE_FILENAME -channel rgba -matte -fill 'rgba(40,40,86,1)' -fuzz 12% -opaque white -flatten REPLACE_WITH_OUTPUT_PNG_DARK_PAGE_FILENAME
		cmdB := exec.Command(m_required_binaries["convert"], pp.Dark.Original, `-channel`, `rgba`, `-matte`, `-fill`, `rgba(40,40,86,1)`, `-fuzz`, `12%`, `-opaque`, `white`, `-flatten`, pp.Dark.Original)
		var cmdB_stdout bytes.Buffer
		var cmdB_stderr bytes.Buffer
		cmdB.Stdout = &cmdB_stdout
		cmdB.Stderr = &cmdB_stderr
		sem_convert.Acquire()
		cmdB_err := cmdB.Run()
		sem_convert.Release()
		if cmdB_err != nil {
			log.Printf("failed to convert %v into %v due to error: %s\n", pp.Light.Original, pp.Dark.Original, cmdB_err)
			return
		}
	}

	original, err := os.Open(pp.Dark.Original)
	if err != nil {
		log.Printf("failed to open pp.OriginalPath(%v) due to error %v", pp.Dark.Original, err)
		return
	}

	// create the large thumbnail from the JPG
	_, dlgErr := os.Stat(pp.Dark.Large)
	if os.IsNotExist(dlgErr) {
		lgResizeErr := resizePng(original, 999, pp.Dark.Large)
		if lgResizeErr != nil {
			log.Printf("failed to resize jpg %v due to error %v", pp.Dark.Large, lgResizeErr)
			return
		}
	}

	// create the medium thumbnail from the JPG
	_, dmdErr := os.Stat(pp.Dark.Medium)
	if os.IsNotExist(dmdErr) {
		mdResizeErr := resizePng(original, 666, pp.Dark.Medium)
		if mdResizeErr != nil {
			log.Printf("failed to resize jpg %v due to error %v", pp.Dark.Medium, mdResizeErr)
			return
		}
	}

	// create the small thumbnail from the JPG
	_, dsmErr := os.Stat(pp.Dark.Small)
	if os.IsNotExist(dsmErr) {
		smResizeErr := resizePng(original, 333, pp.Dark.Small)
		if smResizeErr != nil {
			log.Printf("failed to resize jpg %v due to error %v", pp.Dark.Small, smResizeErr)
			return
		}
	}

}

func performOcrOnPdf(ctx context.Context, pp PendingPage) {
	wg_active_tasks.Add(1)
	defer wg_active_tasks.Done()
	defer func() {
		log.Printf("completed performOcrOnPdf now sending %v (%v.%v) -> ch_ConvertToJpg ", pp.PDFPath, pp.RecordIdentifier, pp.Identifier)
		ch_ConvertToJpg <- pp
	}()

	if ok, err := fileHasData(pp.OCRTextPath); !ok || err != nil {
		/*
			tesseract REPLACE_WITH_FILE_PATH REPLACE_WITH_TEXT_OUTPUT_FILE_PATH -l eng --psm 1
		*/
		img := filepath.Join(pp.PagesDir, pp.Light.Original)
		ocrStat, ppOcrPathErr := os.Stat(pp.OCRTextPath)
		if (ppOcrPathErr == nil || !os.IsNotExist(ppOcrPathErr)) && ocrStat.Size() > 0 {
			ocrText, ocrTextErr := os.ReadFile(pp.OCRTextPath)
			if ocrTextErr != nil && len(string(ocrText)) > 6 {
				log.Printf("finished performOcrOnPdf(%v.%v) because the file %v already has %d bytes inside it!", pp.RecordIdentifier, pp.Identifier, pp.OCRTextPath, ocrStat.Size())
				return
			}
		}
		cmd8 := exec.Command(m_required_binaries["tesseract"], img, pp.OCRTextPath, `-l`, `eng`, `--psm`, `1`)
		var cmd8_stdout bytes.Buffer
		var cmd8_stderr bytes.Buffer
		cmd8.Stdout = &cmd8_stdout
		cmd8.Stderr = &cmd8_stderr
		log.Printf("started performOcrOnPdf(%v.%v) = %v (WAITING)", pp.RecordIdentifier, pp.Identifier, pp.PDFPath)
		sem_tesseract.Acquire()
		log.Printf("running performOcrOnPdf(%v.%v) = %v", pp.RecordIdentifier, pp.Identifier, pp.PDFPath)
		cmd8_err := cmd8.Run()
		sem_tesseract.Release()
		log.Printf("completed performOcrOnPdf(%v.%v) = %v", pp.RecordIdentifier, pp.Identifier, pp.PDFPath)
		if cmd8_err != nil {
			log.Printf("Command `tesseract %v %v -l eng --psm 1` failed with error: %s\n", img, pp.OCRTextPath, cmd8_err)
			return
		}
	}
}

func convertPngToJpg(ctx context.Context, pp PendingPage) {
	wg_active_tasks.Add(1)
	defer wg_active_tasks.Done()
	log.Printf("started convertPngToJpg(%v.%v) = %v", pp.RecordIdentifier, pp.Identifier, pp.PDFPath)
	err := filepath.Walk(pp.PagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.ToLower(filepath.Ext(path)) == ".png" {
			inputFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer inputFile.Close()

			output := strings.TrimSuffix(path, filepath.Ext(path)) + ".jpg"

			_, oErr := os.Stat(output)
			if os.IsNotExist(oErr) {
				err = convertAndOptimizePNG(inputFile, output)
				if err != nil {
					return err
				}

				// Delete the PNG file
				err = os.Remove(path)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("failed to walk the dir of %v because of error %v", pp.PagesDir, err)
		return
	}

	pp = PendingPage{
		Identifier:       pp.Identifier,
		RecordIdentifier: pp.RecordIdentifier,
		PageNumber:       pp.PageNumber,
		PagesDir:         pp.PagesDir,
		PDFPath:          pp.PDFPath,
		OCRTextPath:      pp.OCRTextPath,
		Light: Images{
			Original: filepath.Join(pp.PagesDir, fmt.Sprintf("light.%06d.original.jpg", pp.PageNumber)),
			Large:    filepath.Join(pp.PagesDir, fmt.Sprintf("light.%06d.large.jpg", pp.PageNumber)),
			Medium:   filepath.Join(pp.PagesDir, fmt.Sprintf("light.%06d.medium.jpg", pp.PageNumber)),
			Small:    filepath.Join(pp.PagesDir, fmt.Sprintf("light.%06d.small.jpg", pp.PageNumber)),
			Social:   filepath.Join(pp.PagesDir, fmt.Sprintf("light.%06d.social.jpg", pp.PageNumber)),
		},
		Dark: Images{
			Original: filepath.Join(pp.PagesDir, fmt.Sprintf("dark.%06d.original.jpg", pp.PageNumber)),
			Large:    filepath.Join(pp.PagesDir, fmt.Sprintf("dark.%06d.large.jpg", pp.PageNumber)),
			Medium:   filepath.Join(pp.PagesDir, fmt.Sprintf("dark.%06d.medium.jpg", pp.PageNumber)),
			Small:    filepath.Join(pp.PagesDir, fmt.Sprintf("dark.%06d.small.jpg", pp.PageNumber)),
			Social:   filepath.Join(pp.PagesDir, fmt.Sprintf("dark.%06d.social.jpg", pp.PageNumber)),
		},
	}
	sm_pages.Store(pp.Identifier, pp)
	err = WritePendingPageToJson(pp, filepath.Join(pp.PagesDir, fmt.Sprintf("manifest.%06d.json", pp.PageNumber)))
	if err != nil {
		log.Printf("failed to overwrite the pp file for identifier %v due to error %v", pp.Identifier, err)
		return
	}

	log.Printf("completed convertPngToJpg now sending %v (%v.%v) -> ch_CompletedPage ", pp.PDFPath, pp.RecordIdentifier, pp.Identifier)
	ch_CompletedPage <- pp
}
