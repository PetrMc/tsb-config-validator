package validator

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/PetrMc/tsb-config-validator/src/collector"
	jwt "github.com/dgrijalva/jwt-go"
)

// TokenCheck is a function to go through all the tokens
// that require validation
// Since the token names in TSB are not standardized
// we have to irritate through them

func TokenCheck(tkn *collector.TSBTokens) {

	TknExpiry("oap", tkn.Oapt)
	TknExpiry("Zipkin", tkn.Zipkint)
	TknExpiry("Open Telemetry (Otel)", tkn.Otelt)
	TknExpiry("XCP", tkn.Xcpt)

}

// TknExpiry check if time if the token is still valid
// the current program can't access the private key
// and check tokens signature
func TknExpiry(n string, a string) {

	// the function displays the finding and needs to provide easily readable
	// output - CustomPrint is used for that propose
	p := CustomPrint()

	// First check if token is actually collected
	if len(a) == 0 {
		return
	}

	// Read the token - print Error if any issues found
	tkn, _, err := new(jwt.Parser).ParseUnverified(a, jwt.MapClaims{})
	if err != nil {
		fmt.Println(err.Error())
	}

	// Read the claims in the token
	clm, ok := tkn.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Printf("Can't convert token's claims to standard claims")
	}

	// defining token time placeholder
	var tt time.Time
	// reading the current time
	ct := time.Now()

	// decision flow on token format. (as the docs suggest in can vary)
	switch iat := clm["exp"].(type) {
	case float64:
		tt = time.Unix(int64(iat), 0)
	case json.Number:
		v, _ := iat.Int64()
		tt = time.Unix(v, 0)
	}

	// checking the token and informing operator if expired
	if ct.After(tt) {
		fmt.Printf("\n%v\nToken %v has expired on %v\nPlease use 'tctl' utility to redeploy new tokens \n%v\n", p.Stars, n, tt, p.Stars)
	}

}
