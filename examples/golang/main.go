package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	skyglownotificationlib "github.com/Preloading/SkyglowNotificationLibraries/golang"
)

func main() {
	skyglownotificationlib.ConfigureSession("d.preloading.dev")

	// keep this secret!
	feedbackSecretHex := "368d866b4efa2fb515c6e342f3aa825a780c5e65444a4fb1317a4fb98e1df76e29c1865ba01769c912f619eed9ca53fa228ef57f1b580cd856fcc7ae3c3b40f7b71e0697d9dcb4d202711fa5f82442acde1431ee2290f849e2b0359136ece1686b2569626f14531c9bc486113b378d5d0d4d82e725a7facac4afc9f67eea789f5d1f2dcab8f9e047bf67e844ad6d7de464685fa150403bf43ea786a7e8342c475d1093c450638d5e94f0af28749ea40f10954d13b16ed81d477e764c524e861a163d3fe5257db24f4018a177269bb3efd1c2c3351393d3b50ce19ccd3ed21bd65eab496c20891aac5fae8c67df8b6a283c8608dba9df9e0812bbe5dea494b33b"
	feedbackSecret, _ := hex.DecodeString(feedbackSecretHex)

	// You will need to find this your device token for an app. Use a proxy to retrieve if you just want to demo it.
	// hexDeviceToken := "742e7072656c6f6164696e672e646576badf9650c3c04b5523a93fb847ea6645"
	// deviceToken := make([]byte, hex.DecodedLen(len(hexDeviceToken)))
	// _, err := hex.Decode(deviceToken, []byte(hexDeviceToken))
	// if err != nil {
	// 	panic(err)
	// }

	base64DeviceToken := "ZC5wcmVsb2FkaW5nLmRldheJMhPP3zZrlGx1ie+8Q6g="

	deviceToken := make([]byte, base64.StdEncoding.DecodedLen(len(base64DeviceToken))-1)
	_, err := base64.StdEncoding.Decode(deviceToken, []byte(base64DeviceToken))
	if err != nil {
		log.Fatal("error:", err)
	}

	// get the token's components. you need to store these along with your device token if you want to act upon feedback recieved.
	// routing_key, token_server_address, err := skyglownotificationlib.RoutingInfoFromDeviceToken(feedbackSecret)
	// if err != nil {
	// 	// In real life cases, you probably don't want to crash the app if the token is invalid, especially because APNS tokens can very easily sneak in.
	// 	panic("token is not a valid sgn token")
	// }

	// register the token for feedback. this can only be done once.
	err = skyglownotificationlib.ConfigureTokenForFeedback(deviceToken, feedbackSecret)
	if err != nil {
		fmt.Printf("error when registering token for feedback: %s. if it has already been registered, this can be disregarded.\n", err.Error())
	}

	notification := map[string]interface{}{
		"aps": map[string]interface{}{
			"alert": "bitch",
			"sound": "default",
		},
	}

	// AOL Instant Messenger example
	// notification := map[string]interface{}{
	// 	"aps": map[string]interface{}{
	// 		"alert": map[string]interface{}{
	// 			"loc-key":  "IM2",
	// 			"loc-args": []interface{}{"loganh4005", "loganh4005", "hello!", "loganh4005"},
	// 		},
	// 		"sound": "default.caf",
	// 	},
	// }

	err = skyglownotificationlib.SendNotification(deviceToken, notification)
	if err != nil {
		fmt.Printf("error when sending notification: %s\n", err)
	}

	// Get feedback
	feedback, err := skyglownotificationlib.GetFeedback(feedbackSecret, time.Unix(1, 0))
	if err != nil {
		fmt.Printf("error when recieving feedback: %s\n", err)
	}

	for i, f := range feedback {
		fmt.Printf("Feedback #%v:\n", i)
		fmt.Printf(" type of feedback: %v\n", f.Type)
		fmt.Printf(" reason for feedback: %v\n", f.Reason)
		fmt.Printf(" routing token: %v\n", hex.EncodeToString(f.RoutingToken))
		fmt.Printf(" routing token's server: %v\n", f.RoutingTokenServer)
		fmt.Printf(" created at: %v\n", f.CreatedAt)
	}
}
