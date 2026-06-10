package transport

import "time"

type TransportType string

const (
	TypeFlight   TransportType = "flight"
	TypeTrain    TransportType = "train"
	TypeBus      TransportType = "bus"
	TypePrivate  TransportType = "private"
)

type Tier string

const (
	TierEconomy  Tier = "economy"
	TierStandard Tier = "standard"
	TierDeluxe   Tier = "deluxe"
)

type Transport struct {
	ID            string        `json:"id"`
	Type          TransportType `json:"type"`
	Tier          Tier          `json:"tier"`
	Departure     string        `json:"departure"`
	Arrival       string        `json:"arrival"`
	DepartureTime string        `json:"departure_time"`
	ArrivalTime   string        `json:"arrival_time"`
	Price         float64       `json:"price"`
	SupplierID    string        `json:"supplier_id"`
	SupplierName  string        `json:"supplier_name"`
	Stock         int           `json:"stock"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type BusService struct {
	ID             string    `json:"id"`
	Route          string    `json:"route"`
	VehicleType    string    `json:"vehicle_type"`
	Capacity       int       `json:"capacity"`
	DailyPrice     float64   `json:"daily_price"`
	SupplierID     string    `json:"supplier_id"`
	SupplierName   string    `json:"supplier_name"`
	GuideIncluded  bool      `json:"guide_included"`
	Stock          int       `json:"stock"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func CompareBusServices(services []*BusService) *BusService {
	if len(services) == 0 {
		return nil
	}
	best := services[0]
	for _, s := range services[1:] {
		if s.DailyPrice < best.DailyPrice && s.Stock > 0 {
			best = s
		}
	}
	if best.Stock <= 0 {
		return nil
	}
	return best
}
