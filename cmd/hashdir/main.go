package main

import (
	"golang.org/x/mod/sumdb/dirhash"
	"log"
)

func main() {
	log.Print(dirhash.HashZip("/Users/dvtkrlbs/Downloads/v0.0.0-20190523083050-ea95bdfd59fc.zip", dirhash.DefaultHash))
}
