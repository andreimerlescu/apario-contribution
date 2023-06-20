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
	"log"
	"path/filepath"
)

func receiveImportedRow(ctx context.Context, ch <-chan interface{}) {
	var err error
	for {
		select {
		case <-ctx.Done():
			return
		case ird, ok := <-ch:
			if ok {
				rd, valid := ird.(ResultData)
				if !valid {
					log.Printf("not valid typecasting for ird to rd.(ResultData)")
					return
				}
				rd, err = validatePdf(ctx, rd)
				if err != nil {
					log.Printf("received error on validatePdf for rd.URL %v ; err = %v", rd.URL, err)
				} else {
					log.Printf("validated the downloaded PDF %v from URL %v, sending rd into ch_ExtractText", filepath.Base(rd.PDFPath), rd.URL)
					if ch_ExtractText.CanWrite() {
						err := ch_ExtractText.Write(rd)
						if err != nil {
							log.Printf("failed to write to ch_ExtractText channel due to error: %v", err)
							return
						}
					}
				}
			}
		}
	}
}

func receiveOnExtractTextCh(ctx context.Context, ch <-chan interface{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case ird, ok := <-ch:
			if ok {
				rd, ok := ird.(ResultData)
				if !ok {
					log.Println("failed to assert the ird from ch_ExtractText as type ResultData")
					return
				}
				log.Printf("received rd from ch_ExtractText for URL %v, running extractPlainTextFromPdf(%v)", rd.URL, rd.Identifier)
				go extractPlainTextFromPdf(ctx, rd)
			} else {
				log.Println("ch_ExtractText is closed but received some data")
				return
			}
		}
	}
}

func receiveOnExtractPagesCh(ctx context.Context, ch <-chan interface{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case ird, ok := <-ch:
			if ok {
				rd, ok := ird.(ResultData)
				if !ok {
					log.Println("ch_ExtractPages receive an ird but cannot cast it as a .(ResultData) type")
					return
				}
				log.Printf("received on ch_ExtractPages URL %v, running extractPagesFromPdf(%v)", rd.URL, rd.Identifier)
				go extractPagesFromPdf(ctx, rd)
			} else {
				log.Println("ch_ExtractPages is closed but received some data")
				return
			}
		}
	}
}

func receiveOnGeneratePngCh(ctx context.Context, ch <-chan interface{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case ipp, ok := <-ch:
			if ok {
				pp, ok := ipp.(PendingPage)
				if !ok {
					log.Println("cannot typecast ipp to .(PendingPage)")
					return
				}
				log.Printf("received on ch_GeneratePng, running convertPageToPng(%v) for ID %v (pgNo %d)", filepath.Base(pp.PDFPath), pp.Identifier, pp.PageNumber)
				go convertPageToPng(ctx, pp)
			} else {
				log.Printf("ch_GeneratePng is closed but received some data")
				return
			}
		}
	}
}

func receiveOnGenerateLightCh(ctx context.Context, ch <-chan interface{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case ipp, ok := <-ch:
			if ok {
				pp, ok := ipp.(PendingPage)
				if !ok {
					log.Println("cant typecast ipp to .(PendingPage)")
					return
				}
				log.Printf("received on ch_GenerateLight, running generateLightThumbnails(%v) for ID %v (pgNo %d)", filepath.Base(pp.PNG.Light.Original), pp.Identifier, pp.PageNumber)
				go generateLightThumbnails(ctx, pp)
			} else {
				log.Printf("ch_GenerateLight is closed but received some data")
				return
			}
		}
	}
}

func receiveOnGenerateDarkCh(ctx context.Context, ch <-chan interface{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case ipp, ok := <-ch:
			if ok {
				pp, ok := ipp.(PendingPage)
				if !ok {
					log.Println("cant typecast ipp to .(PendingPage)")
					return
				}
				log.Printf("received on ch_GenerateDark, running generateDarkThumbnails(%v) for ID %v (pgNo %d)", filepath.Base(pp.PNG.Dark.Original), pp.Identifier, pp.PageNumber)
				go generateDarkThumbnails(ctx, pp)
			} else {
				log.Printf("ch_GenerateDark is closed but received some data")
				return
			}
		}
	}
}

func receiveOnPerformOcrCh(ctx context.Context, ch <-chan interface{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case ipp, ok := <-ch:
			if ok {
				pp, ok := ipp.(PendingPage)
				if !ok {
					log.Println("cant typecast ipp to .(PendingPage)")
					return
				}
				log.Printf("received on ch_PerformOcr, running performOcrOnPdf(%v) for ID %v (pgNo %d)", filepath.Base(pp.PDFPath), pp.Identifier, pp.PageNumber)
				go performOcrOnPdf(ctx, pp)
			} else {
				log.Printf("ch_PerformOcr is closed but received some data")
				return
			}
		}
	}
}

func receiveOnConvertToJpg(ctx context.Context, ch <-chan interface{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case ipp, ok := <-ch:
			if ok {
				pp, ok := ipp.(PendingPage)
				if !ok {
					log.Println("cant typecast ipp to .(PendingPage)")
					return
				}
				log.Printf("received on ch_ConvertToJpg in receiveOnConvertToJpg page ID %v (pgNo %d)", pp.Identifier, pp.PageNumber)
				go convertPngToJpg(ctx, pp)
			}
		}
	}
}

func receiveFullTextToAnalyze(ctx context.Context, ch <-chan interface{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case ipp, ok := <-ch:
			if ok {
				pp, ok := ipp.(PendingPage)
				if !ok {
					log.Println("cant typecast ipp to .(PendingPage)")
					return
				}
				go analyze_StartOnFullText(ctx, pp)
			}
		}
	}
}

func receiveAnalyzeCryptonym(ctx context.Context, ch <-chan interface{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case ipp, ok := <-ch:
			if ok {
				pp, ok := ipp.(PendingPage)
				if !ok {
					log.Println("cant typecast ipp to .(PendingPage)")
					return
				}
				go analyzeCryptonyms(ctx, pp)
			}
		}
	}
}

func receiveAnalyzeLocations(ctx context.Context, ch <-chan interface{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case ipp, ok := <-ch:
			if ok {
				pp, ok := ipp.(PendingPage)
				if !ok {
					log.Println("cant typecast ipp to .(PendingPage)")
					return
				}
				go analyzeLocations(ctx, pp)
			}
		}
	}
}

func receiveAnalyzeGematria(ctx context.Context, ch <-chan interface{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case ipp, ok := <-ch:
			if ok {
				pp, ok := ipp.(PendingPage)
				if !ok {
					log.Println("cant typecast ipp to .(PendingPage)")
					return
				}
				go analyzeGematria(ctx, pp)
			}
		}
	}
}

func receiveAnalyzeDictionary(ctx context.Context, ch <-chan interface{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case ipp, ok := <-ch:
			if ok {
				pp, ok := ipp.(PendingPage)
				if !ok {
					log.Println("cant typecast ipp to .(PendingPage)")
					return
				}
				go analyzeWordIndexer(ctx, pp)
			}
		}
	}
}

func receiveCompletedPendingPage(ctx context.Context, ch <-chan interface{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case ipp, ok := <-ch:
			if ok {
				pp, ok := ipp.(PendingPage)
				if !ok {
					log.Println("cant typecast ipp to .(PendingPage)")
					return
				}
				go aggregatePendingPage(ctx, pp)
			}
		}
	}
}

func receiveCompiledDocument(ctx context.Context, ch <-chan interface{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case id, ok := <-ch:
			if ok {
				d, ok := id.(Document)
				if !ok {
					log.Println("cant typecast id into .(Document)")
					return
				}
				go compileDocumentSql(ctx, d)
			}
		}
	}
}
