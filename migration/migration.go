package migration

import (
	"gorm.io/gorm"
)

func Migrate[T any](db *gorm.DB, migrate func(*T), opts ...option) *gorm.DB {
	o := defaultOptions()
	o.apply(opts...)

	session := db.Session(&gorm.Session{QueryFields: true})
	var entries []T
	result := session.FindInBatches(&entries, o.batchSize, func(tx *gorm.DB, batch int) error {
		for i := range entries {
			migrate(&entries[i])
		}

		tx.Save(&entries)

		return nil
	})
	return result
}
