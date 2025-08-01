package skyglownotificationlib

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"golang.org/x/crypto/hkdf"
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

// Sends an unencrypted notification to the client. In general this is more efficient to send a notification, but the intermediete servers may be able to see it.
func (c *Notification) SendNotification(device_token []byte) error {
	if NotificationServer == "" {
		panic("NotificationServer not set before sending notification!")
	}

	if len(device_token) != 32 {
		return fmt.Errorf("routing key size incorrect, expected 32 got %d", len(device_token))
	}
	serverAddressByte := device_token[0:16]
	serverAddressByte = bytes.Trim(serverAddressByte, "\x00")
	serverAddress := string(serverAddressByte)
	secretValue := device_token[16:32]

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
	data, err := json.Marshal(noEncryptionNotification{
		RoutingKey:    hexRoutingKey,
		ServerAddress: serverAddress,

		Message:     c.Message,
		Sound:       c.Sound,
		Action:      c.Action,
		BadgeNumber: c.BadgeNumber,
	})

	if err != nil {
		return err
	}

	resp, err := http.Post(fmt.Sprintf("%s/send", NotificationServer), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	type responceData struct {
		Status string
	}

	decodedResp := responceData{}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := json.Unmarshal(bodyBytes, &decodedResp); err != nil {
		return err
	}

	if decodedResp.Status != "sucess" {
		return errors.New(decodedResp.Status)
	}

	return nil
}

// Sends a encrypted notification to the device.
func (c *Notification) SendEncryptedNotification(device_token []byte) error {
	if NotificationServer == "" {
		panic("NotificationServer not set before sending notification!")
	}

	if len(device_token) != 32 {
		return fmt.Errorf("routing key size incorrect, expected 32 got %d", len(device_token))
	}
	serverAddressByte := device_token[0:16]
	serverAddressByte = bytes.Trim(serverAddressByte, "\x00")
	serverAddress := string(serverAddressByte)
	secretValue := device_token[16:32]

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

	// HKDF salt
	hkdfSalt := fmt.Sprintf("%sHello from the Skyglow Notifications developers!", serverAddress) // The "Hello from" text is **required** to reimplement this.

	// generate HKDF cert
	hkdfReader := hkdf.New(sha256.New, secretValue, []byte(hkdfSalt), nil)

	cryptoKey := make([]byte, 32)
	_, err := io.ReadFull(hkdfReader, cryptoKey)
	if err != nil {
		return fmt.Errorf("failed to derive encryption key: %s", err.Error())
	}

	type encryptedNotificationData struct {
		Message     string `json:"message"`
		Sound       string `json:"alert_sound"`
		Action      string `json:"alert_action"`
		BadgeNumber *int   `json:"badge_number"`
	}

	// json encode notification notificationData
	notificationData, err := json.Marshal(encryptedNotificationData{
		Message:     c.Message,
		Sound:       c.Sound,
		Action:      c.Action,
		BadgeNumber: c.BadgeNumber,
	})

	if err != nil {
		return err
	}

	// encrypt notification data
	block, err := aes.NewCipher(cryptoKey)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nil, nonce, notificationData, nil)

	// send off our encrypted data
	type encryptedNotificationContainer struct {
		ServerAddress string `json:"server_address"`
		RoutingKey    string `json:"routing_key"`
		IsEncrypted   bool   `json:"is_encrypted"`

		DataType   string `json:"data_type"`
		Ciphertext []byte `json:"ciphertext"` // Golang defaults to encoding binary data to Base64
		IV         []byte `json:"iv"`
	}

	data, err := json.Marshal(encryptedNotificationContainer{
		ServerAddress: serverAddress,
		RoutingKey:    hexRoutingKey,
		IsEncrypted:   true,

		DataType:   "json", // This can be either json or plist, but it's easier to encode in JSON than pull in a whole other library. PList however is implemented on the client, feel free to use it!
		Ciphertext: ciphertext,
		IV:         nonce,
	})

	if err != nil {
		return err
	}

	resp, err := http.Post(fmt.Sprintf("%s/send", NotificationServer), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	type responceData struct {
		Status string
	}

	decodedResp := responceData{}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := json.Unmarshal(bodyBytes, &decodedResp); err != nil {
		return err
	}

	if decodedResp.Status != "sucess" {
		return errors.New(decodedResp.Status)
	}

	return nil
}
