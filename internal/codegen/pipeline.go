package codegen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ZulferDev/backtest-go/internal/validator"
)

type Pipeline struct {
	WorkDir string
}

func NewPipeline(workDir string) *Pipeline {
	return &Pipeline{WorkDir: workDir}
}

func (p *Pipeline) Validate(code, filename string) ([]validator.ValidationError, error) {
	return validator.ValidateStrategy(filename, []byte(code))
}

func (p *Pipeline) SaveCode(code, filename string) (string, error) {
	dir := filepath.Join(p.WorkDir, "strategies")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}
	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, []byte(code), 0644)
	return path, err
}

func (p *Pipeline) Compile(path string) error {
	cmd := exec.Command("go", "build", path)
	cmd.Dir = p.WorkDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compile failed: %s", output)
	}
	return nil
}
