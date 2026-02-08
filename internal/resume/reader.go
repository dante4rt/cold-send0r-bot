package resume

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

func ReadText(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".pdf":
		return readPDF(path)
	case ".txt", ".md":
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("reading text file: %w", err)
		}
		return string(data), nil
	default:
		return "", fmt.Errorf("unsupported resume format: %s", ext)
	}
}

func readPDF(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening PDF: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	for i := 1; i <= r.NumPage(); i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}
