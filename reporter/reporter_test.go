package reporter

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rsvihladremio/threaded-top-reporter/parser"
)

func TestGenerateReport_HappyPath(t *testing.T) {
	// Prepare temporary output directory
	dir := t.TempDir()
	out := filepath.Join(dir, "out.html")

	// Build test data with two snapshots
	times := []time.Time{
		time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2021, 1, 1, 12, 5, 0, 0, time.UTC),
	}
	var snaps []parser.Snapshot
	for _, tm := range times {
		snaps = append(snaps, parser.Snapshot{
			Time:     tm,
			Metadata: parser.Metadata{CPUUser: 1.5},
			Processes: []parser.ProcessData{
				{PID: 123, User: "test"},
			},
		})
	}
	data := parser.ReportData{Snapshots: snaps}
	title := "Test Title"
	meta := "Some metadata"

	// Run report generation
	if err := GenerateReport(data, out, title, meta); err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	// Read and validate output
	content, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	html := string(content)

	// Check for title and metadata
	if !strings.Contains(html, "<title>"+title+"</title>") {
		t.Error("missing or incorrect <title>")
	}
	if !strings.Contains(html, meta) {
		t.Error("missing metadata text")
	}

	// Check JSON arrays for times and CPUUser
	expectedTimes := `["12:00:00","12:05:00"]`
	if !strings.Contains(html, expectedTimes) {
		t.Errorf("times JSON not found; want %s", expectedTimes)
	}
	expectedCPU := fmt.Sprintf("[%g,%g]", 1.5, 1.5)
	if !strings.Contains(html, expectedCPU) {
		t.Errorf("cpuUser JSON not found; want %s", expectedCPU)
	}

	expectedSystem := "[0,0]"
	if !strings.Contains(html, expectedSystem) {
		t.Errorf("cpuSystem JSON not found; want %s", expectedSystem)
	}

	// verify our single test process shows up
	if !strings.Contains(html, "\"123-test\"") {
		t.Error("processNamesJson not found")
	}
	if !strings.Contains(html, `"data":[0,0]`) {
		t.Error("processCpuSeriesJson not found")
	}

}

func TestTemplatesAreParseable(t *testing.T) {
	if tmpl.Lookup("base.html") == nil {
		t.Error("template base.html not found")
	}
}

func TestGenerateReport_EscapesTitleAndMetadata(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.html")
	badTitle := "<script>alert(\"xss\")</script>"
	badMeta := "<b>bold</b>"
	data := parser.ReportData{Snapshots: []parser.Snapshot{}}
	if err := GenerateReport(data, out, badTitle, badMeta); err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	htmlStr := string(b)
	// Title should be escaped
	if strings.Contains(htmlStr, badTitle) {
		t.Error("title not escaped")
	}
	if !strings.Contains(htmlStr, html.EscapeString(badTitle)) {
		t.Error("escaped title not found")
	}
	// Metadata should be escaped
	if strings.Contains(htmlStr, badMeta) {
		t.Error("metadata not escaped")
	}
	if !strings.Contains(htmlStr, html.EscapeString(badMeta)) {
		t.Error("escaped metadata not found")
	}
}
