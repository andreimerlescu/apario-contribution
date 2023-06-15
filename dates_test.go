package main

import (
	`reflect`
	`testing`
	`time`
)

func Test_extractDates(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []time.Time
	}{
		{
			name:  "Test Case 1",
			input: "The event was held on 25th June, 2023 and then again on August 3rd, 2023. Save the next date 01/12/2023.",
			expected: []time.Time{
				time.Date(2023, time.June, 25, 0, 0, 0, 0, time.UTC),
				time.Date(2023, time.August, 3, 0, 0, 0, 0, time.UTC),
				time.Date(2023, time.January, 12, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:  "Test Case 2",
			input: "His birthdate is on 14th Feb 2020, and her birthdate is on March 1st, 2019. Their anniversary is on 07/23/2020.",
			expected: []time.Time{
				time.Date(2020, time.February, 14, 0, 0, 0, 0, time.UTC),
				time.Date(2019, time.March, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2020, time.July, 23, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractDates(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, but got %v", tc.expected, result)
			}
		})
	}
}
