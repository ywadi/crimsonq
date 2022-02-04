package Utils

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/oklog/ulid"
)

func PrintANSIlogo() {
	fmt.Print(`
	 .d8888b.          d8b                                           .d88888b.  
	d88P  Y88b         Y8P                                          d88P" "Y88b 
	888    888                                                      888     888 
	888        888d888 888 88888b.d88b.  .d8888b   .d88b.  88888b.  888     888 
	888        888P"   888 888 "888 "88b 88K      d88""88b 888 "88b 888     888 
	888    888 888     888 888  888  888 "Y8888b. 888  888 888  888 888 Y8b 888 
	Y88b  d88P 888     888 888  888  888      X88 Y88..88P 888  888 Y88b.Y8b88P 
	 "Y8888P"  888     888 888  888  888  88888P'  "Y88P"  888  888  "Y888888"  
										Y8b  
																				
	CrimsonQ V1.0 = Demon Running 									
	`)
	fmt.Println()
}

func ToJson(qmsg interface{}) string {
	j, err := json.Marshal(qmsg)
	if err != nil {
		log.Fatal(err)
	}
	return string(j)
}

func LexiPack(i uint64) string {
	str := strconv.FormatUint(i, 16)
	if i < 16 {
		str = "0" + str
	}
	return str
}

func Serialize(qmsg interface{}) []byte {
	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	if err := e.Encode(qmsg); err != nil {
		log.Fatal(err)
	}
	return b.Bytes()
}

//Send pointer to struct to fill out with decode to deserialize
func Deserialize(value []byte, StructType interface{}) {
	buf := bytes.NewBuffer(value)
	dec := gob.NewDecoder(bytes.NewBuffer(buf.Bytes()))
	if err := dec.Decode(StructType); err != nil {
		log.Fatal(err)
	}
}

func GenerateULID() string {
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	return (ulid.MustNew(ulid.Timestamp(t), entropy)).String()
}

func SliceContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
