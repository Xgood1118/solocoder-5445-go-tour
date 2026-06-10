package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"tourtool/internal/product"
)

var (
	historyProductID string
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View product change history",
	Long:  `Display the change history of a product, including price changes, hotel changes, etc.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		productDir := dataDir + "/products"
		p, err := product.LoadProduct(historyProductID, productDir)
		if err != nil {
			return fmt.Errorf("load product failed: %w", err)
		}

		if len(p.ChangeHistory) == 0 {
			fmt.Println("No change history for this product.")
			return nil
		}

		fmt.Printf("=== Change History: %s (%s) ===\n", p.Name, p.ID)
		fmt.Println()

		for i, chg := range p.ChangeHistory {
			fmt.Printf("Change #%d\n", i+1)
			fmt.Printf("  Time:     %s\n", chg.Timestamp.Format("2006-01-02 15:04:05"))
			fmt.Printf("  Field:    %s\n", chg.Field)
			fmt.Printf("  Old:      %s\n", chg.OldValue)
			fmt.Printf("  New:      %s\n", chg.NewValue)
			fmt.Printf("  Operator: %s\n", chg.Operator)
			fmt.Printf("  Reason:   %s\n", chg.Reason)
			fmt.Println()
		}

		fmt.Printf("Total changes: %d\n", len(p.ChangeHistory))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)

	historyCmd.Flags().StringVar(&historyProductID, "id", "", "product ID (required)")
	historyCmd.MarkFlagRequired("id")
}
