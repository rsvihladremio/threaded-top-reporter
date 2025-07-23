package reporter

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/rsvihladremio/threaded-top-reporter/parser"
)

//go:embed templates/*.html
var templateFS embed.FS

var tmpl *template.Template

func init() {
	tmpl = template.Must(template.New("base.html").ParseFS(templateFS, "templates/*.html"))
}

type SnapshotView struct {
	Time         string
	ProcessCount int
}

type ViewModel struct {
	Title              string
	Metadata           string
	TimesJson          template.JS
	CPUUserJson        template.JS
	CPUSystemJson      template.JS
	CPUIdleJson        template.JS
	CPUWaitJson        template.JS
	CPUStealJson       template.JS
	MemTotalJson       template.JS
	MemFreeJson        template.JS
	MemUsedJson        template.JS
	MemBuffCacheJson   template.JS
	SwapTotalJson      template.JS
	SwapFreeJson       template.JS
	SwapUsedJson       template.JS
	ThreadsTotalJson   template.JS
	ThreadsRunningJson template.JS
	ThreadsSleepingJson template.JS
	ThreadsStoppedJson template.JS
	ThreadsZombieJson  template.JS
	LoadAvg1Json       template.JS
	LoadAvg5Json       template.JS
	LoadAvg15Json      template.JS
	ProcessNamesJson   template.JS
	ProcessCpuSeriesJson template.JS
	Snapshots          []SnapshotView
}

// GenerateReport generates an HTML report to outputPath using parsed data.
func GenerateReport(data parser.ReportData, outputPath, title, metadata string) (err error) {
	// build time-series and snapshot views
	var times []string
	var cpuUsers, cpuSystem, cpuIdle, cpuWait, cpuSteal []float64
	var memTotal, memFree, memUsed, memBuffCache []float64
	var swapTotal, swapFree, swapUsed []float64
	var threadsTotal, threadsRunning, threadsSleeping, threadsStopped, threadsZombie []int
	var loadAvg1, loadAvg5, loadAvg15 []float64
	var snaps []SnapshotView

	// Map to track processes by PID and store their CPU usage over time
	processMap := make(map[int]map[string][]float64)
	processNames := make(map[int]string)
	
	// First pass: collect all process names and initialize tracking
	for _, s := range data.Snapshots {
		for _, p := range s.Processes {
			if _, exists := processMap[p.PID]; !exists {
				processMap[p.PID] = make(map[string][]float64)
				processNames[p.PID] = fmt.Sprintf("%d-%s", p.PID, p.Command)
			}
		}
	}

	// Second pass: collect all data points
	for _, s := range data.Snapshots {
		t := s.Time.Format("15:04:05")
		times = append(times, t)
		
		// CPU metrics
		cpuUsers = append(cpuUsers, s.Metadata.CPUUser)
		cpuSystem = append(cpuSystem, s.Metadata.CPUSystem)
		cpuIdle = append(cpuIdle, s.Metadata.CPUIdle)
		cpuWait = append(cpuWait, s.Metadata.CPUWait)
		cpuSteal = append(cpuSteal, s.Metadata.CPUSteal)
		
		// Memory metrics
		memTotal = append(memTotal, s.Metadata.MemTotal)
		memFree = append(memFree, s.Metadata.MemFree)
		memUsed = append(memUsed, s.Metadata.MemUsed)
		memBuffCache = append(memBuffCache, s.Metadata.MemBuffCache)
		swapTotal = append(swapTotal, s.Metadata.SwapTotal)
		swapFree = append(swapFree, s.Metadata.SwapFree)
		swapUsed = append(swapUsed, s.Metadata.SwapUsed)
		
		// Thread state metrics
		threadsTotal = append(threadsTotal, s.Metadata.ThreadsTotal)
		threadsRunning = append(threadsRunning, s.Metadata.ThreadsRunning)
		threadsSleeping = append(threadsSleeping, s.Metadata.ThreadsSleeping)
		threadsStopped = append(threadsStopped, s.Metadata.ThreadsStopped)
		threadsZombie = append(threadsZombie, s.Metadata.ThreadsZombie)
		
		// Load average metrics
		loadAvg1 = append(loadAvg1, s.Metadata.LoadAvg1)
		loadAvg5 = append(loadAvg5, s.Metadata.LoadAvg5)
		loadAvg15 = append(loadAvg15, s.Metadata.LoadAvg15)
		
		// Process snapshot details
		snaps = append(snaps, SnapshotView{
			Time:         s.Time.Format("2006-01-02 15:04:05"),
			ProcessCount: len(s.Processes),
		})
		
		// Track per-process CPU usage
		// For each process seen in this snapshot
		for _, p := range s.Processes {
			if processData, exists := processMap[p.PID]; exists {
				processData["cpu"] = append(processData["cpu"], p.CPU)
			}
		}
		
		// Fill in zeros for processes not seen in this snapshot
	for _, processData := range processMap {
		snapIdx := len(times) - 1
		if len(processData["cpu"]) < snapIdx+1 {
			processData["cpu"] = append(processData["cpu"], 0)
		}
	}
	}

	// Generate process CPU series for ECharts
	var processNamesList []string
	var processCpuSeries []map[string]interface{}
	
	for pid, name := range processNames {
		processNamesList = append(processNamesList, name)
		
		series := map[string]interface{}{
			"name": name,
			"type": "line",
			"data": processMap[pid]["cpu"],
		}
		processCpuSeries = append(processCpuSeries, series)
	}
	
	// Marshal all data to JSON
	tj, err := json.Marshal(times)
	if err != nil {
		return fmt.Errorf("marshal times: %w", err)
	}
	
	// CPU metrics
	cuJson, err := json.Marshal(cpuUsers)
	if err != nil {
		return fmt.Errorf("marshal cpu user series: %w", err)
	}
	csJson, err := json.Marshal(cpuSystem)
	if err != nil {
		return fmt.Errorf("marshal cpu system series: %w", err)
	}
	ciJson, err := json.Marshal(cpuIdle)
	if err != nil {
		return fmt.Errorf("marshal cpu idle series: %w", err)
	}
	cwJson, err := json.Marshal(cpuWait)
	if err != nil {
		return fmt.Errorf("marshal cpu wait series: %w", err)
	}
	cstJson, err := json.Marshal(cpuSteal)
	if err != nil {
		return fmt.Errorf("marshal cpu steal series: %w", err)
	}
	
	// Memory metrics
	mtJson, err := json.Marshal(memTotal)
	if err != nil {
		return fmt.Errorf("marshal mem total series: %w", err)
	}
	mfJson, err := json.Marshal(memFree)
	if err != nil {
		return fmt.Errorf("marshal mem free series: %w", err)
	}
	muJson, err := json.Marshal(memUsed)
	if err != nil {
		return fmt.Errorf("marshal mem used series: %w", err)
	}
	mbcJson, err := json.Marshal(memBuffCache)
	if err != nil {
		return fmt.Errorf("marshal mem buff/cache series: %w", err)
	}
	stJson, err := json.Marshal(swapTotal)
	if err != nil {
		return fmt.Errorf("marshal swap total series: %w", err)
	}
	sfJson, err := json.Marshal(swapFree)
	if err != nil {
		return fmt.Errorf("marshal swap free series: %w", err)
	}
	suJson, err := json.Marshal(swapUsed)
	if err != nil {
		return fmt.Errorf("marshal swap used series: %w", err)
	}
	
	// Thread state metrics
	ttJson, err := json.Marshal(threadsTotal)
	if err != nil {
		return fmt.Errorf("marshal threads total series: %w", err)
	}
	trJson, err := json.Marshal(threadsRunning)
	if err != nil {
		return fmt.Errorf("marshal threads running series: %w", err)
	}
	tsJson, err := json.Marshal(threadsSleeping)
	if err != nil {
		return fmt.Errorf("marshal threads sleeping series: %w", err)
	}
	tstJson, err := json.Marshal(threadsStopped)
	if err != nil {
		return fmt.Errorf("marshal threads stopped series: %w", err)
	}
	tzJson, err := json.Marshal(threadsZombie)
	if err != nil {
		return fmt.Errorf("marshal threads zombie series: %w", err)
	}
	
	// Load average metrics
	la1Json, err := json.Marshal(loadAvg1)
	if err != nil {
		return fmt.Errorf("marshal load avg 1 series: %w", err)
	}
	la5Json, err := json.Marshal(loadAvg5)
	if err != nil {
		return fmt.Errorf("marshal load avg 5 series: %w", err)
	}
	la15Json, err := json.Marshal(loadAvg15)
	if err != nil {
		return fmt.Errorf("marshal load avg 15 series: %w", err)
	}
	
	// Process metrics
	pnJson, err := json.Marshal(processNamesList)
	if err != nil {
		return fmt.Errorf("marshal process names: %w", err)
	}
	pcsJson, err := json.Marshal(processCpuSeries)
	if err != nil {
		return fmt.Errorf("marshal process cpu series: %w", err)
	}

	vm := ViewModel{
		Title:              title,
		Metadata:           metadata,
		TimesJson:          template.JS(tj),
		CPUUserJson:        template.JS(cuJson),
		CPUSystemJson:      template.JS(csJson),
		CPUIdleJson:        template.JS(ciJson),
		CPUWaitJson:        template.JS(cwJson),
		CPUStealJson:       template.JS(cstJson),
		MemTotalJson:       template.JS(mtJson),
		MemFreeJson:        template.JS(mfJson),
		MemUsedJson:        template.JS(muJson),
		MemBuffCacheJson:   template.JS(mbcJson),
		SwapTotalJson:      template.JS(stJson),
		SwapFreeJson:       template.JS(sfJson),
		SwapUsedJson:       template.JS(suJson),
		ThreadsTotalJson:   template.JS(ttJson),
		ThreadsRunningJson: template.JS(trJson),
		ThreadsSleepingJson: template.JS(tsJson),
		ThreadsStoppedJson: template.JS(tstJson),
		ThreadsZombieJson:  template.JS(tzJson),
		LoadAvg1Json:       template.JS(la1Json),
		LoadAvg5Json:       template.JS(la5Json),
		LoadAvg15Json:      template.JS(la15Json),
		ProcessNamesJson:   template.JS(pnJson),
		ProcessCpuSeriesJson: template.JS(pcsJson),
		Snapshots:          snaps,
	}

	// ensure directory
    if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
        return fmt.Errorf("mkdir: %w", err)
    }
    var f *os.File
    f, err = os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close file: %w", closeErr)
		}
	}()

	if err = tmpl.ExecuteTemplate(f, "base.html", vm); err != nil {
        return fmt.Errorf("render template: %w", err)
    }

	fmt.Printf("Report written to %s\n", outputPath)
    return
}
