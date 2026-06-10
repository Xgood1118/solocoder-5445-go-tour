package commands

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"tourtool/internal/exporter"
	"tourtool/internal/product"
	"tourtool/internal/quote"
	"tourtool/pkg/i18n"
)

var (
	exportProductID string
	exportSeason    string
	exportVersion   string
	exportLang      string
	exportFilename  string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export quotation to Excel",
	Long:  `Export product quotation to an Excel file with multi-language support.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		productDir := dataDir + "/products"
		p, err := product.LoadProduct(exportProductID, productDir)
		if err != nil {
			return fmt.Errorf("load product failed: %w", err)
		}

		calc := quote.NewCalculator(idx)
		q, err := calc.Calculate(p, product.PriceSeason(exportSeason))
		if err != nil {
			return fmt.Errorf("calculate quote failed: %w", err)
		}

		if err := ensureOutputDir(); err != nil {
			return err
		}

		version := exporter.VersionSales
		if exportVersion == "customer" {
			version = exporter.VersionCustomer
		}

		lang := i18n.Lang(exportLang)
		exp := exporter.NewExcelExporter(lang, version)

		filename := exportFilename
		if filename == "" {
			filename = fmt.Sprintf("%s_%s_%s.xlsx", p.ID, exportSeason, exportVersion)
		}
		outputPath := filepath.Join(outputDir, filename)

		if err := exp.Export(p, q, outputPath); err != nil {
			return fmt.Errorf("export excel failed: %w", err)
		}

		fmt.Printf("Excel exported successfully!\n")
		fmt.Printf("  File: %s\n", outputPath)
		fmt.Printf("  Language: %s\n", exportLang)
		fmt.Printf("  Version: %s\n", exportVersion)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVar(&exportProductID, "id", "", "product ID (required)")
	exportCmd.Flags().StringVar(&exportSeason, "season", "off", "price season: peak, off, holiday")
	exportCmd.Flags().StringVar(&exportVersion, "version", "sales", "export version: sales, customer")
	exportCmd.Flags().StringVar(&exportLang, "lang", "zh-CN", "language: zh-CN, en-US, ja-JP, ko-KR, th-TH")
	exportCmd.Flags().StringVar(&exportFilename, "filename", "", "output filename")

	exportCmd.MarkFlagRequired("id")
}
