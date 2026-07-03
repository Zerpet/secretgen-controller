// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package output_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"carvel.dev/secretgen-controller/internal/cli/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPrinter_PrintTable(t *testing.T) {
	var buf bytes.Buffer
	p := output.NewPrinter(output.FormatTable, &buf)

	err := p.PrintTable(
		[]string{"NAME", "AGE"},
		[][]string{
			{"my-cert", "5d"},
			{"other-cert", "2h30m"},
		},
	)
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "AGE")
	assert.Contains(t, out, "my-cert")
	assert.Contains(t, out, "other-cert")
	assert.Contains(t, out, "5d")
}

func TestPrinter_PrintObject_JSON(t *testing.T) {
	var buf bytes.Buffer
	p := output.NewPrinter(output.FormatJSON, &buf)

	obj := map[string]string{"key": "value"}
	err := p.PrintObject(obj)
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, `"key"`)
	assert.Contains(t, out, `"value"`)
}

func TestPrinter_PrintObject_YAML(t *testing.T) {
	var buf bytes.Buffer
	p := output.NewPrinter(output.FormatYAML, &buf)

	obj := map[string]string{"key": "value"}
	err := p.PrintObject(obj)
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "key: value")
}

func TestPrinter_PrintObject_TableFallsBackToYAML(t *testing.T) {
	var buf bytes.Buffer
	p := output.NewPrinter(output.FormatTable, &buf)

	obj := map[string]string{"hello": "world"}
	err := p.PrintObject(obj)
	require.NoError(t, err)

	out := buf.String()
	// table format for PrintObject falls back to YAML
	assert.Contains(t, out, "hello: world")
}

func TestPrinter_PrintDescribe(t *testing.T) {
	var buf bytes.Buffer
	p := output.NewPrinter(output.FormatTable, &buf)

	sections := []output.DescribeSection{
		{
			Fields: []output.DescribeField{
				{Key: "Name", Value: "my-cert"},
				{Key: "Namespace", Value: "default"},
			},
		},
		{
			Title: "Spec",
			Fields: []output.DescribeField{
				{Key: "Is CA", Value: "true"},
			},
		},
	}
	err := p.PrintDescribe(sections)
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "Name:")
	assert.Contains(t, out, "my-cert")
	assert.Contains(t, out, "Namespace:")
	assert.Contains(t, out, "Spec:")
	assert.Contains(t, out, "Is CA:")
	assert.Contains(t, out, "true")
}

func TestPrinter_IsTable(t *testing.T) {
	assert.True(t, output.NewPrinter(output.FormatTable, nil).IsTable())
	assert.True(t, output.NewPrinter("", nil).IsTable())
	assert.False(t, output.NewPrinter(output.FormatJSON, nil).IsTable())
	assert.False(t, output.NewPrinter(output.FormatYAML, nil).IsTable())
}

func TestFormatAge(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m"},
		{2*time.Hour + 30*time.Minute, "2h30m"},
		{48 * time.Hour, "2d"},
	}
	for _, tc := range cases {
		ts := metav1.NewTime(time.Now().Add(-tc.d))
		got := output.FormatAge(ts)
		assert.Equal(t, tc.want, got, "duration %v", tc.d)
	}

	// zero time
	assert.Equal(t, "<unknown>", output.FormatAge(metav1.Time{}))
}

func TestJoinStrings(t *testing.T) {
	assert.Equal(t, "<none>", output.JoinStrings(nil, ", "))
	assert.Equal(t, "<none>", output.JoinStrings([]string{}, ", "))
	assert.Equal(t, "a, b, c", output.JoinStrings([]string{"a", "b", "c"}, ", "))
}

func TestBoolToStr(t *testing.T) {
	assert.Equal(t, "true", output.BoolToStr(true))
	assert.Equal(t, "false", output.BoolToStr(false))
}

func TestPrinter_PrintTable_EmptyRows(t *testing.T) {
	var buf bytes.Buffer
	p := output.NewPrinter(output.FormatTable, &buf)

	err := p.PrintTable([]string{"NAME", "AGE"}, [][]string{})
	require.NoError(t, err)

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	assert.Len(t, lines, 1, "only header should be present")
	assert.Contains(t, lines[0], "NAME")
}
