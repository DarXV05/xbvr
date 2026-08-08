package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// SuspectThreshold is how many consecutive runs must recognise nothing before a
// site is called suspect. A single zero can be a block, a rate limit or a blip.
const SuspectThreshold = 3

type SiteHealth struct {
	SiteID           string    `gorm:"primary_key" json:"site_id"`
	LastRecognised   int       `json:"last_recognised"`
	LastAdded        int       `json:"last_added"`
	ConsecutiveZeros int       `json:"consecutive_zeros"`
	LastSeenAt       time.Time `json:"last_seen_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (o *SiteHealth) Save() error {
	db, _ := GetDB()
	defer db.Close()
	return db.Save(o).Error
}

func (o *SiteHealth) GetIfExist(id string) error {
	db, _ := GetDB()
	defer db.Close()
	return db.Where(&SiteHealth{SiteID: id}).First(o).Error
}

// Suspect reports whether the scraper looks broken rather than merely quiet.
func (o *SiteHealth) Suspect() bool {
	return o.ConsecutiveZeros >= SuspectThreshold
}

// RecordScrapeHealth stores the outcome of one run. Recognising zero scene links
// means the listing selector matched nothing; recognising some and adding none
// just means there was nothing new.
func RecordScrapeHealth(siteID string, recognised int, added int) {
	if siteID == "" {
		return
	}

	commonDb, err := GetCommonDB()
	if err != nil {
		return
	}

	var h SiteHealth
	commonDb.Where(&SiteHealth{SiteID: siteID}).First(&h)
	h.SiteID = siteID
	h.LastRecognised = recognised
	h.LastAdded = added
	h.UpdatedAt = time.Now()

	if recognised > 0 {
		h.ConsecutiveZeros = 0
		h.LastSeenAt = time.Now()
	} else {
		h.ConsecutiveZeros++
	}

	SaveWithRetry(commonDb, &h)
}

func SiteHealthFor(db *gorm.DB, siteID string) (SiteHealth, bool) {
	var h SiteHealth
	if db.Where(&SiteHealth{SiteID: siteID}).First(&h).RecordNotFound() {
		return h, false
	}
	return h, true
}
