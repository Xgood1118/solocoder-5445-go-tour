package product

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"tourtool/internal/hotel"
	"tourtool/internal/meal"
)

type TargetAudience string

const (
	AudienceIndividual TargetAudience = "individual"
	AudienceGroup      TargetAudience = "group"
	AudienceFamily     TargetAudience = "family"
	AudienceSenior     TargetAudience = "senior"
	AudienceOutbound   TargetAudience = "outbound"
	AudienceDomestic   TargetAudience = "domestic"
)

type PriceSeason string

const (
	SeasonPeak    PriceSeason = "peak"
	SeasonOff     PriceSeason = "off"
	SeasonHoliday PriceSeason = "holiday"
)

type PriceStrategy struct {
	Season       PriceSeason `json:"season"`
	Multiplier   float64     `json:"multiplier"`
	StartDate    string      `json:"start_date"`
	EndDate      string      `json:"end_date"`
}

type DayType string

const (
	DayDeparture DayType = "departure"
	DaySight     DayType = "sight"
	DayReturn    DayType = "return"
)

type DayPlan struct {
	DayNumber  int        `json:"day_number"`
	DayType    DayType    `json:"day_type"`
	Title      string     `json:"title"`
	HotelID    string     `json:"hotel_id,omitempty"`
	HotelTier  hotel.Tier `json:"hotel_tier,omitempty"`
	SightIDs   []string   `json:"sight_ids,omitempty"`
	Breakfast  string     `json:"breakfast,omitempty"`
	Lunch      string     `json:"lunch,omitempty"`
	Dinner     string     `json:"dinner,omitempty"`
	MealTier   meal.Tier  `json:"meal_tier,omitempty"`
	BusIncluded bool      `json:"bus_included"`
	GuideIncluded bool    `json:"guide_included"`
	Description string    `json:"description"`
}

type ChangeRecord struct {
	ID          string    `json:"id"`
	Field       string    `json:"field"`
	OldValue    string    `json:"old_value"`
	NewValue    string    `json:"new_value"`
	Operator    string    `json:"operator"`
	Reason      string    `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
}

type Product struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Destination    string           `json:"destination"`
	Country        string           `json:"country"`
	Days           int              `json:"days"`
	Nights         int              `json:"nights"`
	Tier           string           `json:"tier"`
	TargetAudience []TargetAudience `json:"target_audience"`
	Itinerary      []*DayPlan       `json:"itinerary"`
	BasePrice      float64          `json:"base_price"`
	PriceStrategies []PriceStrategy `json:"price_strategies"`
	ChangeHistory  []*ChangeRecord  `json:"change_history"`
	GuideDailyCost float64          `json:"guide_daily_cost"`
	InsuranceCost  float64          `json:"insurance_cost"`
	FlightCost     float64          `json:"flight_cost"`
	Status         string           `json:"status"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

func (p *Product) IsOutbound() bool {
	for _, a := range p.TargetAudience {
		if a == AudienceOutbound {
			return true
		}
		if a == AudienceDomestic {
			return false
		}
	}
	return false
}

func (p *Product) HasFamilyAudience() bool {
	for _, a := range p.TargetAudience {
		if a == AudienceFamily {
			return true
		}
	}
	return false
}

func (p *Product) HasSeniorAudience() bool {
	for _, a := range p.TargetAudience {
		if a == AudienceSenior {
			return true
		}
	}
	return false
}

func (p *Product) ValidateDays() error {
	isOutbound := p.IsOutbound()
	if isOutbound {
		if p.Days < 6 || p.Days > 12 {
			return fmt.Errorf("outbound product requires 6-12 days, got %d", p.Days)
		}
	} else {
		if p.Days < 3 || p.Days > 7 {
			return fmt.Errorf("domestic product requires 3-7 days, got %d", p.Days)
		}
	}
	return nil
}

func (p *Product) AddChange(field, oldVal, newVal, operator, reason string) *ChangeRecord {
	record := &ChangeRecord{
		ID:        fmt.Sprintf("chg-%d", len(p.ChangeHistory)+1),
		Field:     field,
		OldValue:  oldVal,
		NewValue:  newVal,
		Operator:  operator,
		Reason:    reason,
		Timestamp: time.Now(),
	}
	p.ChangeHistory = append(p.ChangeHistory, record)
	p.UpdatedAt = time.Now()
	return record
}

func (p *Product) GetSeasonalMultiplier(season PriceSeason) float64 {
	for _, s := range p.PriceStrategies {
		if s.Season == season {
			return s.Multiplier
		}
	}
	return 1.0
}

func SaveProduct(p *Product, dir string) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	filename := filepath.Join(dir, fmt.Sprintf("%s.json", p.ID))
	return os.WriteFile(filename, data, 0644)
}

func LoadProduct(id string, dir string) (*Product, error) {
	filename := filepath.Join(dir, fmt.Sprintf("%s.json", id))
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var p Product
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func ListProducts(dir string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, f := range files {
		id := filepath.Base(f)
		id = id[:len(id)-5]
		ids = append(ids, id)
	}
	return ids, nil
}

type BuildConfig struct {
	Name           string
	Destination    string
	Country        string
	Days           int
	Tier           string
	TargetAudience []TargetAudience
	HotelTier      hotel.Tier
	MealTier       meal.Tier
	SightIDs       []string
	GuideDailyCost float64
	InsuranceCost  float64
	FlightCost     float64
}

func BuildProduct(cfg *BuildConfig) (*Product, error) {
	nights := cfg.Days - 1
	if nights < 0 {
		nights = 0
	}

	p := &Product{
		ID:             fmt.Sprintf("prod-%d", time.Now().Unix()),
		Name:           cfg.Name,
		Destination:    cfg.Destination,
		Country:        cfg.Country,
		Days:           cfg.Days,
		Nights:         nights,
		Tier:           cfg.Tier,
		TargetAudience: cfg.TargetAudience,
		GuideDailyCost: cfg.GuideDailyCost,
		InsuranceCost:  cfg.InsuranceCost,
		FlightCost:     cfg.FlightCost,
		Status:         "active",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	itinerary := make([]*DayPlan, cfg.Days)
	for i := 0; i < cfg.Days; i++ {
		dayNum := i + 1
		dayType := DaySight
		if i == 0 {
			dayType = DayDeparture
		} else if i == cfg.Days-1 {
			dayType = DayReturn
		}

		var sightIDs []string
		if dayType == DaySight && len(cfg.SightIDs) > 0 {
			idx := i % len(cfg.SightIDs)
			sightIDs = []string{cfg.SightIDs[idx]}
		}

		hasHotel := true
		if dayType == DayReturn {
			hasHotel = false
		}

		var breakfast, lunch, dinner string
		breakfast = "included"
		if dayType == DayDeparture {
			lunch = "included"
			dinner = "included"
		} else if dayType == DaySight {
			lunch = "included"
			dinner = "included"
		} else if dayType == DayReturn {
			lunch = "included"
		}

		plan := &DayPlan{
			DayNumber:     dayNum,
			DayType:       dayType,
			Title:         fmt.Sprintf("第%d天", dayNum),
			MealTier:      cfg.MealTier,
			BusIncluded:   true,
			GuideIncluded: true,
			Breakfast:     breakfast,
			Lunch:         lunch,
			Dinner:        dinner,
			SightIDs:      sightIDs,
		}

		if hasHotel {
			plan.HotelTier = cfg.HotelTier
		}

		itinerary[i] = plan
	}
	p.Itinerary = itinerary

	if err := p.ValidateDays(); err != nil {
		return nil, err
	}

	return p, nil
}
