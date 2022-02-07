package RedconQ

import (
	"fmt"
	"strings"
	"time"
	"ywadi/crimsonq/Defs"
	"ywadi/crimsonq/Structs"
	"ywadi/crimsonq/Utils"

	log "github.com/sirupsen/logrus"

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
	log.Info("started server at %s", addr)
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
	log.Info("Heartbeat Started...")
	ticker := time.NewTicker(time.Duration(viper.GetInt64("crimson_settings.heartbeat_seconds")) * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				for _, s := range crimsonQ.QDBPool {
					count, err := crimsonQ.GetKeyCount(s.QdbId)
					if err != nil {
						log.WithFields(log.Fields{"ConsumerId": s.QdbId, "Status": Defs.STATUS_PENDING}).Error("JSON Parse error at heartbeart", err)
					}
					PS.Publish(s.QdbId, "pendingCount:"+fmt.Sprint(count[Defs.STATUS_PENDING]))
				}
			}
		}
	}()
}
