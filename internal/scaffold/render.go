package scaffold

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// RenderedFile is one output file produced by the scaffold render step.
type RenderedFile struct {
	// RelPath is the output path relative to the chosen output directory,
	// with the .tmpl extension stripped if present in the source.
	RelPath string
	// Content is the rendered (or verbatim-copied) file content.
	Content []byte
}

const templateExt = ".tmpl"

// RenderAll walks tmplDir, renders every *.tmpl file against ctx, and
// copies every other file verbatim. The file template.toml is always skipped.
func RenderAll(tmplDir string, ctx RenderContext) ([]RenderedFile, error) {
	var files []RenderedFile

	err := filepath.WalkDir(tmplDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(tmplDir, path)
		if err != nil {
			return err
		}

		// Never include template.toml in output.
		if rel == "template.toml" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		if strings.HasSuffix(rel, templateExt) {
			rendered, err := renderFile(rel, content, ctx)
			if err != nil {
				return fmt.Errorf("render %s: %w", rel, err)
			}
			files = append(files, RenderedFile{
				RelPath: strings.TrimSuffix(rel, templateExt),
				Content: rendered,
			})
		} else {
			files = append(files, RenderedFile{RelPath: rel, Content: content})
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func renderFile(name string, src []byte, ctx RenderContext) ([]byte, error) {
	t, err := template.New(name).Option("missingkey=error").Parse(string(src))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// WriteAll writes rendered files to outDir, creating directories as needed.
func WriteAll(outDir string, files []RenderedFile) error {
	for _, f := range files {
		dest := filepath.Join(outDir, f.RelPath)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(dest), err)
		}
		if err := os.WriteFile(dest, f.Content, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", dest, err)
		}
	}
	return nil
}
