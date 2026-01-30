package main

import (
	"hash/fnv"

	"github.com/sqids/sqids-go"
)

func getUniqueId(fname string) (string, error) {
	h := fnv.New64a()
	h.Write([]byte(fname))
	s, err := sqids.New(sqids.Options{MinLength: 12, Alphabet: "_-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"})
	if err != nil {
		return "", err
	}

	return s.Encode([]uint64{h.Sum64()})
}
