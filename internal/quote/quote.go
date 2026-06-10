package quote

import (
	"fmt"

	"tourtool/internal/hotel"
	"tourtool/internal/loader"
	"tourtool/internal/meal"
	"tourtool/internal/product"
	"tourtool/internal/sight"
	"tourtool/internal/transport"
)

type QuoteBreakdown struct {
	HotelCost     float64 `json:"hotel_cost"`
	SightCost     float64 `json:"sight_cost"`
	MealCost      float64 `json:"meal_cost"`
	TransportCost float64 `json:"transport_cost"`
	GuideCost     float64 `json:"guide_cost"`
	InsuranceCost float64 `json:"insurance_cost"`
	FlightCost    float64 `json:"flight_cost"`
	BaseTotal     float64 `json:"base_total"`
}

type QuoteResult struct {
	ProductID        string  `json:"product_id"`
	ProductName      string  `json:"product_name"`
	Tier             string  `json:"tier"`
	Season           product.PriceSeason `json:"season"`
	SeasonMultiplier float64 `json:"season_multiplier"`

	SinglePrice      float64 `json:"single_price"`
	DoublePrice      float64 `json:"double_price"`
	SingleRoomDiff   float64 `json:"single_room_diff"`

	HasChildPrices   bool    `json:"has_child_prices"`
	ChildNoBedPrice  float64 `json:"child_no_bed_price"`
	ChildWithBedPrice float64 `json:"child_with_bed_price"`

	HasSeniorPrices  bool    `json:"has_senior_prices"`
	SeniorPrice      float64 `json:"senior_price"`
	CompanionPrice   float64 `json:"companion_price"`

	Breakdown        QuoteBreakdown `json:"breakdown"`
	Includes         []string       `json:"includes"`
	Excludes         []string       `json:"excludes"`
}

type Calculator struct {
	idx *loader.ResourceIndex
}

func NewCalculator(idx *loader.ResourceIndex) *Calculator {
	return &Calculator{idx: idx}
}

func (c *Calculator) Calculate(p *product.Product, season product.PriceSeason) (*QuoteResult, error) {
	if p == nil {
		return nil, fmt.Errorf("product is nil")
	}

	breakdown := QuoteBreakdown{}

	hotelCost := 0.0
	sightCost := 0.0
	mealCost := 0.0

	for _, day := range p.Itinerary {
		if day.HotelTier != "" {
			h := c.idx.GetBestHotelByTier(day.HotelTier, p.Destination)
			if h != nil {
				price := hotel.GetEffectivePrice(h)
				hotelCost += price
			}
		}

		for _, sightID := range day.SightIDs {
			s := c.idx.GetBestSightByID(sightID)
			if s != nil {
				sightCost += s.Price
			}
		}

		if day.MealTier != "" {
			if day.Breakfast == "included" {
				m := c.idx.GetBestMealByTypeTier(meal.TypeBreakfast, day.MealTier, p.Destination)
				if m != nil {
					mealCost += m.PricePerPerson
				}
			}
			if day.Lunch == "included" {
				m := c.idx.GetBestMealByTypeTier(meal.TypeLunch, day.MealTier, p.Destination)
				if m != nil {
					mealCost += m.PricePerPerson
				}
			}
			if day.Dinner == "included" {
				m := c.idx.GetBestMealByTypeTier(meal.TypeDinner, day.MealTier, p.Destination)
				if m != nil {
					mealCost += m.PricePerPerson
				}
			}
		}
	}

	bus := c.idx.GetBestBusService(p.Destination)
	busCost := 0.0
	if bus != nil {
		busCost = bus.DailyPrice * float64(p.Days) / 20.0
	}

	guideCost := p.GuideDailyCost * float64(p.Days) / 15.0

	breakdown.HotelCost = hotelCost
	breakdown.SightCost = sightCost
	breakdown.MealCost = mealCost
	breakdown.TransportCost = busCost + p.FlightCost
	breakdown.GuideCost = guideCost
	breakdown.InsuranceCost = p.InsuranceCost
	breakdown.FlightCost = p.FlightCost

	baseTotal := hotelCost + sightCost + mealCost + busCost + guideCost + p.InsuranceCost + p.FlightCost
	breakdown.BaseTotal = baseTotal

	multiplier := p.GetSeasonalMultiplier(season)

	basePrice := baseTotal * multiplier

	doublePrice := basePrice
	singleRoomDiff := hotelCost * multiplier * 0.8
	singlePrice := doublePrice + singleRoomDiff

	result := &QuoteResult{
		ProductID:        p.ID,
		ProductName:      p.Name,
		Tier:             p.Tier,
		Season:           season,
		SeasonMultiplier: multiplier,
		SinglePrice:      round2(singlePrice),
		DoublePrice:      round2(doublePrice),
		SingleRoomDiff:   round2(singleRoomDiff),
		Breakdown:        breakdown,
	}

	if p.HasFamilyAudience() {
		result.HasChildPrices = true
		result.ChildNoBedPrice = round2(basePrice * 0.6)
		result.ChildWithBedPrice = round2(basePrice * 0.8)
	}

	if p.HasSeniorAudience() {
		result.HasSeniorPrices = true
		result.SeniorPrice = round2(basePrice * 0.9)
		result.CompanionPrice = round2(basePrice * 0.85)
	}

	result.Includes = []string{
		"flight_train",
		"accommodation",
		"meals",
		"first_gate_ticket",
		"tour_bus",
		"guide_service",
		"insurance",
	}

	result.Excludes = []string{
		"self_paid_items",
		"single_room_supplement",
		"excess_baggage",
		"personal_expenses",
	}

	return result, nil
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func GetHotelDetails(hotelID string, idx *loader.ResourceIndex) *hotel.Hotel {
	return idx.GetBestHotelByID(hotelID)
}

func GetSightDetails(sightID string, idx *loader.ResourceIndex) *sight.Sight {
	return idx.GetBestSightByID(sightID)
}

func GetBusDetails(busID string, idx *loader.ResourceIndex) *transport.BusService {
	services, ok := idx.BusServicesByID[busID]
	if !ok || len(services) == 0 {
		return nil
	}
	return transport.CompareBusServices(services)
}
