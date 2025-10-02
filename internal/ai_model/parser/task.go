package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

type Task struct {
	Task     string `json:"task"`
	DateTime string `json:"dateTime"`
	Location string `json:"location"`
}

func ParseStrict(modelOutput string) (Task, error) {
	var zero Task

	s := strings.TrimSpace(modelOutput)
	if strings.HasPrefix(s, "```") {
		s = stripCodeFence(s)
	}

	if !strings.HasPrefix(s, "{") || !strings.HasSuffix(s, "}") {
		return zero, errors.New("response must be a single JSON object with no extra text")
	}

	dec := json.NewDecoder(strings.NewReader(s))
	dec.DisallowUnknownFields()

	var out Task
	if err := dec.Decode(&out); err != nil {
		return zero, fmt.Errorf("invalid JSON by schema: %w", err)
	}

	if err := ensureEOF(dec); err != nil {
		return zero, err
	}

	if out.DateTime != "" {
		if _, err := time.Parse(time.RFC3339, out.DateTime); err != nil {
			return zero, fmt.Errorf("dateTime must be RFC-3339 or empty, got %q: %w", out.DateTime, err)
		}
		// Доп. проверка наличия смещения/часового пояса в строке
		if !(strings.Contains(out.DateTime, "Z") || strings.Contains(out.DateTime, "+") || strings.Contains(out.DateTime, "-")) {
			return zero, fmt.Errorf("dateTime must include timezone offset or 'Z': %q", out.DateTime)
		}
	}

	return out, nil
}

func ensureEOF(dec *json.Decoder) error {
	for {
		_, err := dec.Token()
		if err == io.EOF {
			return nil // строго один объект
		}
		if err != nil {
			// Если тут не EOF, значит после объекта шёл мусор/ещё один JSON
			return fmt.Errorf("only one JSON object allowed (trailing content): %w", err)
		}
		// Если попали сюда — значит за объектом есть ещё токены → ошибка
		return errors.New("only one JSON object allowed (trailing tokens present)")
	}
}

func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		s = s[idx+1:]
	}
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
