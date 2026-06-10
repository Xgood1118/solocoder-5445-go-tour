package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"tourtool/internal/hotel"
	"tourtool/internal/meal"
	"tourtool/internal/product"
)

var (
	buildName     string
	buildDest     string
	buildCountry  string
	buildDays     int
	buildTier     string
	buildAudience []string
	buildHotelTier string
	buildMealTier  string
	buildGuideCost float64
	buildInsurance float64
	buildFlight    float64
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a new tour product",
	Long:  `Build a new tour product with specified configuration and save to product directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		audiences := make([]product.TargetAudience, len(buildAudience))
		for i, a := range buildAudience {
			audiences[i] = product.TargetAudience(a)
		}

		cfg := &product.BuildConfig{
			Name:           buildName,
			Destination:    buildDest,
			Country:        buildCountry,
			Days:           buildDays,
			Tier:           buildTier,
			TargetAudience: audiences,
			HotelTier:      hotel.Tier(buildHotelTier),
			MealTier:       meal.Tier(buildMealTier),
			GuideDailyCost: buildGuideCost,
			InsuranceCost:  buildInsurance,
			FlightCost:     buildFlight,
		}

		p, err := product.BuildProduct(cfg)
		if err != nil {
			return fmt.Errorf("build product failed: %w", err)
		}

		if err := ensureOutputDir(); err != nil {
			return err
		}

		productDir := outputDir + "/products"
		if err := product.SaveProduct(p, productDir); err != nil {
			return fmt.Errorf("save product failed: %w", err)
		}

		fmt.Printf("Product built successfully!\n")
		fmt.Printf("  ID: %s\n", p.ID)
		fmt.Printf("  Name: %s\n", p.Name)
		fmt.Printf("  Days: %d\n", p.Days)
		fmt.Printf("  Saved to: %s/%s.json\n", productDir, p.ID)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().StringVar(&buildName, "name", "", "product name (required)")
	buildCmd.Flags().StringVar(&buildDest, "dest", "", "destination city (required)")
	buildCmd.Flags().StringVar(&buildCountry, "country", "China", "destination country")
	buildCmd.Flags().IntVar(&buildDays, "days", 5, "number of days")
	buildCmd.Flags().StringVar(&buildTier, "tier", "standard", "product tier: economy, standard, deluxe")
	buildCmd.Flags().StringSliceVar(&buildAudience, "audience", []string{"domestic", "individual"}, "target audiences: individual, group, family, senior, outbound, domestic")
	buildCmd.Flags().StringVar(&buildHotelTier, "hotel-tier", "4star", "hotel tier: 3star, 4star, 5star")
	buildCmd.Flags().StringVar(&buildMealTier, "meal-tier", "standard", "meal tier: economy, standard, deluxe")
	buildCmd.Flags().Float64Var(&buildGuideCost, "guide-cost", 300, "daily guide cost")
	buildCmd.Flags().Float64Var(&buildInsurance, "insurance", 50, "insurance cost per person")
	buildCmd.Flags().Float64Var(&buildFlight, "flight", 1500, "flight/train cost per person")

	buildCmd.MarkFlagRequired("name")
	buildCmd.MarkFlagRequired("dest")
}
