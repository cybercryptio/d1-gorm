package migration

/*

func migrate_doit(db *gorm.DB) {
	type User struct {
		gorm.Model
		FirstName  string
		FirstNameE string `gorm:"serializer:D1"`
		LastName   string
		LastNameE  string `gorm:"serializer:D1"`
	}

	db.AutoMigrate(&User{})

	var user User
	db.First(&user)
	fmt.Printf("user: %v\n", user)
	var users []User
	batchSize := 10
	db.FindInBatches(&users, batchSize, func(tx *gorm.DB, batch int) error {
		fmt.Printf("=== BATCH [%d] ===\n", batch)
		for i, user := range users {
			// batch processing found records
			user.FirstNameE = user.FirstName
			user.LastNameE = user.LastName
			users[i] = user
		}

		tx.Save(&users)

		fmt.Println("affected:", tx.RowsAffected)

		return nil
	})
}
*/

/*

func migrate_droprename(db *gorm.DB) {
	type User struct {
		gorm.Model
		FirstName string `gorm:"serializer:D1"`
		LastName  string `gorm:"serializer:D1"`
	}

	db.Migrator().DropColumn(&User{}, "first_name")
	db.Migrator().RenameColumn(&User{}, "first_name_e", "first_name")

	db.Migrator().DropColumn(&User{}, "last_name")
	db.Migrator().RenameColumn(&User{}, "last_name_e", "last_name")

	db.AutoMigrate(&User{})

	var user User
	db.First(&user)
	fmt.Printf("user: %v\n", user)

	var users []User
	batchSize := 10
	db.FindInBatches(&users, batchSize, func(tx *gorm.DB, batch int) error {
		fmt.Printf("=== BATCH [%d] ===\n", batch)
		for _, user := range users {
			// batch processing found records
			fmt.Printf("%d: %s %s\n", user.ID, user.FirstName, user.LastName)
		}

		return nil
	})
}
*/
