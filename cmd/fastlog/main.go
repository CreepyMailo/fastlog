package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/CreepyMailo/fastlog/internal/stats"
	"github.com/CreepyMailo/fastlog/internal/worker"
	"github.com/spf13/cobra"
)

var (
	filePath     string
	statusFilter string
	topN         int
	workersCount int
)

func main() {
	execute()
}

func execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "fastlog",
	Short: "Быстрый анализатор логов с конкурентной обработкой",
	Long: `fastlog - утилита для анализа больших файлов логов.
Поддерживает фильтрацию по HTTP-статусу и вывод топа IP-адресов или URL.

Примеры фильтрации статусов:
  -s 404          - только статус 404
  -s 200,301,404  - несколько статусов через запятую
  -s 2xx          - все успешные статусы (200-299)
  -s 3xx          - все редиректы (300-399)
  -s 4xx          - все ошибки клиента (400-499)
  -s 5xx          - все ошибки сервера (500-599)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("файл %s не существует", filePath)
		}

		statusMatcher, err := parseStatusFilter(statusFilter)
		if err != nil {
			return fmt.Errorf("ошибка в фильтре статусов: %w", err)
		}

		start := time.Now()

		fmt.Printf("Анализ файла: %s\n", filePath)
		fmt.Printf("Фильтр по статусу: %s\n", formatStatusFilter(statusFilter))
		fmt.Printf("Количество воркеров: %d\n", workersCount)
		fmt.Println("Обработка...")

		aggregator := stats.NewAggregator()

		config := &worker.Config{
			FilePath:      filePath,
			StatusMatcher: statusMatcher,
			NumWorkers:    workersCount,
			BufferSize:    1000,
			Aggregator:    aggregator,
		}

		pool := worker.NewPool(config)
		if err := pool.Run(); err != nil {
			return fmt.Errorf("ошибка при обработке: %w", err)
		}

		topIPs := aggregator.GetTopIPs(topN)
		topURLs := aggregator.GetTopURLs(topN)

		fmt.Printf("\nАнализ завершен за %v\n", time.Since(start))
		fmt.Printf("Всего обработано строк: %d\n", aggregator.GetTotalLines())
		fmt.Printf("Строк с ошибками парсинга: %d\n", aggregator.GetErrorLines())
		fmt.Printf("Строк, соответствующих фильтру: %d\n", aggregator.GetMatchedLines())

		if len(topIPs) > 0 {
			fmt.Printf("\nТоп-%d IP-адресов:\n", topN)
			for i, item := range topIPs {
				fmt.Printf("  %d. %-15s %d запросов\n", i+1, item.Key, item.Count)
			}
		}

		if len(topURLs) > 0 {
			fmt.Printf("\nТоп-%d URL:\n", topN)
			for i, item := range topURLs {
				if len(item.Key) > 50 {
					fmt.Printf("  %d. %-50.50s... %d запросов\n", i+1, item.Key, item.Count)
				} else {
					fmt.Printf("  %d. %-50s %d запросов\n", i+1, item.Key, item.Count)
				}
			}
		}

		return nil
	},
}

type StatusMatcher func(int) bool

func parseStatusFilter(filter string) (StatusMatcher, error) {
	if filter == "" || filter == "0" {
		return func(status int) bool { return true }, nil
	}

	if strings.HasSuffix(filter, "xx") && len(filter) == 3 {
		firstDigit := filter[0:1]
		switch firstDigit {
		case "2":
			return func(status int) bool {
				return status >= 200 && status < 300
			}, nil
		case "3":
			return func(status int) bool {
				return status >= 300 && status < 400
			}, nil
		case "4":
			return func(status int) bool {
				return status >= 400 && status < 500
			}, nil
		case "5":
			return func(status int) bool {
				return status >= 500 && status < 600
			}, nil
		default:
			return nil, fmt.Errorf("неподдерживаемый диапазон: %s (поддерживаются 2xx, 3xx, 4xx, 5xx)", filter)
		}
	}

	if strings.Contains(filter, ",") {
		parts := strings.Split(filter, ",")
		statuses := make(map[int]bool)

		for _, part := range parts {
			part = strings.TrimSpace(part)
			status, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("некорректный статус в списке: %s", part)
			}
			statuses[status] = true
		}

		return func(status int) bool {
			return statuses[status]
		}, nil
	}

	singleStatus, err := strconv.Atoi(filter)
	if err != nil {
		return nil, fmt.Errorf("некорректный статус: %s (используйте число, список через запятую или 2xx/3xx/4xx/5xx)", filter)
	}

	return func(status int) bool {
		return status == singleStatus
	}, nil
}

func formatStatusFilter(rawFilter string) string {
	if rawFilter == "" || rawFilter == "0" {
		return "все статусы"
	}

	switch rawFilter {
	case "2xx":
		return "2xx (успешные запросы)"
	case "3xx":
		return "3xx (редиректы)"
	case "4xx":
		return "4xx (ошибки клиента)"
	case "5xx":
		return "5xx (ошибки сервера)"
	}

	if strings.Contains(rawFilter, ",") {
		parts := strings.Split(rawFilter, ",")
		return fmt.Sprintf("%s (%d статусов)", rawFilter, len(parts))
	}

	return rawFilter
}

func init() {
	rootCmd.Flags().StringVarP(&filePath, "file", "f", "access.log", "Путь к файлу лога")
	rootCmd.Flags().StringVarP(&statusFilter, "status", "s", "", "Фильтр по HTTP-статусу (404, 200,301,404, 2xx, 3xx, 4xx, 5xx)")
	rootCmd.Flags().IntVarP(&topN, "top", "t", 10, "Количество записей в топе")
	rootCmd.Flags().IntVarP(&workersCount, "workers", "w", runtime.NumCPU(), "Количество воркеров")

	rootCmd.Version = "2.0.0"
}
