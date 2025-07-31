package skyglownotificationlib

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

var (
	NotificationServer string
)

type Notification struct {
	Message     string // The body of the notifcation.
	Sound       string // A sound located inside the app. Leave empty for the default
	BadgeNumber *int   // The badge number displayed next to the app. Leave it at nil for no change.
	Action      string // If empty, this will be slide to view, if "Open", it will be slide to Open
}

func (c *Notification) SendNotification(device_token []byte) error {
	if NotificationServer == "" {
		panic("NotificationServer not set before sending notification!")
	}

	if len(device_token) != 32 {
		return fmt.Errorf("routing key size incorrect, expected 32 got %d", len(device_token))
	}
	serverAddressByte := device_token[0:15]
	serverAddressByte = bytes.Trim(serverAddressByte, "\x00")
	serverAddress := string(serverAddressByte)
	secretValue := device_token[15:31]

	// check if server address fits regex
	isValidServerAddress, _ := regexp.Match("(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]", []byte(serverAddress))
	if !isValidServerAddress {
		return errors.New("routing key address to domain invalid")
	}

	// generate routing key
	s := sha256.New()
	s.Write(secretValue)
	routing_key := s.Sum(nil)
	hexRoutingKey := hex.EncodeToString(routing_key)

	type noEncryptionNotification struct {
		ServerAddress string `json:"server_address"`
		RoutingKey    string `json:"routing_key"`

		// Topic       string `json:"topic"` // Bundle ID
		Message     string `json:"message"`
		Sound       string `json:"alert_sound"`
		Action      string `json:"alert_action"`
		BadgeNumber *int   `json:"badge_number"`
	}

	// json encode notification data
	json.Marshal(noEncryptionNotification{
		RoutingKey:    hexRoutingKey,
		ServerAddress: serverAddress,

		Message:     c.Message,
		Sound:       c.Sound,
		Action:      c.Action,
		BadgeNumber: c.BadgeNumber,
	})

	return nil
}
