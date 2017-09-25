package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/bculberson/sendgrid_utils/contacts/csv"
	"github.com/bculberson/sendgrid_utils/contacts/lists"
)

func getListChanges(recipientsForList map[string]lists.Recipient, recipientsInList map[string]lists.Recipient) (map[string]lists.Recipient, map[string]lists.Recipient, map[string]lists.Recipient) {
	removeRecipients := make(map[string]lists.Recipient)
	for key, recipient := range recipientsInList {
		if _, ok := recipientsForList[strings.ToLower(recipient.Email)]; !ok {
			removeRecipients[key] = recipient
		}
	}
	addRecipients := make(map[string]lists.Recipient)
	modifyRecipients := make(map[string]lists.Recipient)
	for key, rIn := range recipientsForList {
		if rThere, ok := recipientsInList[strings.ToLower(rIn.Email)]; !ok {
			addRecipients[key] = rIn
		} else {
			if rIn.FirstName != rThere.FirstName || rIn.LastName != rThere.LastName {
				modifyRecipients[key] = rIn
			}
		}
	}

	return removeRecipients, addRecipients, modifyRecipients
}

func findField(names []string, fields []string) int {
	fieldPos := -1
	for ix, field := range fields {
		for _, name := range names {
			if strings.ToLower(field) == name {
				fieldPos = ix
				break
			}
		}
	}
	return fieldPos
}

func getEmailsFromCsv(fields []string, records [][]string) (map[string]lists.Recipient, error) {
	results := make(map[string]lists.Recipient, 0)
	emailField := findField([]string{"email", "e-mail"}, fields)
	if emailField == -1 {
		return nil, errors.New("Email field not found in csv")
	}
	firstNameField := findField([]string{"first_name", "firstname", "fname"}, fields)
	lastNameField := findField([]string{"last_name", "lastname", "lname"}, fields)

	for _, record := range records {
		key := strings.ToLower(record[emailField])
		r := lists.Recipient{Email: record[emailField]}
		if firstNameField >= 0 {
			r.FirstName = record[firstNameField]
		}
		if lastNameField >= 0 {
			r.LastName = record[lastNameField]
		}
		results[key] = r
	}
	return results, nil
}

func listSyncCommand() {

	wg := sync.WaitGroup{}
	wg.Add(2)
	var err error

	recipientsInList := make(map[string]lists.Recipient, 0)
	go func() {
		var recipients []lists.Recipient
		recipients, err = lists.GetListRecipients(*syncListListId, sendGridApiKey)
		if err != nil {
			log.Fatalf("Error getting data list %v:", err.Error())
		}
		for _, recipient := range recipients {
			recipientsInList[strings.ToLower(recipient.Email)] = recipient
		}

		wg.Done()
	}()

	recipientsForList := make(map[string]lists.Recipient, 0)
	go func() {
		fields, records, err := csv.ParseCsv(*syncListCsvFile)
		if err != nil {
			log.Fatalf("Error parsing csv: %v", err.Error())
		}
		recipientsForList, err = getEmailsFromCsv(fields, records)
		if err != nil {
			log.Fatalf("Error getting emails from csv: %v", err.Error())
		}
		wg.Done()
	}()

	wg.Wait()

	fmt.Println(fmt.Sprintf("Recipients in csv: %d", len(recipientsForList)))
	fmt.Println(fmt.Sprintf("Recipients in list: %d", len(recipientsInList)))

	remove, add, modify := getListChanges(recipientsForList, recipientsInList)
	fmt.Println(fmt.Sprintf("Removing %d recipients in list", len(remove)))
	fmt.Println(fmt.Sprintf("Adding %d recipients to list", len(add)))
	fmt.Println(fmt.Sprintf("Modifying %d recipients in list", len(modify)))

	if !*debug {
		wg.Add(3)
		go func() {
			recipientsToDelete := make([]string, 0)
			for _, recipient := range remove {
				recipientsToDelete = append(recipientsToDelete, recipient.Email)
			}
			err := lists.RemoveListRecipients(recipientsToDelete, *syncListListId, sendGridApiKey)
			if err != nil {
				log.Fatalf("Error removing recipients from list: %v", err.Error())
			}
			wg.Done()
		}()
		go func() {
			recipientsToAdd := make([]lists.Recipient, 0)
			for _, recipient := range add {
				recipientsToAdd = append(recipientsToAdd, recipient)
			}
			err = lists.AddRecipientsToList(recipientsToAdd, *syncListListId, sendGridApiKey)
			if err != nil {
				log.Fatalf("Error adding recipients to list: %v", err.Error())
			}
			wg.Done()
		}()
		go func() {
			recipientsToModify := make([]lists.Recipient, 0)
			for _, recipient := range modify {
				recipientsToModify = append(recipientsToModify, recipient)
			}
			err = lists.AddRecipientsToList(recipientsToModify, *syncListListId, sendGridApiKey)
			if err != nil {
				log.Fatalf("Error modifying recipients in list: %v", err.Error())
			}
			wg.Done()
		}()
		wg.Wait()
	}
}
