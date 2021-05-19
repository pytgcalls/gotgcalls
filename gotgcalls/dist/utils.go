package main

import (
	"reflect"
	"strconv"
	"strings"
)

func (r *RTCPeerConnectionClient) parseSdp(sdp string) Sdp{
	lines := strings.Split(sdp,"\r\n")
	lookup := func(prefix string) *string{
		for line := range lines {
			if strings.HasPrefix(lines[line], prefix){
				result := lines[line][len(prefix):]
				return &result
			}
		}
		return nil
	}
	rs := lookup("a=ssrc:")
	var rawSource *int
	if rs != nil{
		rs2, _ := strconv.Atoi(strings.Split(*rs, " ")[0])
		rawSource = &rs2
	}

	var fingerprint *string
	var hash *string
	fingerprint = lookup("a=fingerprint:")
	if fingerprint != nil{
		tmpFinger := strings.Split(*fingerprint, " ")
		hash = &tmpFinger[0]
		fingerprint = &tmpFinger[1]
	}
	return Sdp{
		fingerprint: fingerprint,
		hash: hash,
		setup: lookup("a=setup:"),
		pwd: lookup("a=ice-pwd:"),
		ufrag: lookup("a=ice-ufrag:"),
		source: rawSource,
	}
}
func normalizeInt(value interface{}) int{
	if reflect.TypeOf(value).Kind() == reflect.Int64{
		return int(value.(int64))
	}else if reflect.TypeOf(value).Kind() == reflect.Int32{
		return int(value.(int32))
	}else if reflect.TypeOf(value).Kind() == reflect.Float32{
		return int(value.(float32))
	}else if reflect.TypeOf(value).Kind() == reflect.Float64{
		return int(value.(float64))
	}else{
		return value.(int)
	}
}
func normalizeString(value interface{}) *string{
	if value != nil{
		strTmp := value.(string)
		return &strTmp
	}else{
		return nil
	}
}