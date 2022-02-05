package RedconQ

import (
	"fmt"
	"log"
	"strings"
	"time"
	"ywadi/goq/Defs"
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
	HeartBeat()
	crimsonQ = cq
	log.Printf("started server at %s", addr)
	err := redcon.ListenAndServe(addr,
		execCommand,
		func(conn redcon.Conn) bool {
			ConnContext := ConnContext{Auth: false, SelectDB: ""}
			conn.SetContext(ConnContext)
			remoteIp := strings.Split(conn.RemoteAddr(), ":")[0]
			fmt.Println("Client connected from ", remoteIp)
			if viper.GetString("crimson_settings.ip_whitelist") == "*" {
				return true
			} else {
				grant := Utils.SliceContains(viper.GetStringSlice("crimson_settings.ip_whitelist"), remoteIp)
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

func HeartBeat() {
	println("Heartbeat Started...")
	ticker := time.NewTicker(time.Duration(viper.GetInt64("crimson_settings.heartbeat_seconds")) * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				for _, s := range crimsonQ.QDBPool {
					fmt.Println("Heartbeat")
					json, err := crimsonQ.GetAllByStatusJson(s.QdbId, Defs.STATUS_PENDING)
					if err != nil {
						fmt.Println("JSON ERROR")
					}
					PS.Publish(s.QdbId, json)
				}
			}
		}
	}()
}
