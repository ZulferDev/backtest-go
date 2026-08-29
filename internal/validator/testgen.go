package validator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// StrategyInfo holds information about a strategy for test generation
type StrategyInfo struct {
	PackageName  string
	StrategyName string
	Filename     string
}

// ExtractStrategyInfo extracts strategy information from source code
func ExtractStrategyInfo(filename string, src []byte) (*StrategyInfo, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	info := &StrategyInfo{
		PackageName: file.Name.Name,
		Filename:    filename,
	}

	// Find strategy struct
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			// Check if this type implements Strategy interface
			// (simplistic: just check if it's a struct)
			if _, ok := typeSpec.Type.(*ast.StructType); ok {
				info.StrategyName = typeSpec.Name.Name
				return info, nil
			}
		}
	}

	return nil, fmt.Errorf("no strategy struct found")
}

// GenerateTestFile generates a test file template for a strategy
func GenerateTestFile(info *StrategyInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("package %s\n\n", info.PackageName))
	sb.WriteString("import (\n")
	sb.WriteString("\t\"testing\"\n")
	sb.WriteString("\t\"github.com/ZulferDev/backtest-go/pkg/sdk\"\n")
	sb.WriteString(")\n\n")

	// Test Init
	sb.WriteString(fmt.Sprintf("func Test%s_Init(t *testing.T) {\n", info.StrategyName))
	sb.WriteString(fmt.Sprintf("\tstrategy := &%s{}\n", info.StrategyName))
	sb.WriteString("\tctx := &mockInitContext{}\n")
	sb.WriteString("\terr := strategy.Init(ctx)\n")
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t\tt.Fatalf(\"Init() failed: %v\", err)\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	// Test OnBar
	sb.WriteString(fmt.Sprintf("func Test%s_OnBar(t *testing.T) {\n", info.StrategyName))
	sb.WriteString(fmt.Sprintf("\tstrategy := &%s{}\n", info.StrategyName))
	sb.WriteString("\tctx := &mockBarContext{}\n")
	sb.WriteString("\tbar := sdk.OHLCV{\n")
	sb.WriteString("\t\tTimestamp: 1000,\n")
	sb.WriteString("\t\tOpen: 100,\n")
	sb.WriteString("\t\tHigh: 105,\n")
	sb.WriteString("\t\tLow: 95,\n")
	sb.WriteString("\t\tClose: 102,\n")
	sb.WriteString("\t\tVolume: 1000,\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\terr := strategy.OnBar(ctx, bar)\n")
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t\tt.Fatalf(\"OnBar() failed: %v\", err)\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	// Mock contexts
	sb.WriteString("// Mock InitContext\n")
	sb.WriteString("type mockInitContext struct{}\n\n")

	sb.WriteString("// Mock BarContext\n")
	sb.WriteString("type mockBarContext struct{}\n\n")
	sb.WriteString("func (m *mockBarContext) CurrentBar() sdk.OHLCV { return sdk.OHLCV{} }\n")
	sb.WriteString("func (m *mockBarContext) History(lookback int) []sdk.OHLCV { return []sdk.OHLCV{} }\n")
	sb.WriteString("func (m *mockBarContext) HasOpenPosition() bool { return false }\n")
	sb.WriteString("func (m *mockBarContext) CurrentPosition() sdk.Position { return nil }\n")
	sb.WriteString("func (m *mockBarContext) MarketBuy(quantity float64) error { return nil }\n")
	sb.WriteString("func (m *mockBarContext) MarketSell(quantity float64) error { return nil }\n")
	sb.WriteString("func (m *mockBarContext) CloseAll() error { return nil }\n")
	sb.WriteString("func (m *mockBarContext) LogCustomMetric(key string, value float64) {}\n")

	return sb.String()
}
