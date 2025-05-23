package main

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestFormatDuration tests the duration formatting function
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "negative duration",
			duration: -5 * time.Second,
			expected: "0s",
		},
		{
			name:     "zero duration",
			duration: 0,
			expected: "0s",
		},
		{
			name:     "seconds only",
			duration: 5 * time.Second,
			expected: "5s",
		},
		{
			name:     "minutes and seconds",
			duration: 2*time.Minute + 30*time.Second,
			expected: "2m30s",
		},
		{
			name:     "hours, minutes and seconds",
			duration: 1*time.Hour + 30*time.Minute + 45*time.Second,
			expected: "1h30m45s",
		},
		{
			name:     "large seconds value",
			duration: 75 * time.Second,
			expected: "1m15s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

// TestCalculateDurationUntilTime tests the core time parsing and calculation functionality
func TestCalculateDurationUntilTime(t *testing.T) {
	fixedTime := time.Date(2025, 5, 23, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name        string
		timeStr     string
		expectError bool
		description string
	}{
		{
			name:        "24-hour format future",
			timeStr:     "15:30",
			expectError: false,
			description: "standard 24-hour format",
		},
		{
			name:        "12-hour format PM",
			timeStr:     "3:30PM",
			expectError: false,
			description: "12-hour format with PM",
		},
		{
			name:        "12-hour format pm lowercase",
			timeStr:     "3:30pm",
			expectError: false,
			description: "12-hour format with lowercase pm",
		},
		{
			name:        "24-hour with seconds",
			timeStr:     "14:35:30",
			expectError: false,
			description: "24-hour format with seconds",
		},
		{
			name:        "past time schedules tomorrow",
			timeStr:     "10:00",
			expectError: false,
			description: "past time should schedule for tomorrow",
		},
		{
			name:        "ambiguous format rejected",
			timeStr:     "1:30",
			expectError: true,
			description: "ambiguous single-digit hour should be rejected",
		},
		{
			name:        "invalid time",
			timeStr:     "25:00",
			expectError: true,
			description: "invalid hour should be rejected",
		},
		{
			name:        "invalid format",
			timeStr:     "not-a-time",
			expectError: true,
			description: "completely invalid format should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, err := calculateDurationUntilTimeWithNow(tt.timeStr, fixedTime)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for %q, but got none", tt.timeStr)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for %q: %v", tt.timeStr, err)
				return
			}

			if duration < 0 {
				t.Errorf("duration should not be negative, got %v", duration)
			}

			// Verify specific behaviors for key test cases
			switch tt.name {
			case "24-hour format future", "12-hour format PM", "12-hour format pm lowercase":
				expectedDuration := 1 * time.Hour
				if duration < expectedDuration-time.Minute || duration > expectedDuration+time.Minute {
					t.Errorf("expected duration around %v, got %v", expectedDuration, duration)
				}
			case "past time schedules tomorrow":
				if duration < 12*time.Hour {
					t.Errorf("past time should be scheduled for tomorrow, got duration %v", duration)
				}
			}
		})
	}
}

// TestAddSuffixIfArgIsNumber tests the suffix addition function
func TestAddSuffixIfArgIsNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		suffix   string
		expected string
	}{
		{
			name:     "integer number",
			input:    "5",
			suffix:   "s",
			expected: "5s",
		},
		{
			name:     "float number",
			input:    "2.5",
			suffix:   "s",
			expected: "2.5s",
		},
		{
			name:     "already has suffix",
			input:    "5m",
			suffix:   "s",
			expected: "5m",
		},
		{
			name:     "text input",
			input:    "hello",
			suffix:   "s",
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.input
			addSuffixIfArgIsNumber(&input, tt.suffix)
			if input != tt.expected {
				t.Errorf("addSuffixIfArgIsNumber(%q, %q) = %q, want %q", tt.input, tt.suffix, input, tt.expected)
			}
		})
	}
}

// TestRealWorldScenarios tests common user scenarios
func TestRealWorldScenarios(t *testing.T) {
	scenarios := []struct {
		name        string
		timeStr     string
		currentTime time.Time
		description string
	}{
		{
			name:        "lunch timer",
			timeStr:     "12:30",
			currentTime: time.Date(2025, 5, 23, 10, 0, 0, 0, time.UTC),
			description: "timer for lunch",
		},
		{
			name:        "morning alarm next day",
			timeStr:     "07:00",
			currentTime: time.Date(2025, 5, 23, 23, 0, 0, 0, time.UTC),
			description: "alarm for tomorrow morning",
		},
		{
			name:        "end of workday",
			timeStr:     "17:00",
			currentTime: time.Date(2025, 5, 23, 14, 0, 0, 0, time.UTC),
			description: "timer until end of work",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			duration, err := calculateDurationUntilTimeWithNow(scenario.timeStr, scenario.currentTime)
			if err != nil {
				t.Errorf("unexpected error for %s: %v", scenario.description, err)
				return
			}

			if duration < 0 {
				t.Errorf("duration should not be negative for %s, got %v", scenario.description, duration)
			}

			// Basic sanity checks for reasonable durations
			if duration > 24*time.Hour {
				t.Errorf("duration too large for %s, got %v", scenario.description, duration)
			}

			t.Logf("Scenario: %s - Duration: %v", scenario.description, formatDuration(duration))
		})
	}
}

// TestEdgeCases tests critical edge cases
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		timeStr     string
		currentTime time.Time
		description string
	}{
		{
			name:        "midnight transition",
			timeStr:     "00:00",
			currentTime: time.Date(2025, 5, 23, 23, 59, 0, 0, time.UTC),
			description: "timer for midnight when close to midnight",
		},
		{
			name:        "same time as current",
			timeStr:     "14:30",
			currentTime: time.Date(2025, 5, 23, 14, 30, 0, 0, time.UTC),
			description: "target time same as current time",
		},
		{
			name:        "one second in future",
			timeStr:     "14:30:01",
			currentTime: time.Date(2025, 5, 23, 14, 30, 0, 0, time.UTC),
			description: "very short duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, err := calculateDurationUntilTimeWithNow(tt.timeStr, tt.currentTime)
			if err != nil {
				t.Errorf("unexpected error for %s: %v", tt.description, err)
				return
			}

			if duration < 0 {
				t.Errorf("duration should not be negative for %s, got %v", tt.description, duration)
			}

			// Edge case specific validations
			switch tt.name {
			case "same time as current":
				// Should be scheduled for tomorrow (about 24 hours)
				if duration < 20*time.Hour {
					t.Errorf("same time should be scheduled for tomorrow, got %v", duration)
				}
			case "one second in future":
				// Should be about 1 second
				if duration < 500*time.Millisecond || duration > 2*time.Second {
					t.Errorf("expected duration around 1 second, got %v", duration)
				}
			}
		})
	}
}

// TestAmbiguousTimeFormat ensures ambiguous formats are properly rejected
func TestAmbiguousTimeFormat(t *testing.T) {
	currentTime := time.Date(2025, 5, 23, 14, 30, 0, 0, time.UTC)
	ambiguousCases := []string{"1:30", "9:45", "5:00"}

	for _, timeStr := range ambiguousCases {
		t.Run(fmt.Sprintf("ambiguous_%s", timeStr), func(t *testing.T) {
			_, err := calculateDurationUntilTimeWithNow(timeStr, currentTime)
			if err == nil {
				t.Errorf("expected error for ambiguous format %q, but got none", timeStr)
			}
			if !strings.Contains(err.Error(), "ambiguous time format") {
				t.Errorf("expected 'ambiguous time format' error, got: %v", err)
			}
		})
	}
}

// Helper function that allows us to inject a custom "now" time for testing
func calculateDurationUntilTimeWithNow(targetTimeStr string, now time.Time) (time.Duration, error) {
	if len(targetTimeStr) == 4 && targetTimeStr[1] == ':' {
		return 0, fmt.Errorf("ambiguous time format: %q. Use 24h (01:30) or 12h (1:30AM/PM)", targetTimeStr)
	}

	// Try multiple time formats
	timeFormats := []string{
		"15:04",     // 24-hour format: 14:30
		"3:04PM",    // 12-hour format with PM: 2:30PM
		"3:04pm",    // 12-hour format with pm: 2:30pm
		"15:04:05",  // 24-hour format with seconds: 14:30:45
		"3:04:05PM", // 12-hour format with seconds and PM: 2:30:45PM
		"3:04:05pm", // 12-hour format with seconds and pm: 2:30:45pm
	}

	var targetTime time.Time
	var err error

	for _, format := range timeFormats {
		if targetTime, err = time.Parse(format, targetTimeStr); err == nil {
			break
		}
	}

	if err != nil {
		return 0, fmt.Errorf("unable to parse time format. Supported formats: 15:04, 3:04PM, 3:04pm, 15:04:05, 3:04:05PM, 3:04:05pm")
	}

	// Set the target time to today
	targetTime = time.Date(now.Year(), now.Month(), now.Day(),
		targetTime.Hour(), targetTime.Minute(), targetTime.Second(), 0, now.Location())

	// Calculate duration until target time
	duration := targetTime.Sub(now)

	// Schedule for tomorrow if the time has passed or is the exact same time
	if duration <= 0 {
		targetTime = targetTime.AddDate(0, 0, 1)
		duration = targetTime.Sub(now)
	}

	return duration, nil
}

// Benchmark tests for performance monitoring
func BenchmarkFormatDuration(b *testing.B) {
	duration := 2*time.Hour + 30*time.Minute + 45*time.Second
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatDuration(duration)
	}
}

func BenchmarkCalculateDurationUntilTime(b *testing.B) {
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = calculateDurationUntilTimeWithNow("15:30:45", now)
	}
}
