package main

import (
	"encoding/hex"

	skyglownotificationlib "github.com/Preloading/SkyglowNotificationLibraries"
)

func main() {
	skyglownotificationlib.NotificationServer = "http://127.0.0.1:7878"
	notification := skyglownotificationlib.Notification{
		Message: "hi there",
		Sound:   "default",
	}

	// You will need to find this your device token for an app. Use a proxy
	hexString := "742e7072656c6f6164696e672e6465765d889add6b7f946be4722a2eb15a3142"
	decoded := make([]byte, hex.DecodedLen(len(hexString)))
	_, err := hex.Decode(decoded, []byte(hexString))
	if err != nil {
		panic(err)
	}

	err = notification.SendEncryptedNotification(decoded)
	if err != nil {
		panic(err)
	}
}
