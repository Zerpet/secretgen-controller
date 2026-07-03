// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sigsyaml "sigs.k8s.io/yaml"
)

const (
	FormatTable = "table"
	FormatJSON  = "json"
	FormatYAML  = "yaml"
)

// Printer formats and writes resource output.
type Printer struct {
	format string
	out    io.Writer
}

// NewPrinter creates a Printer that writes to out using the given format.
func NewPrinter(format string, out io.Writer) *Printer {
	return &Printer{format: format, out: out}
}

// IsTable reports whether the output format is table.
func (p *Printer) IsTable() bool { return p.format == FormatTable || p.format == "" }

// PrintObject writes obj as JSON or YAML. When format is table it falls back to YAML.
func (p *Printer) PrintObject(obj interface{}) error {
	switch p.format {
	case FormatJSON:
		enc := json.NewEncoder(p.out)
		enc.SetIndent("", "  ")
		return enc.Encode(obj)
	default: // table or yaml → YAML
		b, err := sigsyaml.Marshal(obj)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(p.out, string(b))
		return err
	}
}

// PrintTable writes headers and rows using tab-aligned columns.
func (p *Printer) PrintTable(headers []string, rows [][]string) error {
	w := tabwriter.NewWriter(p.out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	return w.Flush()
}

// DescribeSection is a named group of key-value pairs for describe output.
type DescribeSection struct {
	Title  string
	Fields []DescribeField
}

// DescribeField is a single labeled value.
type DescribeField struct {
	Key   string
	Value string
}

// PrintDescribe writes describe-style multi-section output.
func (p *Printer) PrintDescribe(sections []DescribeSection) error {
	for i, sec := range sections {
		if sec.Title != "" {
			if i > 0 {
				fmt.Fprintln(p.out)
			}
			fmt.Fprintf(p.out, "%s:\n", sec.Title)
		}
		for _, f := range sec.Fields {
			if sec.Title != "" {
				fmt.Fprintf(p.out, "  %-28s %s\n", f.Key+":", f.Value)
			} else {
				fmt.Fprintf(p.out, "%-30s %s\n", f.Key+":", f.Value)
			}
		}
	}
	return nil
}

// FormatAge returns a human-friendly age string for a Kubernetes timestamp.
func FormatAge(t metav1.Time) string {
	if t.IsZero() {
		return "<unknown>"
	}
	d := time.Since(t.Time).Round(time.Second)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

// JoinStrings joins a slice with a separator, returning "<none>" for empty slices.
func JoinStrings(vals []string, sep string) string {
	if len(vals) == 0 {
		return "<none>"
	}
	return strings.Join(vals, sep)
}

// BoolToStr converts a bool to "true"/"false".
func BoolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
