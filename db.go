package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func NewDB(path string) (*DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

type Person struct {
	ID        string `json:"id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname" gorm:"serializer:D1"`
}

func (db *DB) CreatePerson(person *Person) error {
	return db.Create(person).Error
}

func (db *DB) GetPerson(id string) (*Person, error) {
	person := &Person{}
	err := db.Where("id = ?", id).First(person).Error
	if err != nil {
		return nil, err
	}
	return person, nil
}

func (db *DB) GetPeople() ([]Person, error) {
	people := make([]Person, 0)
	err := db.Find(&people).Error
	if err != nil {
		return nil, err
	}
	return people, nil
}

func (db *DB) UpdatePerson(updatedPerson *Person) error {
	person := &Person{}
	err := db.Where("id = ?", updatedPerson.ID).First(person).Error
	if err != nil {
		return err
	}

	return db.Save(updatedPerson).Error
}

func (db *DB) DeletePerson(id string) error {
	person := &Person{}
	return db.Where("id = ?", id).Delete(person).Error
}
