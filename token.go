package main

import (
	"context"
	"time"

	d1client "github.com/cybercryptio/d1-client-go/d1-generic"
	pbauthn "github.com/cybercryptio/d1-client-go/d1-generic/protobuf/authn"
)

func GetStandaloneTokenFactory(d1Client d1client.GenericClient, uid, pwd string) func(context.Context) (string, error) {
	token := ""
	tokenExpiry := time.Now()
	return func(ctx context.Context) (string, error) {
		if time.Now().After(tokenExpiry.Add(time.Duration(-1) * time.Minute)) {
			res, err := d1Client.Authn.LoginUser(ctx, &pbauthn.LoginUserRequest{
				UserId:   uid,
				Password: pwd,
			})
			if err != nil {
				return "", err
			}
			token = res.AccessToken
			tokenExpiry = time.Unix(res.ExpiryTime, 0)
		}
		return token, nil
	}
}
