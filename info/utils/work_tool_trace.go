package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/trace"
	"strings"
	"time"
)

// Простая структура для понятных результатов
type SimpleResult struct {
	Name       string
	Duration   time.Duration
	MemoryUsed string
	Goroutines int
	Issues     []string
}

// ОСНОВНАЯ СТРУКТУРА ДЛЯ UI
type AnalysisResult struct {
	Name             string        `json:"name"`
	Duration         time.Duration `json:"duration_ms"`
	DurationReadable string        `json:"duration_readable"`
	MemoryUsed       string        `json:"memory_used"`
	MemoryBytes      int64         `json:"memory_bytes"`
	Goroutines       int           `json:"goroutines"`
	Status           string        `json:"status"`
	Issues           []string      `json:"issues"`
	Severity         string        `json:"severity"`
	Timestamp        time.Time     `json:"timestamp"`
}

type AnalysisReport struct {
	Results       []AnalysisResult `json:"results"`
	Timestamp     time.Time        `json:"timestamp"`
	GoVersion     string           `json:"go_version"`
	TotalExamples int              `json:"total_examples"`
}

func SimpleAnalyze() (*AnalysisReport, error) {
	fmt.Println("ЗАПУСК ПОСЛЕДОВАТЕЛЬНОГО АНАЛИЗА")
	fmt.Println("===================================")

	// Создаем trace файл
	f, err := os.Create("sequential_trace.out")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	trace.Start(f)
	defer trace.Stop()

	// Запускаем ПОСЛЕДОВАТЕЛЬНЫЙ анализ
	results := runSequentialAnalysis()

	// Создаем отчет для UI
	report := createAnalysisReport(results)

	// Сохраняем в JSON
	if err := saveReportToJSON(report); err != nil {
		return nil, err
	}

	// Выводим понятный отчет
	printSequentialReport(results)

	fmt.Println("\n💡 Для детального анализа запустите: go tool trace sequential_trace.out")
	fmt.Println("📁 JSON отчет сохранен в: analysis_report.json")

	return report, nil
}

// 🎯 ПОСЛЕДОВАТЕЛЬНЫЙ АНАЛИЗ - КЛЮЧЕВОЕ ИЗМЕНЕНИЕ!
func runSequentialAnalysis() []AnalysisResult {
	var results []AnalysisResult

	examples := map[string]func(){
		"simpleGoroutine": simpleGoroutine,
		"withWaitGroup":   withWaitGroup,
		"channelParam":    channelParam,
		"returnChannel":   returnChannel,
		"withRange":       withRange,
		"withSelect":      withSelect,
		"withErrGroup":    withErrGroup,
		"mergeChannels":   mergeChannels,
		"runTasks":        runTasks,
	}

	fmt.Printf("🔄 Будет протестировано %d примеров...\n", len(examples))

	// ЗАПУСКАЕМ КАЖДЫЙ ТЕСТ ПО ОЧЕРЕДИ
	for name, testFunc := range examples {
		fmt.Printf("\n🔍 Тестируем: %s...\n", name)

		// 🧹 ДАЕМ ВРЕМЯ НА ЗАВЕРШЕНИЕ ПРЕДЫДУЩИХ ГОРУТИН
		time.Sleep(100 * time.Millisecond)

		// ЗАПОМИНАЕМ СОСТОЯНИЕ ДО ТЕСТА
		goroutinesBefore := runtime.NumGoroutine()
		var memBefore runtime.MemStats
		runtime.ReadMemStats(&memBefore)
		startTime := time.Now()

		// ВЫПОЛНЯЕМ ТЕСТИРУЕМУЮ ФУНКЦИЮ
		testFunc()

		// ЗАПОМИНАЕМ СОСТОЯНИЕ ПОСЛЕ ТЕСТА
		duration := time.Since(startTime)

		// ДАЕМ ВРЕМЯ НА ЗАВЕРШЕНИЕ ГОРУТИН ТЕКУЩЕГО ТЕСТА
		time.Sleep(50 * time.Millisecond)

		goroutinesAfter := runtime.NumGoroutine()
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)

		// АНАЛИЗИРУЕМ РЕЗУЛЬТАТЫ
		memoryDelta := int64(memAfter.Alloc - memBefore.Alloc)
		goroutineDelta := goroutinesAfter - goroutinesBefore
		issues := analyzePerformance(name, duration, goroutinesBefore, goroutinesAfter, memoryDelta, memBefore, memAfter)
		status, severity := determineStatusAndSeverity(issues)

		// СОЗДАЕМ РЕЗУЛЬТАТ
		result := AnalysisResult{
			Name:             name,
			Duration:         duration,
			DurationReadable: formatDuration(duration),
			MemoryUsed:       formatBytes(uint64(memoryDelta)),
			MemoryBytes:      memoryDelta,
			Goroutines:       goroutineDelta,
			Status:           status,
			Issues:           issues,
			Severity:         severity,
			Timestamp:        time.Now(),
		}

		results = append(results, result)

		// ВЫВОДИМ ПРОМЕЖУТОЧНЫЙ РЕЗУЛЬТАТ
		statusIcon := "✅"
		if status == "warning" {
			statusIcon = "⚠️"
		} else if status == "error" {
			statusIcon = "🚨"
		}
		fmt.Printf("%s %s: %s, память: %s, горутины: %+d\n",
			statusIcon, name, formatDuration(duration),
			formatBytes(uint64(memoryDelta)), goroutineDelta)
	}

	return results
}

// АНАЛИЗ ПРОИЗВОДИТЕЛЬНОСТИ (ОСТАЕТСЯ ПРЕЖНИМ)
func analyzePerformance(name string, duration time.Duration, startGoroutines, endGoroutines int, memoryDelta int64, memBefore, memAfter runtime.MemStats) []string {
	var issues []string

	// Более строгие пороги для последовательного анализа
	if duration > 500*time.Millisecond {
		issues = append(issues, fmt.Sprintf("МЕДЛЕННО: %v", duration))
	}

	// 🎯 ТЕПЕРЬ ЭТОТ ПОРОГ РАБОТАЕТ КОРРЕКТНО!
	if endGoroutines > startGoroutines+3 {
		issues = append(issues, fmt.Sprintf("УТЕЧКА: +%d горутин", endGoroutines-startGoroutines))
	}

	if memoryDelta > 5*1024*1024 { // 5MB
		issues = append(issues, fmt.Sprintf("МНОГО ПАМЯТИ: %s", formatBytes(uint64(memoryDelta))))
	}

	if memAfter.NumGC > memBefore.NumGC+2 {
		issues = append(issues, "ЧАСТЫЙ GC")
	}

	if len(issues) == 0 {
		issues = append(issues, "✅ Нет проблем")
	}

	return issues
}

// ОПРЕДЕЛЯЕМ СТАТУС И СЕРЬЕЗНОСТЬ
func determineStatusAndSeverity(issues []string) (string, string) {
	hasError := false
	hasWarning := false

	for _, issue := range issues {
		if issue != "✅ Нет проблем" {
			if strings.Contains(issue, "УТЕЧКА") || strings.Contains(issue, "МЕДЛЕННО:") {
				hasError = true
			} else {
				hasWarning = true
			}
		}
	}

	if hasError {
		return "error", "high"
	} else if hasWarning {
		return "warning", "medium"
	}
	return "success", "low"
}

// СОЗДАЕМ ОТЧЕТ ДЛЯ UI
func createAnalysisReport(results []AnalysisResult) *AnalysisReport {
	report := &AnalysisReport{
		Results:       results,
		Timestamp:     time.Now(),
		GoVersion:     runtime.Version(),
		TotalExamples: len(results),
	}

	return report
}

// СОХРАНЯЕМ В JSON
func saveReportToJSON(report *AnalysisReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("analysis_report.json", data, 0644)
}

// ВЫВОДИМ ОТЧЕТ
func printSequentialReport(results []AnalysisResult) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("📊 РЕЗУЛЬТАТЫ ПОСЛЕДОВАТЕЛЬНОГО АНАЛИЗА")
	fmt.Println(strings.Repeat("=", 50))

	success, warnings, errors := 0, 0, 0
	var totalDuration time.Duration

	for _, result := range results {
		totalDuration += result.Duration

		switch result.Status {
		case "success":
			success++
		case "warning":
			warnings++
		case "error":
			errors++
		}

		statusIcon := "✅"
		if result.Status == "warning" {
			statusIcon = "⚠️"
		} else if result.Status == "error" {
			statusIcon = "🚨"
		}

		fmt.Printf("\n%s %s\n", statusIcon, result.Name)
		fmt.Printf("   ├─ Время: %s\n", result.DurationReadable)
		fmt.Printf("   ├─ Память: %s\n", result.MemoryUsed)
		fmt.Printf("   ├─ Горутины: %+d\n", result.Goroutines)
		fmt.Printf("   └─ Проблемы: %s\n", formatIssues(result.Issues))
	}

	// СВОДКА
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("📈 СВОДКА:")
	fmt.Printf("   ├─ Всего примеров: %d\n", len(results))
	fmt.Printf("   ├─ Успешных: %d\n", success)
	fmt.Printf("   ├─ С предупреждениями: %d\n", warnings)
	fmt.Printf("   ├─ С ошибками: %d\n", errors)
	fmt.Printf("   ├─ Общее время: %s\n", formatDuration(totalDuration))
	fmt.Printf("   └─ Среднее время: %s\n", formatDuration(totalDuration/time.Duration(len(results))))
}

// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
func formatBytes(bytes uint64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	} else {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%d ns", d.Nanoseconds())
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.1f µs", float64(d.Microseconds()))
	} else if d < time.Second {
		return fmt.Sprintf("%.1f ms", float64(d.Milliseconds()))
	} else {
		return fmt.Sprintf("%.2f s", d.Seconds())
	}
}

func formatIssues(issues []string) string {
	if len(issues) == 0 {
		return "нет"
	}
	return strings.Join(issues, ", ")
}
