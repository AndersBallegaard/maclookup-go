package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func generateMac() string {
	outputmac := ""
	delcounter := 0
	for i := 0; i < 12; i++ {
		rnd := rand.Intn(15)
		outputmac += fmt.Sprintf("%x", rnd)
		delcounter++
		if delcounter >= 2 && i != 11 {
			outputmac += ":"
			delcounter = 0
		}
	}
	return outputmac
}

func generateUrl() string {
	baseurl := "http://localhost:8080/lookup?mac="
	mac := generateMac()
	return baseurl + mac
}

func tester(hits int32) {
	urls := []string{}
	for i := int32(0); i < hits; i++ {
		urls = append(urls, generateUrl())
	}
	starttime := time.Now()
	for _, url := range urls {
		//println(url)
		http.Get(url)
	}
	duration := time.Since(starttime)
	println("Test complete " + fmt.Sprintf("%v",int32(duration.Nanoseconds())/hits) + "ns/request, total execution time " + fmt.Sprintf("%v",duration.Nanoseconds()) + "ns")
}

func main() {
	println("Mac lookup api performance testing")
	tester(100000)
}