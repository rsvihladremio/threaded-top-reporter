package parser

import (
	"bytes"
	"testing"
	"time" // Added for time.Parse in snapshot time validation
)

// testCase defines the structure for parser test cases
type testCase struct {
	name                          string
	input                         string
	expectedNumSnapshots          int
	expectedTotalProcesses        int
	expectedFirstSnapshotMetadata Metadata
	expectedFirstProcess          *ProcessData
	expectedSnapshotTimes         []string
}

// Test data from the provided example
const sampleTopOutput = `top - 12:02:03 up  3:07,  0 users,  load average: 3.18, 1.16, 0.41
Threads: 262 total,   6 running, 256 sleeping,   0 stopped,   0 zombie
%Cpu(s): 85.7 us,  7.1 sy,  0.0 ni,  5.7 id,  1.4 wa,  0.0 hi,  0.0 si,  0.0 st
MiB Mem :  16008.2 total,  10953.7 free,   3713.5 used,   1341.1 buff/cache
MiB Swap:      0.0 total,      0.0 free,      0.0 used.  12032.0 avail Mem 

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
    997 dremio    20   0 7009048   3.4g  98412 R  87.5  21.9   1:36.52 C2 CompilerThre
    996 dremio    20   0 7009048   3.4g  98412 R  81.2  21.9   1:35.89 C2 CompilerThre
   5190 dremio    20   0 7009064   3.4g  98412 S  18.8  21.9   0:03.83 rbound-command1
   5715 dremio    20   0 7009064   3.4g  98412 S  18.8  21.9   0:04.49 rbound-command5
   3293 dremio    20   0 7009064   3.4g  98412 S  12.5  21.9   0:05.55 e0 - 1927b3c3-f

top - 12:02:04 up  3:07,  0 users,  load average: 3.18, 1.16, 0.41
Threads: 262 total,   2 running, 260 sleeping,   0 stopped,   0 zombie
%Cpu(s): 75.3 us,  3.2 sy,  0.0 ni, 20.4 id,  0.0 wa,  0.0 hi,  1.0 si,  0.0 st
MiB Mem :  16008.2 total,  10953.7 free,   3713.5 used,   1341.1 buff/cache
MiB Swap:      0.0 total,      0.0 free,      0.0 used.  12032.0 avail Mem 

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
    996 dremio    20   0 7008232   3.4g  98412 S  82.2  21.9   1:36.72 C2 CompilerThre
    997 dremio    20   0 7008232   3.4g  98412 R  82.2  21.9   1:37.35 C2 CompilerThre
    998 dremio    20   0 7008232   3.4g  98412 S  14.9  21.9   0:36.57 C1 CompilerThre
   5715 dremio    20   0 7008416   3.4g  98412 R   9.9  21.9   0:04.59 1927b3c3-3473-d
   3293 dremio    20   0 7008232   3.4g  98412 S   8.9  21.9   0:05.64 e0


top - 12:02:05 up  3:07,  0 users,  load average: 3.18, 1.16, 0.41
Threads: 263 total,  10 running, 253 sleeping,   0 stopped,   0 zombie
%Cpu(s): 83.3 us,  3.2 sy,  0.0 ni, 11.8 id,  0.2 wa,  0.0 hi,  1.5 si,  0.0 st
MiB Mem :  16008.2 total,  10953.7 free,   3713.5 used,   1341.1 buff/cache
MiB Swap:      0.0 total,      0.0 free,      0.0 used.  12032.0 avail Mem 

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
    996 dremio    20   0 7018380   3.4g 101336 R  67.3  22.0   1:37.40 C2 CompilerThre
    997 dremio    20   0 7018380   3.4g 101336 R  59.4  22.0   1:37.95 C2 CompilerThre
    998 dremio    20   0 7018380   3.4g 101336 R  24.8  22.0   0:36.82 C1 CompilerThre
   5732 dremio    20   0 7018380   3.4g 101336 S  13.9  22.0   0:00.93 foreman15
   5715 dremio    20   0 7018380   3.4g 101336 S  12.9  22.0   0:04.72 rbound-command5
`

func TestParseTopOutput(t *testing.T) {
	tests := []testCase{ // Use the defined testCase struct
		{
			name:                   "Sample Top Output",
			input:                  sampleTopOutput,
			expectedNumSnapshots:   3,  // 3 distinct snapshots in the sample
			expectedTotalProcesses: 15, // 5 processes per snapshot * 3 snapshots
			expectedFirstSnapshotMetadata: Metadata{
				ThreadsTotal:    262,
				ThreadsRunning:  6,
				ThreadsSleeping: 256,
				ThreadsStopped:  0,
				ThreadsZombie:   0,
				CPUUser:         85.7,
				CPUSystem:       7.1,
				CPUIdle:         5.7,
				MemTotal:        16008.2,
				MemFree:         10953.7,
				MemUsed:         3713.5,
				MemBuffCache:    1341.1,
				SwapTotal:       0.0,
				SwapFree:        0.0,
				SwapUsed:        0.0,
				LoadAvg1:        3.18,
				LoadAvg5:        1.16,
				LoadAvg15:       0.41,
				Uptime:          "up 3:07",
				Users:           0,
			},
			expectedFirstProcess: &ProcessData{
				PID:     997,
				User:    "dremio",
				PR:      20,
				NI:      0,
				VIRT:    "7009048",
				RES:     "3.4g",
				SHR:     "98412",
				S:       "R",
				CPU:     87.5,
				MEM:     21.9,
				TIME:    "1:36.52",
				Command: "C2 CompilerThre",
			},
			expectedSnapshotTimes: []string{"12:02:03", "12:02:04", "12:02:05"}, // Expected times for each snapshot
		},
		{
			name:                          "Empty Input",
			input:                         "",
			expectedNumSnapshots:          0,
			expectedTotalProcesses:        0,
			expectedFirstSnapshotMetadata: Metadata{},
			expectedFirstProcess:          nil,
			expectedSnapshotTimes:         []string{},
		},
		{
			name:                          "Only Metadata",
			input:                         "top - 12:00:00\nThreads: 100 total",
			expectedNumSnapshots:          1,
			expectedTotalProcesses:        0,
			expectedFirstSnapshotMetadata: Metadata{},
			expectedFirstProcess:          nil,
			expectedSnapshotTimes:         []string{"12:00:00"},
		},
		{
			name:                          "Missing process fields (should log and skip)",
			input:                         "PID USER PR NI VIRT RES SHR S %CPU %MEM TIME+ COMMAND\n123 user 1 0 1m 1m 1m S invalid_cpu 0.1 0:00.00 cmd",
			expectedNumSnapshots:          1, // An implicit snapshot is created for non-top- lines
			expectedTotalProcesses:        0,
			expectedFirstSnapshotMetadata: Metadata{},
			expectedFirstProcess:          nil,
			expectedSnapshotTimes:         []string{}, // No 'top -' line means no time parsed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportData, err := ParseTopOutput([]byte(tt.input)) // Changed TopData to ReportData
			if err != nil {
				t.Fatalf("ParseTopOutput() error = %v, wantErr %v", err, false)
			}

			if len(reportData.Snapshots) != tt.expectedNumSnapshots {
				t.Errorf("ParseTopOutput() got %d snapshots, want %d", len(reportData.Snapshots), tt.expectedNumSnapshots)
			}

			totalProcesses := 0
			if len(reportData.Snapshots) > 0 {
				for _, snapshot := range reportData.Snapshots {
					totalProcesses += len(snapshot.Processes)
				}
			}

			if totalProcesses != tt.expectedTotalProcesses {
				t.Errorf("ParseTopOutput() got %d total processes, want %d", totalProcesses, tt.expectedTotalProcesses)
			}

			if tt.expectedFirstProcess != nil {
				if len(reportData.Snapshots) == 0 || len(reportData.Snapshots[0].Processes) == 0 {
					t.Errorf("Expected first process but none found in the first snapshot")
				} else {
					actual := reportData.Snapshots[0].Processes[0] // Access via snapshots slice
					expected := tt.expectedFirstProcess
					if actual.PID != expected.PID ||
						actual.User != expected.User ||
						actual.PR != expected.PR ||
						actual.NI != expected.NI ||
						actual.VIRT != expected.VIRT ||
						actual.RES != expected.RES ||
						actual.SHR != expected.SHR ||
						actual.S != expected.S ||
						actual.CPU != expected.CPU ||
						actual.MEM != expected.MEM ||
						actual.TIME != expected.TIME ||
						actual.Command != expected.Command {
						t.Errorf("First process mismatch.\nGot: %+v\nWant: %+v", actual, expected)
					}
				}
			}

			// Check snapshot times if expected
			if len(tt.expectedSnapshotTimes) > 0 {
				if len(reportData.Snapshots) != len(tt.expectedSnapshotTimes) {
					t.Errorf("ParseTopOutput() got %d snapshot times, want %d", len(reportData.Snapshots), len(tt.expectedSnapshotTimes))
				} else {
					for i, expectedTimeStr := range tt.expectedSnapshotTimes {
						// Parse the expected time string into a time.Time object for comparison
						expectedTime, err := time.Parse("15:04:05", expectedTimeStr)
						if err != nil {
							t.Fatalf("Failed to parse expected time string %q: %v", expectedTimeStr, err)
						}

						actualTime := reportData.Snapshots[i].Time
						// Compare only the time parts (Hour, Minute, Second)
						if actualTime.Hour() != expectedTime.Hour() ||
							actualTime.Minute() != expectedTime.Minute() ||
							actualTime.Second() != expectedTime.Second() {
							t.Errorf("Snapshot %d time mismatch. Got: %s, Want: %s", i, actualTime.Format("15:04:05"), expectedTimeStr)
						}
					}
				}
			}

			// Check metadata values for the first snapshot
			if tt.expectedNumSnapshots > 0 {
				firstSnapshotMetadata := reportData.Snapshots[0].Metadata
				expected := tt.expectedFirstSnapshotMetadata
				actual := firstSnapshotMetadata
				if actual.ThreadsTotal != expected.ThreadsTotal ||
					actual.ThreadsRunning != expected.ThreadsRunning ||
					actual.ThreadsSleeping != expected.ThreadsSleeping ||
					actual.ThreadsStopped != expected.ThreadsStopped ||
					actual.ThreadsZombie != expected.ThreadsZombie ||
					actual.CPUUser != expected.CPUUser ||
					actual.CPUSystem != expected.CPUSystem ||
					actual.CPUIdle != expected.CPUIdle ||
					actual.MemTotal != expected.MemTotal ||
					actual.MemFree != expected.MemFree ||
					actual.MemUsed != expected.MemUsed ||
					actual.MemBuffCache != expected.MemBuffCache ||
					actual.SwapTotal != expected.SwapTotal ||
					actual.SwapFree != expected.SwapFree ||
					actual.SwapUsed != expected.SwapUsed ||
					actual.LoadAvg1 != expected.LoadAvg1 ||
					actual.LoadAvg5 != expected.LoadAvg5 ||
					actual.LoadAvg15 != expected.LoadAvg15 ||
					actual.Uptime != expected.Uptime ||
					actual.Users != expected.Users {
					t.Errorf("Metadata mismatch.\nGot: %+v\nWant: %+v", actual, expected)
				}
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"123", 123},
		{"0", 0},
		{"-45", -45},
		{"abc", 0}, // Should return 0 on error
		{"", 0},    // Should return 0 on error
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := parseInt(tt.input); got != tt.expected {
				t.Errorf("parseInt(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseLargeInput(t *testing.T) {
	// Create a large input to ensure performance and memory efficiency
	var buf bytes.Buffer
	for i := 0; i < 100; i++ { // 100 snapshots
		buf.WriteString(sampleTopOutput)
	}
	largeInput := buf.Bytes()

	reportData, err := ParseTopOutput(largeInput)
	if err != nil {
		t.Fatalf("ParseTopOutput for large input failed: %v", err)
	}

	expectedSnapshots := 300
	if len(reportData.Snapshots) != expectedSnapshots {
		t.Errorf("Expected %d snapshots for large input, but got %d", expectedSnapshots, len(reportData.Snapshots))
	}

	expectedTotalProcesses := 15 * 100 // 15 processes per snapshot * 100 snapshots
	actualTotalProcesses := 0
	for _, snapshot := range reportData.Snapshots {
		actualTotalProcesses += len(snapshot.Processes)
	}

	if actualTotalProcesses != expectedTotalProcesses {
		t.Errorf("Expected %d total processes for large input, but got %d", expectedTotalProcesses, actualTotalProcesses)
	}
	// Further assertions can be added here if specific data points are important for large input
}

// FuzzParseTopOutput can be added for more robust testing,
// but requires a separate 'fuzz' directory and specific harness.
// func FuzzParseTopOutput(f *testing.F) {
// 	f.Add([]byte(sampleTopOutput))
// 	f.Fuzz(func(t *testing.T, data []byte) {
// 		_, _ = ParseTopOutput(data)
// 	})
// }
