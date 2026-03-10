package parser

import (
	"regexp"
	"strconv"
)

// LogEntry представляет распарсенную запись лога
type LogEntry struct {
	IP     string
	URL    string
	Status int
	Valid  bool
}

// Parser отвечает за парсинг строк лога
type Parser struct {
	pattern *regexp.Regexp
}

// NewParser создает новый парсер для combined log формата
func NewParser() *Parser {
	// Регулярное выражение для combined log формата (Apache/Nginx)
	pattern := regexp.MustCompile(`^(?P<ip>\S+) \S+ \S+ \[(?P<time>[^\]]+)\] "(?P<method>\S+) (?P<url>\S+) \S+" (?P<status>\d+) \S+`)

	return &Parser{
		pattern: pattern,
	}
}

// Parse разбирает строку лога в структуру LogEntry
func (p *Parser) Parse(line string) LogEntry {
	matches := p.pattern.FindStringSubmatch(line)
	if matches == nil {
		return LogEntry{Valid: false}
	}

	// Индексы групп в регулярном выражении
	ipIndex := p.pattern.SubexpIndex("ip")
	urlIndex := p.pattern.SubexpIndex("url")
	statusIndex := p.pattern.SubexpIndex("status")

	if ipIndex == -1 || urlIndex == -1 || statusIndex == -1 {
		return LogEntry{Valid: false}
	}

	// Парсим статус
	status, err := strconv.Atoi(matches[statusIndex])
	if err != nil {
		return LogEntry{Valid: false}
	}

	return LogEntry{
		IP:     matches[ipIndex],
		URL:    matches[urlIndex],
		Status: status,
		Valid:  true,
	}
}

// ParseNginx разбирает логи в формате Nginx (альтернативный вариант)
func (p *Parser) ParseNginx(line string) LogEntry {
	// Формат: $remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent"
	pattern := regexp.MustCompile(`^(?P<ip>\S+) - \S+ \[[^\]]+\] "(?P<method>\S+) (?P<url>\S+) \S+" (?P<status>\d+) \d+ "([^"]*)" "([^"]*)"`)

	matches := pattern.FindStringSubmatch(line)
	if matches == nil {
		return LogEntry{Valid: false}
	}

	ipIndex := pattern.SubexpIndex("ip")
	urlIndex := pattern.SubexpIndex("url")
	statusIndex := pattern.SubexpIndex("status")

	if ipIndex == -1 || urlIndex == -1 || statusIndex == -1 {
		return LogEntry{Valid: false}
	}

	status, err := strconv.Atoi(matches[statusIndex])
	if err != nil {
		return LogEntry{Valid: false}
	}

	return LogEntry{
		IP:     matches[ipIndex],
		URL:    matches[urlIndex],
		Status: status,
		Valid:  true,
	}
}
