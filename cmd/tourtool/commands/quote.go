package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"tourtool/internal/product"
	"tourtool/internal/quote"
)

var (
	quoteProductID string
	quoteSeason    string
	quoteJSON      bool
)

var quoteCmd = &cobra.Command{
	Use:   "quote",
	Short: "Generate quotation for a product",
	Long:  `Calculate and display quotation for a given product ID.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		productDir := dataDir + "/products"
		p, err := product.LoadProduct(quoteProductID, productDir)
		if err != nil {
			return fmt.Errorf("load product failed: %w", err)
		}

		calc := quote.NewCalculator(idx)
		result, err := calc.Calculate(p, product.PriceSeason(quoteSeason))
		if err != nil {
			return fmt.Errorf("calculate quote failed: %w", err)
		}

		if quoteJSON {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			printQuote(result)
		}

		return nil
	},
}

func printQuote(q *quote.QuoteResult) {
	fmt.Printf("=== Quotation: %s ===\n", q.ProductName)
	fmt.Printf("Tier: %s | Season: %s (x%.2f)\n", q.Tier, q.Season, q.SeasonMultiplier)
	fmt.Println()
	fmt.Printf("Single Price:     ¥%.2f\n", q.SinglePrice)
	fmt.Printf("Double Price:     ¥%.2f\n", q.DoublePrice)
	fmt.Printf("Single Room Diff: ¥%.2f\n", q.SingleRoomDiff)

	if q.HasChildPrices {
		fmt.Printf("Child (no bed):   ¥%.2f\n", q.ChildNoBedPrice)
		fmt.Printf("Child (with bed): ¥%.2f\n", q.ChildWithBedPrice)
	}

	if q.HasSeniorPrices {
		fmt.Printf("Senior Price:     ¥%.2f\n", q.SeniorPrice)
		fmt.Printf("Companion Price:  ¥%.2f\n", q.CompanionPrice)
	}

	fmt.Println()
	fmt.Println("--- Breakdown ---")
	fmt.Printf("Hotel:     ¥%.2f\n", q.Breakdown.HotelCost)
	fmt.Printf("Sight:     ¥%.2f\n", q.Breakdown.SightCost)
	fmt.Printf("Meal:      ¥%.2f\n", q.Breakdown.MealCost)
	fmt.Printf("Transport: ¥%.2f\n", q.Breakdown.TransportCost)
	fmt.Printf("Guide:     ¥%.2f\n", q.Breakdown.GuideCost)
	fmt.Printf("Insurance: ¥%.2f\n", q.Breakdown.InsuranceCost)
}

func init() {
	rootCmd.AddCommand(quoteCmd)

	quoteCmd.Flags().StringVar(&quoteProductID, "id", "", "product ID (required)")
	quoteCmd.Flags().StringVar(&quoteSeason, "season", "off", "price season: peak, off, holiday")
	quoteCmd.Flags().BoolVar(&quoteJSON, "json", false, "output as JSON")

	quoteCmd.MarkFlagRequired("id")
}
