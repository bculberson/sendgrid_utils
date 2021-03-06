package lists

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bculberson/sendgrid_utils/contacts"
)

const pageSize = 1000
const batchSize = 1000

type CustomField struct {
	Id    int         `json:"id"`
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type Recipient struct {
	Email        string        `json:"email"`
	FirstName    string        `json:"first_name"`
	LastName     string        `json:"last_name"`
	Id           string        `json:"id"`
	UpdatedAt    int64         `json:"updated_at"`
	CreatedAt    int64         `json:"created_at"`
	LastOpened   *int64        `json:"last_opened"`
	LastClicked  *int64        `json:"last_clicked"`
	LastEmailed  *int64        `json:"last_clicked"`
	CustomFields []CustomField `json:"custom_fields"`
}

type recipientIn struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

type List struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"recipient_count"`
}

func GetLists(sendGridApiKey string) ([]List, error) {
	url := "https://api.sendgrid.com/v3/contactdb/lists"
	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sendGridApiKey))
		req.Header.Add("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode == http.StatusOK {
			var result struct {
				Lists []List `json:"lists"`
			}
			err = json.NewDecoder(res.Body).Decode(&result)
			if err != nil {
				return nil, err
			}
			return result.Lists, nil
		} else if res.StatusCode == http.StatusTooManyRequests {
			log.Printf("Over rate limit for %s, retrying\n", url)
			time.Sleep(time.Second)
			continue
		} else {
			return nil, errors.New(fmt.Sprintf("Error setting lists, StatusCode: %d", res.StatusCode))
		}
	}
}

func CreateList(name string, sendGridApiKey string) (int, error) {
	create := struct {
		Name string `json:"name"`
	}{name}
	url := "https://api.sendgrid.com/v3/contactdb/lists"
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(&create)
	if err != nil {
		return 0, err
	}

	for {
		req, err := http.NewRequest("POST", url, b)
		if err != nil {
			return 0, err
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sendGridApiKey))
		req.Header.Add("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0, err
		}
		if res.StatusCode == http.StatusCreated {
			var result struct {
				Id int `json:"id"`
			}
			err = json.NewDecoder(res.Body).Decode(&result)
			if err != nil {
				return 0, err
			}
			return result.Id, nil
		} else if res.StatusCode == http.StatusTooManyRequests {
			log.Printf("Over rate limit for %s, retrying\n", url)
			time.Sleep(time.Second)
			continue
		} else {
			return 0, errors.New(fmt.Sprintf("Error creating list, StatusCode: %d", res.StatusCode))
		}
	}
}

func DeleteList(listId int, deleteRecipients bool, sendGridApiKey string) error {
	url := fmt.Sprintf("https://api.sendgrid.com/v3/contactdb/lists/%d", listId)
	if deleteRecipients {
		url = url + "?delete_contacts=true"
	}
	for {
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			return err
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sendGridApiKey))
		req.Header.Add("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		if res.StatusCode == http.StatusAccepted {
			contacts.WaitForNoDelay(sendGridApiKey)
			return nil
		} else if res.StatusCode == http.StatusTooManyRequests {
			log.Printf("Over rate limit for %s, retrying\n", url)
			time.Sleep(time.Second)
			continue
		} else {
			return errors.New(fmt.Sprintf("Error deleting list, StatusCode: %d", res.StatusCode))
		}
	}
}

func AddRecipientsToList(recipients []Recipient, listId int, sendGridApiKey string) error {
	batch := make([]Recipient, 0)
	for _, recipient := range recipients {
		batch = append(batch, recipient)
		if len(batch) == batchSize {
			err := addRecipientsToList(batch, listId, sendGridApiKey)
			if err != nil {
				return err
			}
			batch = make([]Recipient, 0)
		}
	}
	if len(batch) > 0 {
		err := addRecipientsToList(batch, listId, sendGridApiKey)
		if err != nil {
			return err
		}
	}
	return nil
}

func addRecipientsToList(recipients []Recipient, listId int, sendGridApiKey string) error {
	url := fmt.Sprintf("https://api.sendgrid.com/v3/contactdb/recipients?list_id=%d", listId)

	// generate body
	recipientsIn := make([]recipientIn, 0)
	for _, recipient := range recipients {
		recipientsIn = append(recipientsIn, recipientIn{Email: recipient.Email, FirstName: recipient.FirstName, LastName: recipient.LastName})
	}
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(&recipientsIn)
	if err != nil {
		return err
	}

	for {
		req, err := http.NewRequest("POST", url, b)
		if err != nil {
			return err
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sendGridApiKey))
		req.Header.Add("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		if res.StatusCode == http.StatusCreated {
			contacts.WaitForNoDelay(sendGridApiKey)
			return nil
		} else if res.StatusCode == http.StatusTooManyRequests {
			log.Printf("Over rate limit for %s, retrying\n", url)
			time.Sleep(time.Second)
			continue
		} else {
			return errors.New(fmt.Sprintf("Error adding recipients to list, StatusCode: %d", res.StatusCode))
		}
	}
}

func RemoveListRecipients(recipients []Recipient, listId int, sendGridApiKey string) error {
	batch := make([]Recipient, 0)
	for _, recipient := range recipients {
		batch = append(batch, recipient)
		if len(batch) == batchSize {
			err := removeListRecipients(batch, listId, sendGridApiKey)
			if err != nil {
				return err
			}
			batch = make([]Recipient, 0)
		}
	}
	if len(batch) > 0 {
		err := removeListRecipients(batch, listId, sendGridApiKey)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeListRecipients(recipients []Recipient, listId int, sendGridApiKey string) error {
	ids := make([]string, len(recipients))
	for ix, r := range recipients {
		ids[ix] = base64.StdEncoding.EncodeToString([]byte(strings.ToLower(r.Email)))
	}
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(&ids)
	if err != nil {
		return err
	}

	for {
		url := fmt.Sprintf("https://api.sendgrid.com/v3/contactdb/lists/%d/recipients", listId)
		req, err := http.NewRequest("DELETE", url, b)
		if err != nil {
			return err
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sendGridApiKey))
		req.Header.Add("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		if res.StatusCode == http.StatusNoContent {
			contacts.WaitForNoDelay(sendGridApiKey)
			return nil
		} else if res.StatusCode == http.StatusTooManyRequests {
			log.Printf("Over rate limit for %s, retrying\n", url)
			time.Sleep(time.Second)
			continue
		} else {
			return errors.New(fmt.Sprintf("Error deleting recipient from list, StatusCode: %d", res.StatusCode))
		}
	}
}

func GetListRecipients(listId int, sendGridApiKey string) ([]Recipient, error) {
	page := 1
	result := make([]Recipient, 0)
	for {
		url := fmt.Sprintf("https://api.sendgrid.com/v3/contactdb/lists/%d/recipients?page_size=%d&page=%d", listId, pageSize, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sendGridApiKey))
		req.Header.Add("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode == http.StatusNotFound {
			break
		} else if res.StatusCode == http.StatusOK {
			var recipients struct {
				Recipients []Recipient `json:"recipients"`
			}

			err = json.NewDecoder(res.Body).Decode(&recipients)
			if err != nil {
				return nil, err
			}
			result = append(result, recipients.Recipients...)
		} else if res.StatusCode == http.StatusTooManyRequests {
			log.Printf("Over rate limit for %s, retrying\n", url)
			time.Sleep(time.Second)
			continue
		} else {
			return nil, errors.New(fmt.Sprintf("Error retrieving list from SendGrid, StatusCode: %d", res.StatusCode))
		}
		page = page + 1
	}
	return result, nil
}
