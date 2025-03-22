package mold

import (
	"bytes"
	"io/fs"
	"testing"
	"testing/fstest"
)

type testFile struct {
	name string
	data string
}

func createTestFS(extraFiles ...testFile) fs.FS {
	mapFS := fstest.MapFS{
		"view.html":     &fstest.MapFile{Data: []byte(`Hello, {{.Name}}!<br>{{partial "partial.html" .Location}}`)},
		"layout.html":   &fstest.MapFile{Data: []byte(`<html><body>{{render}}<br>{{partial "partial2.html" .Age}}</body></html>`)},
		"partial.html":  &fstest.MapFile{Data: []byte(`Location: {{.}}`)},
		"partial2.html": &fstest.MapFile{Data: []byte(`Age: {{.}}`)},
	}
	for _, f := range extraFiles {
		mapFS[f.name] = &fstest.MapFile{Data: []byte(f.data)}
	}
	return mapFS
}

func TestRender(t *testing.T) {
	testFS := createTestFS()

	engine := Must(New(testFS, WithLayout("layout.html")))

	data := map[string]any{
		"Name":     "John Doe",
		"Location": "Mars",
		"Age":      40,
	}

	// view
	var buf bytes.Buffer
	if err := engine.Render(&buf, "view.html", data); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	expected := "<html><body>Hello, John Doe!<br>Location: Mars<br>Age: 40</body></html>"
	if buf.String() != expected {
		t.Errorf("Render() got = %q, want %q", buf.String(), expected)
	}

}

func TestRender_LayoutNotFound(t *testing.T) {
	testFS := createTestFS()

	if _, err := New(testFS, WithLayout("missing.html")); err == nil {
		t.Errorf("New() expected error, got nil")
	}
}

func TestRender_LayoutParseError(t *testing.T) {
	testFS := createTestFS(testFile{"layout.html", "{{partial}}"})

	if _, err := New(testFS, WithLayout("layout.html")); err == nil {
		t.Errorf("New() expected error, got nil")
	}
}

func TestRender_ViewParseError(t *testing.T) {
	testFS := createTestFS(testFile{"parse.html", "{{partial}}"})

	if _, err := New(testFS); err == nil {
		t.Errorf("New() expected error, got nil")
	}
}

func TestRender_FileNotFound(t *testing.T) {
	testFS := createTestFS()

	engine, err := New(testFS)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	var buf bytes.Buffer
	if err := engine.Render(&buf, "nonexistent.html", nil); err == nil {
		t.Errorf("Render() expected error, got nil")
	}
}
