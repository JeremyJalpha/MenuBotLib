package menubotlib

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

// TribeSize represents the possible tribe size values.
type TribeSize string

const (
	S_Small  TribeSize = "Small"
	S_Medium TribeSize = "Medium"
	S_Large  TribeSize = "Large"
)

type ActivityLevel string

const (
	T_Low    ActivityLevel = "Low"
	T_Medium ActivityLevel = "Medium"
	T_High   ActivityLevel = "High"
)

type Sighting struct {
	SightingID      int
	CellNumber      string
	SightingAddress string
	TribeSize       TribeSize
	ActivityLevel   ActivityLevel
	CurrentActivity string
	DateTimePosted  sql.NullTime
}

func validateActivityLevel(level string) (ActivityLevel, error) {
	switch level {
	case "1", "low":
		return T_Low, nil
	case "2", "medium":
		return T_Medium, nil
	case "3", "high":
		return T_High, nil
	default:
		return "", errors.New("invalid activity level")
	}
}

func validateTribeSize(size string) (TribeSize, error) {
	switch size {
	case "1", "small":
		return S_Small, nil
	case "2", "medium":
		return S_Medium, nil
	case "3", "large":
		return S_Large, nil
	default:
		return "", errors.New("invalid tribe size")
	}
}

// InsertSighting inserts a new sighting into the database.
func insertSighting(db *sql.DB, s Sighting) error {
	// Prepare an SQL statement to insert a new sighting
	queryString := `INSERT INTO sighting (cellnumber, sightingaddress, tribeSize, activityLevel, currentactivity, datetimeposted) 
                    VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(queryString, s.CellNumber, s.SightingAddress, s.TribeSize, s.ActivityLevel, s.CurrentActivity, s.DateTimePosted)
	if err != nil {
		return fmt.Errorf("failed to insert sighting: %w", err)
	}

	return nil
}

func AddSighting(db *sql.DB, s Sighting, u UserInfo) error {

	if !(u.IsVerified.Valid && u.IsVerified.Bool) {
		return errors.New("user not verified, permission denied")
	}

	err := insertSighting(db, s)
	if err != nil {
		return fmt.Errorf("failed to insert sighting: %w", err)
	}

	return nil
}
