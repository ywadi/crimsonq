package RedconQ

import (
	"log"
	"ywadi/goq/Structs"

	"github.com/tidwall/redcon"
)

type ConnContext struct {
	SelectDB string
	Auth     bool
}

var crimsonQ *Structs.S_GOQ

func StartRedCon(addr string, cq *Structs.S_GOQ) {
	initCommands()
	go cq.Init("TODO replace with settings")
	crimsonQ = cq
	go log.Printf("started server at %s", addr)
	err := redcon.ListenAndServe(addr,
		execCommand,
		func(conn redcon.Conn) bool {
			ConnContext := ConnContext{Auth: false, SelectDB: ""}
			conn.SetContext(ConnContext)
			return true
		},
		func(conn redcon.Conn, err error) {
			// This is called when the connection has been closed
			// log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)

		},
	)
	if err != nil {
		log.Fatal(err)
	}
}
