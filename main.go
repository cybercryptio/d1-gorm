package main

import (
	"fmt"
	"os"

	"github.com/cybercryptio/d1-gorm/migration"

	d1client "github.com/cybercryptio/d1-client-go/d1-generic"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/schema"
)

type Server struct {
	db     *DB
	router *gin.Engine
}

var server Server

func main() {
	uid, ok := os.LookupEnv("D1_UID")
	if !ok {
		fmt.Println("D1_UID is not set")
		os.Exit(1)
	}
	pwd, ok := os.LookupEnv("D1_PWD")
	if !ok {
		fmt.Println("D1_PWD is not set")
		os.Exit(1)
	}

	d1Client, err := d1client.NewGenericClient("localhost:9000", "")
	if err != nil {
		fmt.Println("Error creating d1 client:", err)
		os.Exit(1)
	}

	tokenFactory := GetStandaloneTokenFactory(d1Client, uid, pwd)

	d1Cryptor := NewD1Cryptor(d1Client, tokenFactory)

	schema.RegisterSerializer("D1", d1Cryptor)

	db, err := NewDB("./gorm.db")
	if err != nil {
		fmt.Println("Error creating DB:", err)
		os.Exit(1)
	}

	migration.TryMigrate(db.DB)
	return

	db.AutoMigrate(&Person{})

	router := gin.Default()
	router.POST("/people", CreatePerson)
	router.GET("/people/:id", GetPerson)
	router.GET("/people/", GetPeople)
	router.PUT("/people/:id", UpdatePerson)
	router.DELETE("/people/:id", DeletePerson)

	server = Server{db: db, router: router}

	server.router.Run(":8080")
}

func CreatePerson(c *gin.Context) {
	person := &Person{}
	c.BindJSON(person)

	err := server.db.CreatePerson(person)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, person)
}

func GetPerson(c *gin.Context) {
	id := c.Params.ByName("id")

	person, err := server.db.GetPerson(id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, person)
}

func GetPeople(c *gin.Context) {
	people, err := server.db.GetPeople()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, people)
}

func UpdatePerson(c *gin.Context) {
	person := &Person{}
	c.BindJSON(person)

	person.ID = c.Params.ByName("id")

	err := server.db.UpdatePerson(person)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, person)
}

func DeletePerson(c *gin.Context) {
	id := c.Params.ByName("id")

	err := server.db.DeletePerson(id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "deleted"})
}
