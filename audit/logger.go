package audit

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type AuditLog struct {
	Timestamp string                 `json:"timestamp"`
	ClientIP  string                 `json:"clientIP"`
	UserAgent string                 `json:"userAgent"`
	Method    string                 `json:"method"`
	Path      string                 `json:"path"`
	Action    string                 `json:"action"`
	Target    string                 `json:"target"`
	Status    string                 `json:"status"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type AuditLogger struct {
	file   *os.File
	mu     sync.Mutex
	enable bool
}

func NewAuditLogger(logPath string, enable bool) (*AuditLogger, error) {
	if !enable {
		return &AuditLogger{enable: false}, nil
	}
	// 确保目录存在
	if err := os.MkdirAll(dirOf(logPath), 0755); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &AuditLogger{file: file, enable: enable}, nil
}

func (a *AuditLogger) LogEvent(event AuditLog) error {
	if !a.enable {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	event.Timestamp = time.Now().Format(time.RFC3339)
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = a.file.Write(append(b, '\n'))
	return err
}

func (a *AuditLogger) Close() error {
	if a.file != nil {
		return a.file.Close()
	}
	return nil
}

func dirOf(path string) string {
	idx := len(path) - 1
	for idx >= 0 {
		if path[idx] == '/' || path[idx] == '\\' {
			break
		}
		idx--
	}
	if idx <= 0 {
		return "."
	}
	return path[:idx]
}
