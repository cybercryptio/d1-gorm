// Copyright 2022 CYBERCRYPT
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License

package d1gorm_test

import (
	"fmt"
	"log"
	"os"

	client "github.com/cybercryptio/d1-client-go/d1-generic"
	d1gorm "github.com/cybercryptio/d1-gorm"
	"github.com/cybercryptio/d1-gorm/crypto"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	d1Serializer := d1gorm.NewD1Serializer(crypto.NewD1Cryptor(client))

	// Register the D1Serializer to be used for your database schema
	schema.RegisterSerializer("D1", d1Serializer)

	// Create a connection to your database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		log.Fatal(err)
	}
	_ = db.AutoMigrate(&Person{})

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
