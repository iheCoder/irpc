package client

import (
	"log"
	"testing"
	"time"
)

func TestResetTime(t *testing.T) {
	timeout := 2 * time.Second
	timer := time.NewTimer(timeout)
	log.Printf("hello")
	go printTimely(timer)
	time.Sleep(time.Second)
	timer.Reset(timeout)

	time.Sleep(10 * time.Second)
}

func printTimely(timer *time.Timer) {
	for range timer.C {
		log.Printf("world")
	}
}
