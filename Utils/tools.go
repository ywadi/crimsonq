package Utils

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"
	"strconv"
)

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
