package main

import (
	"log"
	"os"

	"fmt"
	"github.com/bculberson/sendgrid_utils/contacts/lists"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app   = kingpin.New("sg_utils", "command-line utilites for sendgrid")
	debug = kingpin.Flag("debug", "Enable debug mode.").Short('d').Default("false").Bool()

	syncList        = app.Command("syncList", "Sync an existing list with csv")
	syncListCsvFile = syncList.Arg("inputFile", "Input file (csv)").Required().ExistingFile()
	syncListListId  = syncList.Arg("ListId", "List Id").Required().Int()

	createList = app.Command("createList", "Create a list")
	listName   = createList.Arg("name", "Name for list").Required().String()

	deleteList           = app.Command("deleteList", "Delete a list")
	deleteListRecipients = deleteList.Flag("deleteRecipients", "Delete the recipients from database").Default("false").Bool()
	deleteListListId     = deleteList.Arg("id", "Id for list").Required().Int()

	exportLists             = app.Command("exportLists", "Get all lists as csv")
	exportListsCsvFile      = exportLists.Arg("outputFile", "Output file (csv)").Required().String()

	sendGridApiKey = ""
)

func init() {
	sendGridApiKey = os.Getenv("SENDGRID_API_KEY")
	if sendGridApiKey == "" {
		log.Fatal("SENDGRID_API_KEY not set in env")
	}
}

func main() {
	kingpin.Version("0.0.1")

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case syncList.FullCommand():
		listSyncCommand()
	case createList.FullCommand():
		id, err := lists.CreateList(*listName, sendGridApiKey)
		if err != nil {
			log.Fatalf("Error creating list: %v", err.Error())
		}
		fmt.Printf("%d\n", id)
	case deleteList.FullCommand():
		err := lists.DeleteList(*deleteListListId, *deleteListRecipients, sendGridApiKey)
		if err != nil {
			log.Fatalf("Error deleting list: %v", err.Error())
		}
	case exportLists.FullCommand():
		lists, err := lists.GetLists(sendGridApiKey)
		if err != nil {
			log.Fatalf("Error getting lists: %v", err.Error())
		}
		//export
		f, err := os.Create(*exportListsCsvFile)
		if err != nil {
			log.Fatalf("Error creating file: %v", err.Error())
		}
		defer f.Close()
		f.WriteString("id,name,count\n")
		for _, list := range lists {
			f.WriteString(fmt.Sprintf("%d,\"%s\",%d\n",list.Id,list.Name,list.Count))
		}
	}
}
