package skyglownotificationlib

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Feedback struct {
	Type               string // What type of feedback is this/how should this be handled? Examples are: token-removed
	Reason             string // The reason why this feedback was triggered.
	RoutingToken       []byte // The routing token that is affected
	RoutingTokenServer string // The server that the routing token belongs to.
	CreatedAt          time.Time
}

func GetFeedback(feedback_key []byte, after time.Time) ([]Feedback, error) {
	hexFeedbackKey := hex.EncodeToString(feedback_key)
	afterString := after.Format(time.RFC3339)

	resp, err := http.Get(fmt.Sprintf("%s/get_feedback?feedback_key=%s&after=%s", httpNotificationServer, hexFeedbackKey, afterString))
	if err != nil {
		return nil, err
	}

	type feedbackJson struct {
		RoutingToken  []byte    `json:"routing_token"`
		ServerAddress string    `json:"server_address"`
		Type          int       `json:"type"`
		Reason        string    `json:"reason"`
		CreatedAt     time.Time `json:"created_at"`
	}

	type responceData struct {
		Status string
		Data   []feedbackJson `json:"data"`
	}

	decodedResp := responceData{}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.Unmarshal(bodyBytes, &decodedResp); err != nil {
		return nil, errors.New(string(bodyBytes))
	}

	if decodedResp.Status != "sucess" {
		return nil, errors.New(decodedResp.Status)
	}

	if len(decodedResp.Data) <= 0 {
		return nil, nil
	}

	finalFeedback := make([]Feedback, len(decodedResp.Data))
	for i := range finalFeedback {
		finalFeedback[i] = Feedback{
			Type: func() string {
				switch decodedResp.Data[i].Type {
				case 0:
					return "token-removed"
				default:
					return "unknown"
				}
			}(),
			Reason: decodedResp.Data[i].Reason,
			RoutingToken: func() []byte {
				t, e := hex.DecodeString(string(decodedResp.Data[i].RoutingToken))
				if e != nil {
					return nil
				}
				return t
			}(),
			RoutingTokenServer: decodedResp.Data[i].ServerAddress,
			CreatedAt:          decodedResp.Data[i].CreatedAt,
		}
	}

	return finalFeedback, nil
}

// Sets up a specific token to register feedback, (like when your app is uninstalled, invalidating the token)
//
// The feedback key should be a random array of bytes, that is reused for every token. (needs to be stored)
func ConfigureTokenForFeedback(device_token []byte, feedback_key []byte) error {
	routing_key, serverAddress, err := RoutingInfoFromDeviceToken(device_token)
	if err != nil {
		return err
	}
	hexRoutingKey := hex.EncodeToString(routing_key)

	hexFeedbackKey := hex.EncodeToString(feedback_key)

	type configureForFeedback struct {
		ServerAddress string `json:"server_address"`
		RoutingKey    string `json:"routing_key"`
		FeedbackKey   string `json:"feedback_key"`
	}

	// json encode notification data
	data, err := json.Marshal(configureForFeedback{
		RoutingKey:    hexRoutingKey,
		ServerAddress: *serverAddress,
		FeedbackKey:   hexFeedbackKey,
	})

	if err != nil {
		return err
	}

	resp, err := http.Post(fmt.Sprintf("%s/register_token_for_feedback", httpNotificationServer), "application/json", bytes.NewBuffer(data))
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
		return errors.New(string(bodyBytes))
	}

	if decodedResp.Status != "sucess" {
		return errors.New(decodedResp.Status)
	}

	return nil
}
