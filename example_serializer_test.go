package d1gorm_test

import (
	"fmt"
	"log"
	"os"

	client "github.com/cybercryptio/d1-client-go/d1-generic"
	d1gorm "github.com/cybercryptio/d1-gorm"
	"github.com/cybercryptio/d1-gorm/encrypt"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var endpoint = os.Getenv("D1_ENDPOINT")
var uid = os.Getenv("D1_UID")
var password = os.Getenv("D1_PASS")
var creds = insecure.NewCredentials()

type Person struct {
	ID        string
	FirstName string
	LastName  string `gorm:"serializer:D1"`
}

func Example() {
	// Create a new D1 Generic client
	client, err := client.NewGenericClient(endpoint,
		client.WithTransportCredentials(creds),
		client.WithPerRPCCredentials(
			client.NewStandalonePerRPCToken(endpoint, uid, password, creds),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Instantiate the D1Serializer with a Cryptor that uses the created client
	d1Serializer := d1gorm.NewD1Serializer(encrypt.NewD1Cryptor(client))

	// Register the D1Serializer to be used for your database schema
	schema.RegisterSerializer("D1", d1Serializer)

	// Create a connection to your database
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&Person{})

	// Now all the created Persons will have their fields tagged with "serializer:D1" encrypted before being written to the database
	michael := &Person{"1", "Michael", "Jackson"}
	// Michael's last name will be encrypted
	db.Create(michael)

	// The encrypted data is transparently decrypted when reading from the database
	ret := &Person{}
	db.Where("id = ?", "1").First(ret)

	fmt.Printf("First Name: %s Last Name: %s", ret.FirstName, ret.LastName)
	// Out: First Name: Michael Last Name: Jackson
}
