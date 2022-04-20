package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/eviltomorrow/omega/internal/api/terminal"
	"github.com/eviltomorrow/omega/internal/conf"
)

var mode = flag.String("mode", "local", "mode of terminal")

func main() {
	flag.Parse()

	switch *mode {
	case "local":
		err := terminal.NewLocal("/bin/bash", "localhost", 10*time.Second)
		if err != nil {
			log.Fatal(err)
		}

	case "ssh":
		err := terminal.NewSSH("/bin/bash", "localhost", 10*time.Second, &terminal.Resource{
			Host:     "192.168.95.118",
			Port:     22,
			Username: "root",
			// Password:       "sigmac95118",
			PrivateKeyPath: "/home/shepard/.ssh/id_rsa",
			Timeout:        conf.Duration{Duration: 10 * time.Second},
		})
		if err != nil {
			log.Fatal(err)
		}

	default:
		panic(fmt.Errorf("not support mode, mode: %v", *mode))
	}

}
