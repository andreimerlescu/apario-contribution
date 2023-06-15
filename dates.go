package main

import (
	`log`
	"regexp"
	`strconv`
	`strings`
	`time`
)

func extractDates(in string) []time.Time {
	re1 := regexp.MustCompile(`(?i)(\d{1,2})(st|nd|rd|th)?\s(January|Jan|February|Feb|March|Mar|April|Apr|May|June|Jun|July|Jul|August|Aug|September|Sep|October|Oct|November|Nov|December|Dec),?\s(\d{2,4})`)
	re2 := regexp.MustCompile(`(?i)(January|Jan|February|Feb|March|Mar|April|Apr|May|June|Jun|July|Jul|August|Aug|September|Sep|October|Oct|November|Nov|December|Dec)\s(0\d{1}|\d{1,2})(st|nd|rd|th)?,?\s(\d{2,4})`)
	re3 := regexp.MustCompile(`(?i)(0\d{1}|\d{1,2})\/(0\d{1}|\d{1,2})\/(\d{2,4})`)
	re4 := regexp.MustCompile(`\d{4}`)
	re5 := regexp.MustCompile(`(?i)(January|Jan|February|Feb|March|Mar|April|Apr|May|June|Jun|July|Jul|August|Aug|September|Sep|October|Oct|November|Nov|December|Dec)\.?\s(\d{1,2})(st|nd|rd|th)?,?\s(\d{2,4})`)

	match1 := re1.FindAllStringSubmatch(in, -1)
	match2 := re2.FindAllStringSubmatch(in, -1)
	match3 := re3.FindAllStringSubmatch(in, -1)
	match4 := re4.FindAllStringSubmatch(in, -1)
	match5 := re5.FindAllStringSubmatch(in, -1)

	var matches [][]string
	matches = append(matches, match1...)
	matches = append(matches, match2...)
	matches = append(matches, match3...)
	matches = append(matches, match4...)
	matches = append(matches, match5...)

	var dates []time.Time

	for _, m := range matches {
		if len(m) == 5 {
			_d := m[2]
			_m := m[1]
			if m[2] == "st" || m[2] == "nd" || m[2] == "rd" || m[2] == "th" {
				_d = m[1]
				_m = m[3]
			}
			day, dayErr := strconv.Atoi(_d)
			if dayErr != nil {
				log.Printf("failed to parse the day %v inside the date %v with error %v", m[3], m, dayErr)
				continue
			}
			month := getMonthFromString(_m)
			year, yearErr := strconv.Atoi(m[4])
			if yearErr != nil {
				log.Printf("failed to parse the year %v inside the date %v with error %v", m[3], m, yearErr)
				continue
			}
			date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
			dates = append(dates, date)
		} else if len(m) == 4 {
			month := getMonthFromString(m[1])
			year, yearErr := strconv.Atoi(m[3])
			if yearErr != nil {
				log.Printf("failed to parse the year %v inside the date %v with error %v", m[3], m, yearErr)
				continue
			}
			day, dayErr := strconv.Atoi(m[2])
			if dayErr != nil {
				log.Printf("failed to parse the day %v inside the date %v with error %v", m[2], m, dayErr)
				continue
			}
			date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
			dates = append(dates, date)
		} else if len(m) == 3 {
			year, yearErr := strconv.Atoi(m[1])
			if yearErr != nil {
				log.Printf("failed to parse the year %v inside the date %v with error %v", m[3], m, yearErr)
				continue
			}
			date := time.Date(year, time.June, 1, 0, 0, 0, 0, time.UTC)
			dates = append(dates, date)
		}
	}

	return uniqueTimes(dates)
}

func getMonthFromString(monthStr string) time.Month {
	monthStr = strings.ToLower(monthStr)
	return m_months[monthStr]
}

func uniqueTimes(times []time.Time) []time.Time {
	seen := make(map[time.Time]bool)
	var unique []time.Time
	for _, t := range times {
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		if _, ok := seen[t]; !ok {
			unique = append(unique, t)
			seen[t] = true
		}
	}
	return unique
}
