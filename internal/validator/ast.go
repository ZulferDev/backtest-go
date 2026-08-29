package validator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// ValidationError represents a code validation error
type ValidationError struct {
	Pos     token.Position
	Message string
	Rule    string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s (rule: %s)", e.Pos, e.Message, e.Rule)
}

// Validator validates Go source code for strategy safety
type Validator struct {
	fset   *token.FileSet
	errors []ValidationError
}

// NewValidator creates a new code validator
func NewValidator() *Validator {
	return &Validator{
		fset:   token.NewFileSet(),
		errors: []ValidationError{},
	}
}

// ValidateFile validates a Go source file
func (v *Validator) ValidateFile(filename string, src []byte) error {
	v.errors = []ValidationError{}

	// Parse the file
	file, err := parser.ParseFile(v.fset, filename, src, parser.AllErrors)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Run validation checks
	v.checkImports(file)
	v.checkGoroutines(file)
	v.checkSyscalls(file)
	v.checkUnsafeFunctions(file)

	if len(v.errors) > 0 {
		return fmt.Errorf("%d validation errors found", len(v.errors))
	}

	return nil
}

// GetErrors returns all validation errors
func (v *Validator) GetErrors() []ValidationError {
	return v.errors
}

// addError adds a validation error
func (v *Validator) addError(pos token.Pos, message, rule string) {
	v.errors = append(v.errors, ValidationError{
		Pos:     v.fset.Position(pos),
		Message: message,
		Rule:    rule,
	})
}

// checkImports validates import statements
func (v *Validator) checkImports(file *ast.File) {
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if isUnsafeImport(path) {
			v.addError(imp.Pos(), fmt.Sprintf("unsafe import: %s", path), "no-unsafe-imports")
		}
	}
}

// checkGoroutines detects goroutine usage
func (v *Validator) checkGoroutines(file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		if goStmt, ok := n.(*ast.GoStmt); ok {
			v.addError(goStmt.Pos(), "goroutines are not allowed in strategies", "no-goroutines")
		}
		return true
	})
}

// checkSyscalls detects direct system calls
func (v *Validator) checkSyscalls(file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check for syscall.* calls
		if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := selExpr.X.(*ast.Ident); ok {
				if ident.Name == "syscall" {
					v.addError(callExpr.Pos(), "direct syscalls are not allowed", "no-syscalls")
				}
			}
		}

		return true
	})
}

// checkUnsafeFunctions detects usage of unsafe package
func (v *Validator) checkUnsafeFunctions(file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check for unsafe.* calls
		if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := selExpr.X.(*ast.Ident); ok {
				if ident.Name == "unsafe" {
					v.addError(callExpr.Pos(), "unsafe package usage is not allowed", "no-unsafe")
				}
			}
		}

		return true
	})
}

// isUnsafeImport checks if an import path is unsafe
func isUnsafeImport(path string) bool {
	unsafeImports := []string{
		"os",
		"os/exec",
		"net",
		"net/http",
		"syscall",
		"unsafe",
		"io/ioutil",
		"plugin",
		"reflect",
	}

	for _, unsafe := range unsafeImports {
		if path == unsafe || strings.HasPrefix(path, unsafe+"/") {
			return true
		}
	}

	return false
}

// ValidateStrategy is a convenience function for validating strategy code
func ValidateStrategy(filename string, src []byte) ([]ValidationError, error) {
	v := NewValidator()
	err := v.ValidateFile(filename, src)
	return v.GetErrors(), err
}
