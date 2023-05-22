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
	`context`
	`time`
)

func analyze_StartOnFullText(ctx context.Context, pp PendingPage) {
	// sends the fullText into the channels
	ch_AnalyzeCryptonyms <- pp
}

func analyzeCryptonyms(ctx context.Context, pp PendingPage) {
	ch_AnalyzeLocations <- pp
}

func analyzeLocations(ctx context.Context, pp PendingPage) {
	ch_AnalyzeGematria <- pp
}

func analyzeGematria(ctx context.Context, pp PendingPage) {
	ch_AnalyzeDictionary <- pp
}

func analyzeWordIndexer(ctx context.Context, pp PendingPage) {
	for {
		if a_b_dictionary_loaded.Load() {
			break
		}
		select {
		case <-time.After(9 * time.Second):
			continue
		case <-ctx.Done():
			return
		}
	}

	ch_CompletedPage <- pp

}
