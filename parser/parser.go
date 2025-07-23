package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time" // Added for time.Time and parsing
)

// ReportData represents the structured data extracted from the top output,
// organized into individual snapshots.
type ReportData struct {
	Snapshots []Snapshot
}

// Metadata holds parsed system metrics from a snapshot
type Metadata struct {
	ThreadsTotal    int
	ThreadsRunning  int
	ThreadsSleeping int
	ThreadsStopped  int
	ThreadsZombie   int
	CPUUser         float64
	CPUSystem       float64
	CPUIdle         float64
	CPUWait         float64
	CPUSteal        float64
	MemTotal        float64
	MemFree         float64
	MemUsed         float64
	MemBuffCache    float64
	SwapTotal       float64
	SwapFree        float64
	SwapUsed        float64
	LoadAvg1        float64
	LoadAvg5        float64
	LoadAvg15       float64
	Uptime          string
	Users           int
}

// Snapshot holds data for a single 'top' output snapshot.
type Snapshot struct {
	Time      time.Time // Timestamp of the snapshot
	Metadata  Metadata
	Processes []ProcessData
}

// ProcessData holds information about a single process.
type ProcessData struct {
	PID     int
	User    string
	PR      int
	NI      int
	VIRT    string
	RES     string
	SHR     string
	S       string
	CPU     float64
	MEM     float64
	TIME    string
	Command string
}

// ParseTopOutput parses the raw top output and returns structured data.
func ParseTopOutput(data []byte) (ReportData, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	var reportData ReportData
	reportData.Snapshots = make([]Snapshot, 0) // Initialize slice of snapshots

	var currentSnapshotIdx = -1 // Index of the current snapshot being processed


	// Read line by line
	var lineNumber int
	for {
		lineNumber++
		line, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				break // End of file
			}
			return ReportData{}, fmt.Errorf("error reading line %d: %v", lineNumber, err)
		}
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Regex to extract time from "top -" line
		topTimeRegex := regexp.MustCompile(`^top - (\d{2}:\d{2}:\d{2})`)

		isNewSnapshotLine := strings.HasPrefix(line, "top - ")
		isGeneralMetadataLine := strings.HasPrefix(line, "Threads:") || strings.HasPrefix(line, "%Cpu(s):") || strings.HasPrefix(line, "MiB Mem :") || strings.HasPrefix(line, "MiB Swap:")

		if isNewSnapshotLine {
			// Start a new snapshot
			newSnapshot := Snapshot{
				Metadata:  Metadata{},
				Processes: make([]ProcessData, 0),
			}

			// Extract and parse the time from the "top -" line
			timeMatch := topTimeRegex.FindStringSubmatch(line)
			if len(timeMatch) > 1 {
				parsedTime, err := time.Parse("15:04:05", timeMatch[1])
				if err != nil {
					log.Printf("Line %d: Error parsing time '%s': %v", lineNumber, timeMatch[1], err)
					// Continue without time, or handle error more strictly if needed
				} else {
					newSnapshot.Time = parsedTime
				}
			}

			reportData.Snapshots = append(reportData.Snapshots, newSnapshot)
			currentSnapshotIdx = len(reportData.Snapshots) - 1
			// Parse this line's metadata to the *new* currentSnapshot
			if err := parseMetadata(line, &reportData.Snapshots[currentSnapshotIdx].Metadata); err != nil {
				log.Printf("Line %d: Error parsing metadata: %v", lineNumber, err)
			}
			continue
		}

		// Ensure we have a snapshot to add data to.
		// This handles cases where a file might start with non-"top -" metadata or process lines.
		if currentSnapshotIdx == -1 {
			newSnapshot := Snapshot{
				Metadata:  Metadata{},
				Processes: make([]ProcessData, 0),
			}
			reportData.Snapshots = append(reportData.Snapshots, newSnapshot)
			currentSnapshotIdx = 0
		}

		// Get a reference to the current snapshot to modify its contents
		currentSnapshot := &reportData.Snapshots[currentSnapshotIdx]

		// Check for other metadata lines
		if isGeneralMetadataLine {
			if err := parseMetadata(line, &currentSnapshot.Metadata); err != nil {
				log.Printf("Line %d: Error parsing metadata: %v", lineNumber, err)
			}
			continue
		}

		// Check for header line (PID USER ...) and skip
        if strings.HasPrefix(line, "PID") {
            continue
        }

        // Attempt to parse as a process line using fields
        fields := strings.Fields(line)
        if len(fields) >= 12 {
            pid, err := strconv.Atoi(fields[0])
            if err != nil {
                log.Printf("Line %d: Error converting PID '%s' to int: %v", lineNumber, fields[0], err)
            } else {
                cpu, err := strconv.ParseFloat(fields[8], 64)
                if err != nil {
                    log.Printf("Line %d: Error converting CPU '%s' to float: %v", lineNumber, fields[8], err)
                } else {
                    mem, err := strconv.ParseFloat(fields[9], 64)
                    if err != nil {
                        log.Printf("Line %d: Error converting MEM '%s' to float: %v", lineNumber, fields[9], err)
                    } else {
                        process := ProcessData{
                            PID:     pid,
                            User:    fields[1],
                            PR:      parseInt(fields[2]),
                            NI:      parseInt(fields[3]),
                            VIRT:    fields[4],
                            RES:     fields[5],
                            SHR:     fields[6],
                            S:       fields[7],
                            CPU:     cpu,
                            MEM:     mem,
                            TIME:    fields[10],
                            Command: strings.Join(fields[11:], " "),
                        }
                        currentSnapshot.Processes = append(currentSnapshot.Processes, process)
                    }
                }
            }
        } else {
            log.Printf("Line %d could not be parsed as process data or unrecognized format: %s", lineNumber, line)
        }
	}

	return reportData, nil
}

// addMetadata parses a line and adds it as a key-value pair to the provided metadata map.
// It handles simple "Key: Value" formats.
func parseMetadata(line string, metadata *Metadata) error {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid metadata line format")
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	switch {
	case key == "Threads":
		_, err := fmt.Sscanf(value, "%d total, %d running, %d sleeping, %d stopped, %d zombie",
			&metadata.ThreadsTotal, &metadata.ThreadsRunning, &metadata.ThreadsSleeping,
			&metadata.ThreadsStopped, &metadata.ThreadsZombie)
		if err != nil {
			return fmt.Errorf("error parsing Threads: %v", err)
		}

	case key == "%Cpu(s)":
        var ni, hi, si float64
        _, err := fmt.Sscanf(value, "%f us, %f sy, %f ni, %f id, %f wa, %f hi, %f si, %f st",
            &metadata.CPUUser, &metadata.CPUSystem, &ni, &metadata.CPUIdle, &metadata.CPUWait, &hi, &si, &metadata.CPUSteal)
        if err != nil {
            return fmt.Errorf("error parsing CPU: %v", err)
        }

	case key == "MiB Mem":
		_, err := fmt.Sscanf(value, "%f total, %f free, %f used, %f buff/cache",
			&metadata.MemTotal, &metadata.MemFree, &metadata.MemUsed, &metadata.MemBuffCache)
		if err != nil {
			return fmt.Errorf("error parsing Memory: %v", err)
		}

	case key == "MiB Swap":
        // allow the trailing dot after "used."
        _, err := fmt.Sscanf(value, "%f total, %f free, %f used.",
            &metadata.SwapTotal, &metadata.SwapFree, &metadata.SwapUsed)
        if err != nil {
            return fmt.Errorf("error parsing Swap: %v", err)
        }

	case strings.HasPrefix(line, "top -"):
        // Parse uptime, user count, and load averages from the full top header
        // e.g. "top - 12:02:03 up  3:07,  0 users,  load average: 3.18, 1.16, 0.41"
        topRegex := regexp.MustCompile(`^top - \d{2}:\d{2}:\d{2} up\s+([^,]+),\s+(\d+)\s+users?,\s+load average:\s*([\d.]+),\s*([\d.]+),\s*([\d.]+)`)
        if m := topRegex.FindStringSubmatch(line); len(m) == 6 {
            metadata.Uptime = "up " + m[1]
            if u, err := strconv.Atoi(m[2]); err == nil {
                metadata.Users = u
            }
            if f1, err := strconv.ParseFloat(m[3], 64); err == nil {
                metadata.LoadAvg1 = f1
            }
            if f5, err := strconv.ParseFloat(m[4], 64); err == nil {
                metadata.LoadAvg5 = f5
            }
            if f15, err := strconv.ParseFloat(m[5], 64); err == nil {
                metadata.LoadAvg15 = f15
            }
        }
	}

	return nil
}

func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}
