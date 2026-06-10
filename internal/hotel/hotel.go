package hotel

import (
	"fmt"
	"sync"
	"time"
)

type Tier string

const (
	Tier3Star    Tier = "3star"
	Tier4Star    Tier = "4star"
	Tier5Star    Tier = "5star"
	TierEconomy  Tier = "economy"
	TierStandard Tier = "standard"
	TierDeluxe   Tier = "deluxe"
)

type Hotel struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Destination   string    `json:"destination"`
	Country       string    `json:"country"`
	StarRating    int       `json:"star_rating"`
	Tier          Tier      `json:"tier"`
	Address       string    `json:"address"`
	Description   string    `json:"description"`
	Accessible    bool      `json:"accessible"`
	BasePrice     float64   `json:"base_price"`
	SupplierID    string    `json:"supplier_id"`
	SupplierName  string    `json:"supplier_name"`
	RoomType      string    `json:"room_type"`
	Stock         int       `json:"stock"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PriceLock struct {
	HotelID    string    `json:"hotel_id"`
	SupplierID string    `json:"supplier_id"`
	LockedPrice float64  `json:"locked_price"`
	LockedAt   time.Time `json:"locked_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	LockedBy   string    `json:"locked_by"`
}

var (
	lockStore = make(map[string]*PriceLock)
	lockMu    sync.RWMutex
	lockTTL   = 24 * time.Hour
)

func SetLockTTL(ttl time.Duration) {
	lockTTL = ttl
}

func CompareHotels(hotels []*Hotel) *Hotel {
	if len(hotels) == 0 {
		return nil
	}
	best := hotels[0]
	for _, h := range hotels[1:] {
		if h.BasePrice < best.BasePrice && h.Stock > 0 {
			best = h
		}
	}
	if best.Stock <= 0 {
		return nil
	}
	return best
}

func CompareByTier(hotels []*Hotel, tier Tier) *Hotel {
	var filtered []*Hotel
	for _, h := range hotels {
		if h.Tier == tier && h.Stock > 0 {
			filtered = append(filtered, h)
		}
	}
	return CompareHotels(filtered)
}

func LockPrice(hotelID, supplierID string, price float64, lockedBy string) (*PriceLock, error) {
	lockMu.Lock()
	defer lockMu.Unlock()

	now := time.Now()
	lock := &PriceLock{
		HotelID:     hotelID,
		SupplierID:  supplierID,
		LockedPrice: price,
		LockedAt:    now,
		ExpiresAt:   now.Add(lockTTL),
		LockedBy:    lockedBy,
	}
	key := fmt.Sprintf("%s:%s", hotelID, supplierID)
	lockStore[key] = lock
	return lock, nil
}

func GetLockedPrice(hotelID, supplierID string) (*PriceLock, bool) {
	lockMu.RLock()
	defer lockMu.RUnlock()

	key := fmt.Sprintf("%s:%s", hotelID, supplierID)
	lock, ok := lockStore[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(lock.ExpiresAt) {
		return nil, false
	}
	return lock, true
}

func GetEffectivePrice(hotel *Hotel) float64 {
	lock, ok := GetLockedPrice(hotel.ID, hotel.SupplierID)
	if ok {
		return lock.LockedPrice
	}
	return hotel.BasePrice
}

func IsLockExpired(lock *PriceLock) bool {
	return time.Now().After(lock.ExpiresAt)
}

func ListLocks() []*PriceLock {
	lockMu.RLock()
	defer lockMu.RUnlock()

	locks := make([]*PriceLock, 0, len(lockStore))
	for _, lock := range lockStore {
		locks = append(locks, lock)
	}
	return locks
}

func CleanExpiredLocks() int {
	lockMu.Lock()
	defer lockMu.Unlock()

	count := 0
	now := time.Now()
	for key, lock := range lockStore {
		if now.After(lock.ExpiresAt) {
			delete(lockStore, key)
			count++
		}
	}
	return count
}
