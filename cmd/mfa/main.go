package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/mdp/qrterminal/v3"

	"go.yhsif.com/pandablog/app/lib/timezone"
	"go.yhsif.com/pandablog/app/lib/totp"
	"go.yhsif.com/pandablog/app/logging"
)

func init() {
	logging.InitText(slog.LevelDebug)
	// Set the time zone.
	timezone.Set()
}

func main() {
	username := os.Getenv("PBB_USERNAME")
	if len(username) == 0 {
		log.Fatalln("Environment variable missing:", "PBB_USERNAME")
	}

	issuer := os.Getenv("PBB_ISSUER")
	if len(issuer) == 0 {
		log.Fatalln("Environment variable missing:", "PBB_ISSUER")
	}

	// Generate a MFA.
	URI, secret, err := totp.GenerateURL(username, issuer)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Output the TOTP URI and config information.
	fmt.Printf("PBB_MFA_KEY=%v\n", secret)
	fmt.Println("")
	fmt.Println("Send this to a mobile phone to add it to an app like Google Authenticator or scan the QR code below:")
	fmt.Printf("%v\n", URI)

	config := qrterminal.Config{
		Level:     qrterminal.L,
		Writer:    os.Stdout,
		BlackChar: qrterminal.WHITE,
		WhiteChar: qrterminal.BLACK,
		QuietZone: 1,
	}
	qrterminal.GenerateWithConfig(URI, config)
}
