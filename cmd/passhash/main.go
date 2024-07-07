package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"log/slog"
	"os"
	"syscall"

	"golang.org/x/term"

	"go.yhsif.com/pandablog/app/lib/passhash"
	"go.yhsif.com/pandablog/app/lib/timezone"
	"go.yhsif.com/pandablog/app/logging"
)

func init() {
	logging.InitText(slog.LevelDebug)
	// Set the time zone.
	timezone.Set()
}

func main() {
	var pass string
	if len(os.Args) >= 2 {
		pass = os.Args[1]
	} else {
		fmt.Print("Please input your desired password: ")
		p, err := term.ReadPassword(syscall.Stdin)
		fmt.Println()
		if err != nil {
			log.Fatalf("Unable to read password: %v", err)
		}
		pass = string(p)
	}

	// Generate a new private key.
	s, err := passhash.HashString(pass)
	if err != nil {
		log.Fatalln(err.Error())
	}

	sss := base64.StdEncoding.EncodeToString([]byte(s))
	fmt.Printf("PBB_PASSWORD_HASH=%v\n", sss)
}
