package sight

import "time"

type Tier string

const (
	TierEconomy   Tier = "economy"
	TierStandard  Tier = "standard"
	TierDeluxe    Tier = "deluxe"
)

type Sight struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Destination string    `json:"destination"`
	Country     string    `json:"country"`
	Description string    `json:"description"`
	Tier        Tier      `json:"tier"`
	Price       float64   `json:"price"`
	SupplierID  string    `json:"supplier_id"`
	Stock       int       `json:"stock"`
	UpdatedAt   time.Time `json:"updated_at"`
}
