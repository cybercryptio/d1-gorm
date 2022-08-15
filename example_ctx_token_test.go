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
	"context"
	"fmt"
	"log"

	client "github.com/cybercryptio/d1-client-go/d1-generic"
	d1gorm "github.com/cybercryptio/d1-gorm"
	"github.com/cybercryptio/d1-gorm/crypto"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type ctxKeyType string

var tokenKey = ctxKeyType("token")

func getToken() string {
	token, err := client.NewStandalonePerRPCToken(endpoint, uid, password, creds)(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return token
}

func Example_passTokenInCtx() {
	// Create a new D1 Generic client which fetches the token from the context on each request
	client, err := client.NewGenericClient(endpoint,
		client.WithTransportCredentials(creds),
		client.WithPerRPCCredentials(
			client.PerRPCToken(func(ctx context.Context) (string, error) {
				token, found := ctx.Value(tokenKey).(string)
				if !found {
					return "", fmt.Errorf("token not found in context")
				}
				return token, nil
			}),
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
	// Note that we pass in the context the access token to be used by the D1 Generic client
	db.WithContext(context.WithValue(context.Background(), tokenKey, getToken())).Create(michael)
}
