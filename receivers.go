package main

import (
	"context"
	"log"
	"path/filepath"
)

func receiveImportedRow(ctx context.Context, ch <-chan ResultData) {
	var err error
	for {
		select {
		case <-ctx.Done():
			return
		case rd, ok := <-ch:
			if ok {
				rd, err = validatePdf(rd)
				if err != nil {
					log.Printf("received error on validatePdf for rd.URL %v ; err = %v", rd.URL, err)
				} else {
					log.Printf("validated the downloaded PDF %v from URL %v, sending rd into ch_ExtractText", filepath.Base(rd.PDFPath), rd.URL)
					ch_ExtractText <- rd
				}
			}
		}
	}
}
func receiveOnExtractTextCh(ctx context.Context, ch <-chan ResultData) {
	for {
		select {
		case <-ctx.Done():
			return
		case rd, ok := <-ch:
			if ok {
				log.Printf("received rd from ch_ExtractText for URL %v, running extractPlainTextFromPdf(%v)", rd.URL, rd.Identifier)
				go extractPlainTextFromPdf(rd)
			} else {
				log.Printf("ch_ExtractText is closed but received some data")
				return
			}
		}
	}
}
func receiveOnExtractPagesCh(ctx context.Context, ch <-chan ResultData) {
	for {
		select {
		case <-ctx.Done():
			return
		case rd, ok := <-ch:
			if ok {
				log.Printf("received on ch_ExtractPages URL %v, running extractPagesFromPdf(%v)", rd.URL, rd.Identifier)
				go extractPagesFromPdf(rd)
			} else {
				log.Printf("ch_ExtractPages is closed but received some data")
				return
			}
		}
	}
}
func receiveOnGeneratePngCh(ctx context.Context, ch <-chan PendingPage) {
	for {
		select {
		case <-ctx.Done():
			return
		case pp, ok := <-ch:
			if ok {
				log.Printf("received on ch_GeneratePng, running convertPageToPng(%v) for ID %v (pgNo %d)", filepath.Base(pp.PDFPath), pp.Identifier, pp.PageNumber)
				go convertPageToPng(pp)
			} else {
				log.Printf("ch_GeneratePng is closed but received some data")
				return
			}
		}
	}
}
func receiveOnGenerateLightCh(ctx context.Context, ch <-chan PendingPage) {
	for {
		select {
		case <-ctx.Done():
			return
		case pp, ok := <-ch:
			if ok {
				log.Printf("received on ch_GenerateLight, running generateLightThumbnails(%v) for ID %v (pgNo %d)", filepath.Base(pp.Light.Original), pp.Identifier, pp.PageNumber)
				go generateLightThumbnails(pp)
			} else {
				log.Printf("ch_GenerateLight is closed but received some data")
				return
			}
		}
	}
}
func receiveOnGenerateDarkCh(ctx context.Context, ch <-chan PendingPage) {
	for {
		select {
		case <-ctx.Done():
			return
		case pp, ok := <-ch:
			if ok {
				log.Printf("received on ch_GenerateDark, running generateDarkThumbnails(%v) for ID %v (pgNo %d)", filepath.Base(pp.Dark.Original), pp.Identifier, pp.PageNumber)
				go generateDarkThumbnails(pp)
			} else {
				log.Printf("ch_GenerateDark is closed but received some data")
				return
			}
		}
	}
}
func receiveOnPerformOcrCh(ctx context.Context, ch <-chan PendingPage) {
	for {
		select {
		case <-ctx.Done():
			return
		case pp, ok := <-ch:
			if ok {
				log.Printf("received on ch_PerformOcr, running performOcrOnPdf(%v) for ID %v (pgNo %d)", filepath.Base(pp.PDFPath), pp.Identifier, pp.PageNumber)
				go performOcrOnPdf(pp)
			} else {
				log.Printf("ch_PerformOcr is closed but received some data")
				return
			}
		}
	}
}

func receiveOnConvertToJpg(ctx context.Context, ch <-chan PendingPage) {
	for {
		select {
		case <-ctx.Done():
			return
		case pp, ok := <-ch:
			if ok {
				log.Printf("received on ch_ConvertToJpg in receiveOnConvertToJpg page ID %v (pgNo %d)", pp.Identifier, pp.PageNumber)
				go convertPngToJpg(pp)
			}
		}
	}
}
