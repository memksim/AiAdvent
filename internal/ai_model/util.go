package ai_model

import (
	"log"
	"os"
)

func MustReadFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("cannot read %s: %v", path, err)
	}
	return string(b)
}
