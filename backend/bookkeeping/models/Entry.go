package models

import (
	"gorm.io/gorm"
)

type Entry struct {
	gorm.Model          // Already includes "id", "created_at", "updated_at", "deleted_at"
	JournalID   uint    `json:"journal_id" bson:"journal_id"`
	OwnerID     uint    `json:"owner_id" bson:"owner_id"`
	Description string  `json:"description" bson:"description"`
	Debit       float64 `json:"debit" bson:"debit"`
	Credit      float64 `json:"credit" bson:"credit"`
	Balance     float64 `json:"balance" bson:"balance"`
}

// Saves an entry to the database
func (entry *Entry) Save() (*Entry, error) {
	err := Database.Model(&entry).Create(&entry).Error
	if err != nil {
		return &Entry{}, err
	}
	return entry, nil
}

func FetchAllEntries(userId uint) ([]Entry, error) {
	var entries []Entry
	err := Database.Model(&Entry{}).Where("owner_id = ?", userId).Find(&entries).Error
	if err != nil {
		return []Entry{}, err
	}
	return entries, nil
}

// Fetches all entries from the database
func FetchEntriesByJournalId(userId uint, journalId uint) ([]Entry, error) {
	var entries []Entry
	err := Database.Model(&Entry{}).Where("owner_id = ? AND journal_id = ?", userId, journalId).Find(&entries).Error
	if err != nil {
		return []Entry{}, err
	}
	return entries, nil
}

// Fetches an entry by ID from the database
func FetchEntryByID(userId uint, journalId uint, entryId uint) (*Entry, error) {
	var entry Entry
	err := Database.Model(&Entry{}).Where("id = ? AND owner_id = ? AND journal_id = ?", entryId, userId, journalId).First(&entry).Error
	if err != nil {
		return &Entry{}, err
	}
	return &entry, nil
}

// Deletes an entry from the database
func DeleteEntry(userId uint, journalId uint, entryId uint) error {
	err := Database.Model(&Entry{}).Where("id = ? AND owner_id = ? AND journal_id = ?", entryId, userId, journalId).Delete(&Entry{}).Error
	if err != nil {
		return err
	}
	return nil
}
