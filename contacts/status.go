package contacts

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

const ValueOk = "ok"

type StatusResponse struct {
	Id    string `json:"id"`
	Value string `json:"value"`
}

func WaitForNoDelay(sendGridApiKey string) {
	for {
		delayed, err := isDelayed(sendGridApiKey)
		if err != nil {
			continue
		} else if !delayed {
			break
		}
		time.Sleep(time.Second)
	}
}

func isDelayed(sendGridApiKey string) (bool, error) {
	url := "https://api.sendgrid.com/v3/contactdb/status"
	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return false, err
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sendGridApiKey))
		req.Header.Add("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return false, err
		}
		if res.StatusCode == http.StatusOK {
			var result struct {
				Status []StatusResponse `json:"status"`
			}
			err = json.NewDecoder(res.Body).Decode(&result)
			if err != nil {
				return false, err
			}
			delayed := false
			for _, status := range result.Status {
				if !(status.Value == ValueOk) {
					delayed = true
					break
				}
			}
			return delayed, nil
		} else if res.StatusCode == http.StatusTooManyRequests {
			log.Printf("Over rate limit for %s, retrying\n", url)
			time.Sleep(time.Second)
			continue
		} else {
			return false, errors.New(fmt.Sprintf("Error setting lists, StatusCode: %d", res.StatusCode))
		}
	}
}
