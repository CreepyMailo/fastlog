package worker

import (
	"bufio"
	"os"
	"sync"

	"github.com/CreepyMailo/fastlog/internal/parser"
	"github.com/CreepyMailo/fastlog/internal/stats"
)

type Config struct {
	FilePath      string
	StatusMatcher func(int) bool
	NumWorkers    int
	BufferSize    int
	Aggregator    *stats.Aggregator
}

type Pool struct {
	config *Config
	parser *parser.Parser
	lines  chan string
	wg     sync.WaitGroup
}

func NewPool(config *Config) *Pool {
	return &Pool{
		config: config,
		parser: parser.NewParser(),
		lines:  make(chan string, config.BufferSize),
	}
}

func (p *Pool) Run() error {
	file, err := os.Open(p.config.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for i := 0; i < p.config.NumWorkers; i++ {
		p.wg.Add(1)
		go p.worker()
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		p.lines <- scanner.Text()
	}
	close(p.lines)

	p.wg.Wait()

	return scanner.Err()
}

func (p *Pool) worker() {
	defer p.wg.Done()

	for line := range p.lines {
		entry := p.parser.Parse(line)

		if !entry.Valid {
			p.config.Aggregator.AddErrorLine()
			continue
		}

		p.config.Aggregator.AddLine()

		if p.config.StatusMatcher(entry.Status) {
			p.config.Aggregator.AddMatchedLine()
			p.config.Aggregator.AddIP(entry.IP)
			p.config.Aggregator.AddURL(entry.URL)
		}
	}
}
