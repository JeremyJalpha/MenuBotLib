package menubotlib

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type NullString struct {
	sql.NullString
}

func (ns NullString) Value() string {
	if ns.Valid {
		return ns.String
	}
	return "Not set"
}

type NullBool struct {
	sql.NullBool
}

func (nb NullBool) Value() string {
	if nb.Valid {
		return fmt.Sprintf("%v", nb.Bool)
	}
	return "Not set"
}

type UserInfo struct {
	CellNumber     string
	NickName       NullString
	Email          NullString
	SocialMedia    NullString
	Consent        NullBool
	DateTimeJoined sql.NullTime
}

// NewUserInfo creates a new UserInfo object and returns it and whether the user previously existed or not.
func NewUserInfo(db *sql.DB, senderNumber string, isAutoInc bool) (UserInfo, CustomerOrder, bool) {
	var cO CustomerOrder
	uI := UserInfo{CellNumber: senderNumber}
	
	err := uI.SetUserInfoFromDB(db)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			uI.DateTimeJoined = sql.NullTime{Time: time.Now(), Valid: true}
			err := uI.InsertNewUserInfoWithOnlyCellNum(db)
			if err != nil {
				log.Println("failed to insert user: " + senderNumber + "\n" + err.Error())
			}
			return uI, cO, false
		} else {
			log.Println("Select user falied with: " + err.Error())
			return uI, cO, false
		}
	}
	err = cO.SetCurrentOrderFromDB(db, senderNumber, isAutoInc)
	if err != nil {
		log.Println("failed to set order for " + senderNumber + "\n" + err.Error())
		return uI, cO, true
	}
	return uI, cO, true
}

func (c *UserInfo) GetUserInfoAsAString() string {
	dateTimeJoined := c.DateTimeJoined.Time.Format("2006-01-02 15:04:05")

	info := fmt.Sprintf(`Date Time Joined: %s
    
Your Nickname: %s
Your Email: %s
Social: %s

Consent: %s
(_needed to store & process your personal data_)`, dateTimeJoined, c.NickName.Value(), c.Email.Value(), c.SocialMedia.Value(), c.Consent.Value())

	return info
}

// We need a general Get UserInfo function the below reflects the code not having a ORM.
// Get User Info from database
func (c *UserInfo) SetUserInfoFromDB(db *sql.DB) error {
	queryString := `SELECT nickname, email, socialmedia, consent, datetimejoined FROM userinfo WHERE cellnumber = $1`
	err := db.QueryRow(queryString, c.CellNumber).Scan(&c.NickName, &c.Email, &c.SocialMedia, &c.Consent, &c.DateTimeJoined)
	if err != nil {
		return err
	}
	return nil
}

// Insert new user into database
func (c *UserInfo) InsertNewUserInfoWithOnlyCellNum(db *sql.DB) error {
	// Prepare an SQL statement to insert a new user
	queryString := `INSERT INTO userinfo (cellnumber, datetimejoined) VALUES ($1, $2)`
	_, err := db.Exec(queryString, c.CellNumber, c.DateTimeJoined)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return fmt.Errorf("error:11, user already exists")
		}
		return err
	}
	return nil
}

// Update User field in database
func (c *UserInfo) UpdateSingularUserInfoField(db *sql.DB, updateCol, newValue string) error {
	// Prepare an SQL statement to update the field
	queryString := fmt.Sprintf(`UPDATE userinfo SET %s = $1 WHERE cellnumber = $2`, updateCol)
	_, err := db.Exec(queryString, newValue, c.CellNumber)
	if err != nil {
		return err
	}

	return nil
}
