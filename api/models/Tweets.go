package models

import (
	"errors"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

// Tweet struct
type Tweet struct {
	ID           uint64    `gorm:"primary_key;auto_increment" json:"id"`
	Text         string    `gorm:"size:255;not null" json:"text"`
	Lang         string    `gorm:"size:15;not null" json:"lang"`
	ReplyCount   int       `gorm:"default:0" json:"reply_count"`
	QuoteCount   int       `gorm:"default:0" json:"quote_count"`
	RetweetCount int       `gorm:"default:0" json:"retweet_count"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	Author       User      `json:"author"`
	AuthorID     uint32    `gorm:"not null" json:"author_id"`
	UpdatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// Prepare tweet
func (t *Tweet) Prepare() {
	t.Text = html.EscapeString(strings.TrimSpace(t.Text))
	t.Author = User{}
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
}

// Validate tweet
func (t *Tweet) Validate() map[string]string {

	var err error

	var errorMessages = make(map[string]string)

	if t.Text == "" {
		err = errors.New("Required Title")
		errorMessages["Required_title"] = err.Error()

	}
	if t.AuthorID < 1 {
		err = errors.New("Required Author")
		errorMessages["Required_author"] = err.Error()
	}
	return errorMessages
}

// SaveTweet ..
func (t *Tweet) SaveTweet(db *gorm.DB) (*Tweet, error) {
	var err error
	err = db.Debug().Model(&Tweet{}).Create(&t).Error
	if err != nil {
		return &Tweet{}, err
	}
	if t.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", t.AuthorID).Take(&t.Author).Error
		if err != nil {
			return &Tweet{}, err
		}
	}
	return t, nil
}

//FindAllTweets ..
func (t *Tweet) FindAllTweets(db *gorm.DB, offset int) (*[]Tweet, error) {
	var err error
	tweets := []Tweet{}
	err = db.Debug().Model(&Tweet{}).Limit(10).Offset(offset).Order("created_at desc").Find(&tweets).Error
	if err != nil {
		return &[]Tweet{}, err
	}
	if len(tweets) > 0 {
		for _, p := range tweets {
			err := db.Debug().Model(&User{}).Where("id = ?", p.AuthorID).Take(&p.Author).Error
			if err != nil {
				return &[]Tweet{}, err
			}
		}
	}
	return &tweets, nil
}

// FindTweetByID ...
func (t *Tweet) FindTweetByID(db *gorm.DB, tid uint64) (*Tweet, error) {
	var err error
	err = db.Debug().Model(&Tweet{}).Where("id = ?", tid).Take(&t).Error
	if err != nil {
		return &Tweet{}, err
	}
	if t.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", t.AuthorID).Take(&t.Author).Error
		if err != nil {
			return &Tweet{}, err
		}
	}
	return t, nil
}

// UpdateTweet ...
func (t *Tweet) UpdateTweet(db *gorm.DB) (*Tweet, error) {

	var err error

	err = db.Debug().Model(&Tweet{}).Where("id = ?", t.ID).Updates(Tweet{Text: t.Text, UpdatedAt: time.Now()}).Error
	if err != nil {
		return &Tweet{}, err
	}
	if t.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", t.AuthorID).Take(&t.Author).Error
		if err != nil {
			return &Tweet{}, err
		}
	}
	return t, nil
}

// DeleteTweet ...
func (t *Tweet) DeleteTweet(db *gorm.DB) (int64, error) {

	db = db.Debug().Model(&Tweet{}).Where("id = ?", t.ID).Take(&Tweet{}).Delete(&Tweet{})
	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

// FindUserTweets ...
func (t *Tweet) FindUserTweets(db *gorm.DB, uid uint32) (*[]Tweet, error) {

	var err error
	tweets := []Tweet{}
	err = db.Debug().Model(&Tweet{}).Where("author_id = ?", uid).Limit(100).Order("created_at desc").Find(&tweets).Error
	if err != nil {
		return &[]Tweet{}, err
	}
	if len(tweets) > 0 {
		for _, p := range tweets {
			err := db.Debug().Model(&User{}).Where("id = ?", p.AuthorID).Take(&p.Author).Error
			if err != nil {
				return &[]Tweet{}, err
			}
		}
	}
	return &tweets, nil
}
