package models

import (
	"gorm.io/gorm"
)

type Journal struct {
	gorm.Model        // Already includes "id", "created_at", "updated_at", "deleted_at"
	Name       string `json:"name" bson:"name"`
	OwnerID    uint   `json:"owner_id" bson:"owner_id"`
}

// Saves a journal to the database
func (journal *Journal) Save() (*Journal, error) {
	err := Database.Model(&journal).Create(&journal).Error
	if err != nil {
		return &Journal{}, err
	}
	return journal, nil
}

// Fetches all journals from the database
func FetchAllJournals(userId uint) ([]Journal, error) {
	var journals []Journal
	err := Database.Model(&Journal{}).Where("owner_id = ?", userId).Find(&journals).Error
	if err != nil {
		return []Journal{}, err
	}
	return journals, nil
}

// Fetches a journal by ID from the database
func FetchJournalByID(userId uint, journalId uint) (*Journal, error) {
	var journal Journal
	err := Database.Model(&Journal{}).Where("id = ? AND owner_id = ?", journalId, userId).First(&journal).Error
	if err != nil {
		return &Journal{}, err
	}
	return &journal, nil
}

// Deletes a journal from the database
func DeleteJournal(userId uint, journalId uint) error {
	err := Database.Model(&Journal{}).Where("id = ? AND owner_id = ?", journalId, userId).Delete(&Journal{}).Error
	if err != nil {
		return err
	}
	return nil
}
