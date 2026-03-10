package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/CreepyMailo/fastlog/internal/stats"
	"github.com/CreepyMailo/fastlog/internal/worker"
	"github.com/spf13/cobra"
)

var (
	filePath     string
	statusFilter int
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
Поддерживает фильтрацию по HTTP-статусу и вывод топа IP-адресов или URL.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("файл %s не существует", filePath)
		}

		start := time.Now()

		fmt.Printf("Анализ файла: %s\n", filePath)
		fmt.Printf("Фильтр по статусу: %d\n", statusFilter)
		fmt.Printf("Количество воркеров: %d\n", workersCount)
		fmt.Println("Обработка...")

		aggregator := stats.NewAggregator()

		config := &worker.Config{
			FilePath:     filePath,
			StatusFilter: statusFilter,
			NumWorkers:   workersCount,
			BufferSize:   1000,
			Aggregator:   aggregator,
		}

		pool := worker.NewPool(config)
		if err := pool.Run(); err != nil {
			return fmt.Errorf("ошибка при обработке: %w", err)
		}

		topIPs := aggregator.GetTopIPs(topN)
		topURLs := aggregator.GetTopURLs(topN)

		fmt.Printf("\nАнализ завершен за %v\n", time.Since(start))
		fmt.Printf("Всего обработано строк: %d\n", aggregator.GetTotalLines())
		fmt.Printf("Строк с ошибками: %d\n", aggregator.GetErrorLines())

		fmt.Printf("\nТоп-%d IP-адресов:\n", topN)
		for i, item := range topIPs {
			fmt.Printf("  %d. %-15s %d запросов\n", i+1, item.Key, item.Count)
		}

		fmt.Printf("\nТоп-%d URL:\n", topN)
		for i, item := range topURLs {
			if len(item.Key) > 50 {
				fmt.Printf("  %d. %-50.50s... %d запросов\n", i+1, item.Key, item.Count)
			} else {
				fmt.Printf("  %d. %-50s %d запросов\n", i+1, item.Key, item.Count)
			}
		}

		return nil
	},
}

func init() {
	// Определяем флаги командной строки
	rootCmd.Flags().StringVarP(&filePath, "file", "f", "access.log", "Путь к файлу лога")
	rootCmd.Flags().IntVarP(&statusFilter, "status", "s", 0, "Фильтр по HTTP-статусу (например, 404)")
	rootCmd.Flags().IntVarP(&topN, "top", "t", 10, "Количество записей в топе")
	rootCmd.Flags().IntVarP(&workersCount, "workers", "w", runtime.NumCPU(), "Количество воркеров")

	// Добавляем информацию о версии
	rootCmd.Version = "1.0.0"
}
