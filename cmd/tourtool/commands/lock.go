package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"tourtool/internal/hotel"
)

var (
	lockHotelID    string
	lockSupplierID string
	lockPrice      float64
	lockBy         string
	lockList       bool
	lockClean      bool
)

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Lock hotel price for 24 hours",
	Long:  `Lock a hotel supplier's price for 24 hours. Useful when confirming with customers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if lockList {
			locks := hotel.ListLocks()
			if len(locks) == 0 {
				fmt.Println("No active price locks.")
				return nil
			}
			fmt.Printf("Active price locks (%d):\n", len(locks))
			for _, lock := range locks {
				remaining := time.Until(lock.ExpiresAt)
				fmt.Printf("  Hotel: %s | Supplier: %s | Price: ¥%.2f | By: %s | Expires in: %.0fh\n",
					lock.HotelID, lock.SupplierID, lock.LockedPrice, lock.LockedBy, remaining.Hours())
			}
			return nil
		}

		if lockClean {
			count := hotel.CleanExpiredLocks()
			fmt.Printf("Cleaned %d expired locks.\n", count)
			return nil
		}

		if lockHotelID == "" || lockSupplierID == "" {
			return fmt.Errorf("--hotel-id and --supplier-id are required for locking")
		}

		lock, err := hotel.LockPrice(lockHotelID, lockSupplierID, lockPrice, lockBy)
		if err != nil {
			return err
		}

		fmt.Printf("Price locked successfully!\n")
		fmt.Printf("  Hotel:    %s\n", lock.HotelID)
		fmt.Printf("  Supplier: %s\n", lock.SupplierID)
		fmt.Printf("  Price:    ¥%.2f\n", lock.LockedPrice)
		fmt.Printf("  Locked by: %s\n", lock.LockedBy)
		fmt.Printf("  Expires:   %s\n", lock.ExpiresAt.Format("2006-01-02 15:04:05"))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(lockCmd)

	lockCmd.Flags().StringVar(&lockHotelID, "hotel-id", "", "hotel ID to lock")
	lockCmd.Flags().StringVar(&lockSupplierID, "supplier-id", "", "supplier ID")
	lockCmd.Flags().Float64Var(&lockPrice, "price", 0, "locked price")
	lockCmd.Flags().StringVar(&lockBy, "by", "system", "operator name")
	lockCmd.Flags().BoolVar(&lockList, "list", false, "list all active locks")
	lockCmd.Flags().BoolVar(&lockClean, "clean", false, "clean expired locks")
}
