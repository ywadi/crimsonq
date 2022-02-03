package main

import (
	"ywadi/goq/RedconQ"
	"ywadi/goq/Structs"
)

func main() {
	crimsonQ := Structs.S_GOQ{}
	RedconQ.StartRedCon(":9001", &crimsonQ)
}
