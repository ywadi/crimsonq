package RedconQ

import (
	"fmt"
	"log"
	"strings"
	"ywadi/goq/Structs"
	"ywadi/goq/Utils"

	"github.com/spf13/viper"
	"github.com/tidwall/redcon"
)

type ConnContext struct {
	SelectDB string
	Auth     bool
}

var crimsonQ *Structs.S_GOQ

func StartRedCon(addr string, cq *Structs.S_GOQ) {
	initCommands()
	go cq.Init()
	crimsonQ = cq
	go log.Printf("started server at %s", addr)
	err := redcon.ListenAndServe(addr,
		execCommand,
		func(conn redcon.Conn) bool {
			ConnContext := ConnContext{Auth: false, SelectDB: ""}
			conn.SetContext(ConnContext)
			remoteIp := strings.Split(conn.RemoteAddr(), ":")[0]
			fmt.Println("Client connected from ", remoteIp)
			if viper.GetString("redcon_settings.ip_whitelist") == "*" {
				return true
			} else {
				grant := Utils.SliceContains(viper.GetStringSlice("redcon_settings.ip_whitelist"), remoteIp)
				return grant
			}

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
