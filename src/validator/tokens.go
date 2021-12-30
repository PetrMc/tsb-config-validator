package validator

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/PetrMc/tsb-config-validator/src/collector"
	jwt "github.com/dgrijalva/jwt-go"
)

func TokenCheck(tkn *collector.TSBTokens) {

	TknExpiry("oap", tkn.Oapt)
	TknExpiry("Zipkin", tkn.Zipkint)
	TknExpiry("Open Telemetry (Otel)", tkn.Otelt)
	TknExpiry("XCP", tkn.Xcpt)

}
func TknExpiry(n string, a string) {

	p := CustomPrint()

	if len(a) == 0 {
		return
	}

	tkn, _, err := new(jwt.Parser).ParseUnverified(a, jwt.MapClaims{})
	if err != nil {
		fmt.Println(err.Error())
	}

	clm, ok := tkn.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Printf("Can't convert token's claims to standard claims")
	}

	var tm time.Time
	now := time.Now()
	switch iat := clm["exp"].(type) {
	case float64:
		tm = time.Unix(int64(iat), 0)
	case json.Number:
		v, _ := iat.Int64()
		tm = time.Unix(v, 0)
	}

	if now.After(tm) {
		fmt.Printf("\n%v\nToken %v has expired on %v\nPlease use 'tctl' utility to redeploy new tokens \n%v\n", p.Stars, n, tm, p.Stars)
	}

}
