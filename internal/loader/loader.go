package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"tourtool/internal/hotel"
	"tourtool/internal/meal"
	"tourtool/internal/sight"
	"tourtool/internal/transport"
)

type IDMapping struct {
	SupplierID string `json:"supplier_id"`
	InternalID string `json:"internal_id"`
}

type ResourceIndex struct {
	mu sync.RWMutex

	HotelsByID      map[string][]*hotel.Hotel
	HotelsByTier    map[hotel.Tier][]*hotel.Hotel
	HotelsByName    map[string][]*hotel.Hotel

	SightsByID      map[string][]*sight.Sight
	SightsByTier    map[sight.Tier][]*sight.Sight

	MealsByID       map[string][]*meal.Meal
	MealsByTypeTier map[string][]*meal.Meal

	BusServicesByID map[string][]*transport.BusService

	HotelIDMap      map[string]string
	SightIDMap      map[string]string
	MealIDMap       map[string]string
}

var index *ResourceIndex

func NewResourceIndex() *ResourceIndex {
	return &ResourceIndex{
		HotelsByID:      make(map[string][]*hotel.Hotel),
		HotelsByTier:    make(map[hotel.Tier][]*hotel.Hotel),
		HotelsByName:    make(map[string][]*hotel.Hotel),
		SightsByID:      make(map[string][]*sight.Sight),
		SightsByTier:    make(map[sight.Tier][]*sight.Sight),
		MealsByID:       make(map[string][]*meal.Meal),
		MealsByTypeTier: make(map[string][]*meal.Meal),
		BusServicesByID: make(map[string][]*transport.BusService),
		HotelIDMap:      make(map[string]string),
		SightIDMap:      make(map[string]string),
		MealIDMap:       make(map[string]string),
	}
}

func GetIndex() *ResourceIndex {
	if index == nil {
		index = NewResourceIndex()
	}
	return index
}

func LoadAll(dataDir string) (*ResourceIndex, error) {
	idx := NewResourceIndex()
	index = idx

	if err := loadMappings(filepath.Join(dataDir, "mappings"), idx); err != nil {
		return nil, fmt.Errorf("load mappings: %w", err)
	}

	if err := loadHotels(filepath.Join(dataDir, "hotels"), idx); err != nil {
		return nil, fmt.Errorf("load hotels: %w", err)
	}

	if err := loadSights(filepath.Join(dataDir, "sights"), idx); err != nil {
		return nil, fmt.Errorf("load sights: %w", err)
	}

	if err := loadMeals(filepath.Join(dataDir, "meals"), idx); err != nil {
		return nil, fmt.Errorf("load meals: %w", err)
	}

	if err := loadTransport(filepath.Join(dataDir, "transport"), idx); err != nil {
		return nil, fmt.Errorf("load transport: %w", err)
	}

	return idx, nil
}

func loadMappings(dir string, idx *ResourceIndex) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return err
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		var mappings struct {
			Hotels  []IDMapping `json:"hotels"`
			Sights  []IDMapping `json:"sights"`
			Meals   []IDMapping `json:"meals"`
		}
		if err := json.Unmarshal(data, &mappings); err != nil {
			return fmt.Errorf("parse %s: %w", f, err)
		}
		for _, m := range mappings.Hotels {
			idx.HotelIDMap[m.SupplierID] = m.InternalID
		}
		for _, m := range mappings.Sights {
			idx.SightIDMap[m.SupplierID] = m.InternalID
		}
		for _, m := range mappings.Meals {
			idx.MealIDMap[m.SupplierID] = m.InternalID
		}
	}
	return nil
}

func loadHotels(dir string, idx *ResourceIndex) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return err
	}
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		var hotels []*hotel.Hotel
		if err := json.Unmarshal(data, &hotels); err != nil {
			var single hotel.Hotel
			if err2 := json.Unmarshal(data, &single); err2 != nil {
				return fmt.Errorf("parse %s: %w", f, err)
			}
			hotels = []*hotel.Hotel{&single}
		}
		for _, h := range hotels {
			if internalID, ok := idx.HotelIDMap[h.ID]; ok {
				h.ID = internalID
			}
			idx.HotelsByID[h.ID] = append(idx.HotelsByID[h.ID], h)
			idx.HotelsByTier[h.Tier] = append(idx.HotelsByTier[h.Tier], h)
			idx.HotelsByName[h.Name] = append(idx.HotelsByName[h.Name], h)
		}
	}
	return nil
}

func loadSights(dir string, idx *ResourceIndex) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return err
	}
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		var sights []*sight.Sight
		if err := json.Unmarshal(data, &sights); err != nil {
			var single sight.Sight
			if err2 := json.Unmarshal(data, &single); err2 != nil {
				return fmt.Errorf("parse %s: %w", f, err)
			}
			sights = []*sight.Sight{&single}
		}
		for _, s := range sights {
			if internalID, ok := idx.SightIDMap[s.ID]; ok {
				s.ID = internalID
			}
			idx.SightsByID[s.ID] = append(idx.SightsByID[s.ID], s)
			idx.SightsByTier[s.Tier] = append(idx.SightsByTier[s.Tier], s)
		}
	}
	return nil
}

func loadMeals(dir string, idx *ResourceIndex) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return err
	}
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		var meals []*meal.Meal
		if err := json.Unmarshal(data, &meals); err != nil {
			var single meal.Meal
			if err2 := json.Unmarshal(data, &single); err2 != nil {
				return fmt.Errorf("parse %s: %w", f, err)
			}
			meals = []*meal.Meal{&single}
		}
		for _, m := range meals {
			if internalID, ok := idx.MealIDMap[m.ID]; ok {
				m.ID = internalID
			}
			idx.MealsByID[m.ID] = append(idx.MealsByID[m.ID], m)
			key := fmt.Sprintf("%s:%s", m.Type, m.Tier)
			idx.MealsByTypeTier[key] = append(idx.MealsByTypeTier[key], m)
		}
	}
	return nil
}

func loadTransport(dir string, idx *ResourceIndex) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return err
	}
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		var busServices []*transport.BusService
		if err := json.Unmarshal(data, &busServices); err != nil {
			var single transport.BusService
			if err2 := json.Unmarshal(data, &single); err2 != nil {
				continue
			}
			busServices = []*transport.BusService{&single}
		}
		for _, b := range busServices {
			idx.BusServicesByID[b.ID] = append(idx.BusServicesByID[b.ID], b)
		}
	}
	return nil
}

func (idx *ResourceIndex) GetBestHotelByID(id string) *hotel.Hotel {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	hotels, ok := idx.HotelsByID[id]
	if !ok {
		return nil
	}
	return hotel.CompareHotels(hotels)
}

func (idx *ResourceIndex) GetBestHotelByTier(tier hotel.Tier, destination string) *hotel.Hotel {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	hotels, ok := idx.HotelsByTier[tier]
	if !ok {
		return nil
	}
	var filtered []*hotel.Hotel
	for _, h := range hotels {
		if destination == "" || h.Destination == destination {
			filtered = append(filtered, h)
		}
	}
	return hotel.CompareHotels(filtered)
}

func (idx *ResourceIndex) GetBestSightByID(id string) *sight.Sight {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	sights, ok := idx.SightsByID[id]
	if !ok {
		return nil
	}
	best := sights[0]
	for _, s := range sights[1:] {
		if s.Price < best.Price && s.Stock > 0 {
			best = s
		}
	}
	if best.Stock <= 0 {
		return nil
	}
	return best
}

func (idx *ResourceIndex) GetBestMealByTypeTier(mealType meal.MealType, tier meal.Tier, destination string) *meal.Meal {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	key := fmt.Sprintf("%s:%s", mealType, tier)
	meals, ok := idx.MealsByTypeTier[key]
	if !ok {
		return nil
	}
	var filtered []*meal.Meal
	for _, m := range meals {
		if destination == "" || m.Destination == destination {
			filtered = append(filtered, m)
		}
	}
	return meal.CompareMeals(filtered, mealType, tier)
}

func (idx *ResourceIndex) GetBestBusService(destination string) *transport.BusService {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	var services []*transport.BusService
	for _, list := range idx.BusServicesByID {
		for _, b := range list {
			if destination == "" || b.Route == destination {
				services = append(services, b)
			}
		}
	}
	return transport.CompareBusServices(services)
}
