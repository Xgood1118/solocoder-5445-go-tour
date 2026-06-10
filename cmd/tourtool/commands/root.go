package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"tourtool/internal/loader"
	"tourtool/pkg/i18n"
)

var (
	dataDir   string
	outputDir string
	lang      string
	idx       *loader.ResourceIndex
)

var rootCmd = &cobra.Command{
	Use:   "tourtool",
	Short: "Tour product packaging and quotation CLI tool",
	Long:  `A CLI tool for travel agency to package tour products and generate quotations.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		i18n.SetLang(i18n.Lang(lang))
		var err error
		idx, err = loader.LoadAll(dataDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load resources: %v\n", err)
			idx = loader.NewResourceIndex()
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&dataDir, "data", "d", "data", "data directory path")
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "output", "output directory path")
	rootCmd.PersistentFlags().StringVarP(&lang, "lang", "l", "zh-CN", "language (zh-CN, en-US, ja-JP, ko-KR, th-TH)")
}

func ensureOutputDir() error {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return os.MkdirAll(outputDir, 0755)
	}
	return nil
}
