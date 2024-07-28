package renderer_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ardikabs/dpl/internal/renderer"
	"github.com/stretchr/testify/require"
)

var overrideTestData = flag.Bool("override-testdata", false, "if override the test output data.")

func TestKustomize_Render(t *testing.T) {
	inputFiles, err := filepath.Glob(filepath.Join("testdata/kustomize", "**/*.in.yaml"))
	require.NoError(t, err)

	for _, inputFile := range inputFiles {
		releaseName := filepath.Base(filepath.Dir(inputFile))
		t.Run(releaseName, func(t *testing.T) {
			kustomize := &renderer.Kustomize{}

			bytes := &bytes.Buffer{}
			opts := []renderer.RenderOption{renderer.WithCustomWriter(bytes)}

			workdir := filepath.Dir(inputFile)

			err := kustomize.Render(workdir, releaseName, &renderer.KustomizeParams{
				KustomizationRef:   filepath.Base(inputFile),
				ImageReferenceName: "main",
				ImageName:          "ghcr.io/ardikabs/etc/mockserver",
				ImageTag:           "v1.0.0",
			},
				opts...,
			)
			require.NoError(t, err)

			outputFile := strings.ReplaceAll(inputFile, ".in.yaml", ".out.yaml")

			if *overrideTestData {
				require.NoError(t, os.WriteFile(outputFile, bytes.Bytes(), 0644))
			}

			out, err := os.ReadFile(outputFile)
			require.NoError(t, err)

			require.Equal(t, string(out), bytes.String())
		})
	}
}
