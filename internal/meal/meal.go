package meal

import "time"

type MealType string

const (
	TypeBreakfast MealType = "breakfast"
	TypeLunch     MealType = "lunch"
	TypeDinner    MealType = "dinner"
)

type Tier string

const (
	TierEconomy  Tier = "economy"
	TierStandard Tier = "standard"
	TierDeluxe   Tier = "deluxe"
)

type Meal struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Restaurant   string   `json:"restaurant"`
	Destination  string   `json:"destination"`
	Type         MealType `json:"type"`
	Tier         Tier     `json:"tier"`
	PricePerPerson float64 `json:"price_per_person"`
	SupplierID   string   `json:"supplier_id"`
	SupplierName string   `json:"supplier_name"`
	Stock        int      `json:"stock"`
	Description  string   `json:"description"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func CompareMeals(meals []*Meal, mealType MealType, tier Tier) *Meal {
	var filtered []*Meal
	for _, m := range meals {
		if m.Type == mealType && m.Tier == tier && m.Stock > 0 {
			filtered = append(filtered, m)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	best := filtered[0]
	for _, m := range filtered[1:] {
		if m.PricePerPerson < best.PricePerPerson {
			best = m
		}
	}
	return best
}
