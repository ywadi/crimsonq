package main

import (
	"fmt"
	"sync"
	"ywadi/goq/Structs"
)

func main() {
	qmsg := Structs.S_QMSG{}
	qmsg.Init(
		"mykey",
		"123",
		"/a/b/c",
		"my bigass message",
	)
	fmt.Println(qmsg.JsonStringify())
	stayAlive()
}

func stayAlive() {
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
