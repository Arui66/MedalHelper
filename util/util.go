package util

import (
	"fmt"
	"math/rand"
	"net/url"
	"time"
)

// Map2String can transfer a string-string map into a raw string
func Map2string(params map[string]string) string {
	var query string
	for k, v := range params {
		query += k + "=" + v + "&"
	}
	query = query[:len(query)-1]
	return query
}

// Map2String can transfer a string-string map into url value struct
func Map2Params(params map[string]string) url.Values {
	value := url.Values{}
	for key, param := range params {
		value[key] = []string{param}
	}
	return value
}

// GetTimestamp can obtain current ts
func GetTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

// StringContains judge whether val exist in array
func IntContain(array []int, val int) (index int) {
	index = -1
	for i := 0; i < len(array); i++ {
		if array[i] == val {
			index = i
			return
		}
	}
	return
}

func RandomString(length int) string {
	source := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	rand.Shuffle(length, func(i, j int) {
		source[i], source[j] = source[j], source[i]
	})
	return string(source[:length])
}
