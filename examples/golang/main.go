package main

import (
	"encoding/hex"

	skyglownotificationlib "github.com/Preloading/SkyglowNotificationLibraries"
)

func main() {
	skyglownotificationlib.NotificationServer = "http://127.0.0.1:7878"
	notification := map[string]interface{}{
		"aps": map[string]interface{}{
			"alert": "session is ass",
			"sound": "default.caf",
		},
	}

	// You will need to find this your device token for an app. Use a proxy
	// hexString := "642e7072656c6f6164696e672e6465767e4740298ae97f8a05aa97040860130a"
	// hexString := "73676e2e736b79676c6f772e65730000f256841a394b67ff4476bc08464ee67a"
	hexString := "742e7072656c6f6164696e672e646576a6e34fa095125b1d0165174c86407389"
	decoded := make([]byte, hex.DecodedLen(len(hexString)))
	_, err := hex.Decode(decoded, []byte(hexString))
	if err != nil {
		panic(err)
	}

	err = skyglownotificationlib.SendNotification(decoded, notification)
	if err != nil {
		panic(err)
	}
}
