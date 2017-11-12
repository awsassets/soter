package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/asdine/storm"
	"github.com/namsral/flag"
	"github.com/thoj/go-ircevent"
)

const (
	nickname = "soter"
	username = "Soter"
	realname = "Saviour, Deliverer"
)

var (
	db *storm.DB
)

type Addr struct {
	Host   string
	Port   int
	UseTLS bool
}

func (a *Addr) String() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

func ParseAddr(s string) (addr *Addr, err error) {
	addr = &Addr{}

	parts := strings.Split(s, ":")
	fmt.Printf("%v", parts)
	if len(parts) != 2 {
		return nil, fmt.Errorf("malformed address: %s", s)
	}

	addr.Host = parts[0]

	if parts[1][0] == '+' {
		port, err := strconv.Atoi(parts[1][1:])
		if err != nil {
			return nil, fmt.Errorf("invalid port: %s", parts[1])
		}
		addr.Port = port
		addr.UseTLS = true
	} else {
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid port: %s", parts[1])
		}
		addr.Port = port
	}

	if addr.Port < 1 || addr.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", addr.Port)
	}

	return addr, nil
}

func main() {
	var (
		err error

		version bool
		config  string
		debug   bool

		dbpath   string
		operuser string
		operpass string

		authed bool
	)

	flag.BoolVar(&version, "v", false, "display version information")
	flag.StringVar(&config, "c", "", "config file")
	flag.BoolVar(&debug, "d", false, "debug logging")

	flag.StringVar(&operuser, "operuser", "", "irc operator username")
	flag.StringVar(&operpass, "operpass", "", "irc operator password")
	flag.StringVar(&dbpath, "dbpath", "soter.db", "path to database")

	flag.Parse()

	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}

	if version {
		fmt.Printf("soter v%s", FullVersion())
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		log.Fatalf("Ussage: %s <address>[:port]", os.Args[0])
	}

	addr, err := ParseAddr(flag.Arg(0))
	if err != nil {
		log.Fatalf("error parsing addr: %s", err)
	}

	db, err = storm.Open(dbpath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	conn := irc.IRC(nickname, username)
	conn.RealName = realname

	conn.VerboseCallbackHandler = debug
	conn.Debug = debug

	conn.UseTLS = addr.UseTLS
	conn.KeepAlive = 30 * time.Second
	conn.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	conn.AddCallback("001", func(e *irc.Event) {
		log.Info("Connected!")

		var channels []Channel

		err := db.All(&channels)
		if err != nil {
			log.Fatalf("error loading channels from db: %s", err)
		}

		log.Infof("Opering up with %s", operuser)
		conn.SendRawf("OPER %s %s", operuser, operpass)

		for _, channel := range channels {
			conn.Join(channel.Name)
			conn.Mode(channel.Name)
			log.Infof("Joined %s", channel.Name)
		}
	})
	conn.AddCallback("381", func(e *irc.Event) {
		authed = true
		log.Infof("Successfully opered up!")
	})

	conn.AddCallback("324", func(e *irc.Event) {
		var mode Mode

		err = db.One("Channel", e.Arguments[1], &mode)
		if err != nil && err == storm.ErrNotFound {
			mode = NewMode(e.Arguments[1], e.Arguments[2:])
			err := db.Save(&mode)
			if err != nil {
				log.Fatalf("error saving mode to db: %s", err)
			}
		} else if err != nil {
			log.Fatalf("error looking up mode in db: %s", err)
		}

		if !mode.Equal(e.Arguments[2:]) {
			conn.Mode(mode.Channel, mode.Modes...)
		}
	})
	conn.AddCallback("MODE", func(e *irc.Event) {
		if e.Arguments[0][0] != '#' {
			return
		}

		var mode Mode
		err = db.One("Channel", e.Arguments[0], &mode)
		if err != nil && err == storm.ErrNotFound {
			mode = NewMode(e.Arguments[0], e.Arguments[1:])
			err := db.Save(&mode)
			if err != nil {
				log.Fatalf("error saving mode to db: %s", err)
			}
		} else if err != nil {
			log.Fatalf("error looking up mode in db: %s", err)
		}

		mode.AppendModes(e.Arguments[1:])

		err := db.Save(&mode)
		if err != nil {
			log.Fatalf("error saving mode to db: %s", err)
		}
	})

	conn.AddCallback("332", func(e *irc.Event) {
		var topic Topic

		err = db.One("Channel", e.Arguments[1], &topic)
		if err != nil && err == storm.ErrNotFound {
			topic = NewTopic(e.Arguments[1], e.Arguments[2])
			err := db.Save(&topic)
			if err != nil {
				log.Fatalf("error saving topic to db: %s", err)
			}
		} else if err != nil {
			log.Fatalf("error looking up topic in db: %s", err)
		}

		if topic.Topic != e.Arguments[2] {
			conn.SendRawf("TOPIC %s :%s", topic.Channel, topic.Topic)
		}
	})
	conn.AddCallback("TOPIC", func(e *irc.Event) {
		var topic Topic
		err = db.One("Channel", e.Arguments[0], &topic)
		if err != nil && err == storm.ErrNotFound {
			topic = NewTopic(e.Arguments[0], e.Arguments[1])
			err := db.Save(&topic)
			if err != nil {
				log.Fatalf("error saving topic to db: %s", err)
			}
		} else if err != nil {
			log.Fatalf("error looking up topic in db: %s", err)
		}

		topic.SetTopic(e.Arguments[1])

		err := db.Save(&topic)
		if err != nil {
			log.Fatalf("error saving topic to db: %s", err)
		}
	})
	conn.AddCallback("JOIN", func(e *irc.Event) {
		channel := e.Arguments[0]
		if e.Nick == "soter" {
			conn.Mode(channel, "+o", e.Nick)
		}
	})

	conn.AddCallback("INVITE", func(e *irc.Event) {
		var channel Channel
		err = db.One("Name", e.Arguments[0], &channel)
		if err != nil && err == storm.ErrNotFound {
			channel = NewChannel(e.Arguments[1])
			err := db.Save(&channel)
			if err != nil {
				log.Fatalf("error saving channel to db: %s", err)
			}
		} else if err != nil {
			log.Fatalf("error looking up channel in db: %s", err)
		}

		conn.Join(e.Arguments[1])
		conn.Mode(e.Arguments[1])
	})

	err = conn.Connect(addr.String())
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}

	conn.Loop()
}
