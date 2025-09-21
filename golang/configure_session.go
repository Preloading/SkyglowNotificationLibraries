package skyglownotificationlib

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

var (
	httpNotificationServer string
)

// Configures this library for most actions, like sending notifications, recieving feedback, etc.
// This must be called before any other SGN lib function
//
// notification server - This needs to be set to a Skyglow Notification server that you trust to correctly route notifications to their destinations.
func ConfigureSession(notificationServer string) error {
	txts, err := net.LookupTXT(fmt.Sprintf("_sgn.%s", notificationServer))
	if err != nil {
		return errors.New("failed to lookup txt record, does the SGN server exist?")
	}

	// Split the input by spaces to get key-value pairs
	parts := strings.Fields(txts[0])

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("invalid format in part: %s", part)
		}

		key := kv[0]
		value := kv[1]

		switch key {
		case "tcp_addr":
		case "tcp_port":
		// 	port, err := strconv.Atoi(value)
		// 	if err != nil {
		// 		return result, fmt.Errorf("invalid TCP port: %v", err)
		// 	}
		// 	result.TCPPort = port
		case "http_addr":
			// TODO: Validate this is starts with either https or http, and that it is not localhost or reserved IPs
			httpNotificationServer = value
		}
	}

	if httpNotificationServer == "" {
		return fmt.Errorf("the notification server has an invalid config")
	}

	return nil
}
