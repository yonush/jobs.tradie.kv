package main

//TODO
// - move datastore functionality from handlers.go and replace with abstractions or ORM
import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ostafen/clover"
)

type Jobs struct {
	Jobid     int    `json:"jobid"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Name      struct {
		First string `json:"first"`
		Last  string `json:"last"`
	} `json:"name"`
	Address string   `json:"address"`
	Phone   string   `json:"phone"`
	Email   string   `json:"email"`
	Notes   []string `json:"notes,omitempty"`
}

//import the JSON data into a collection
func loadFromJson(filename string) error {
	objects := make([]map[string]interface{}, 0)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &objects); err != nil {
		return err
	}

	collectionName := strings.TrimSuffix(filepath.Base(filename), ".json")
	db, _ := clover.Open("./data/" + collectionName)
	if err := db.CreateCollection(collectionName); err != nil {
		return err
	}

	for _, obj := range objects {
		doc := clover.NewDocumentOf(obj)
		docs = append(docs, doc)
	}
	db.Insert(collectionName, docs...)
	db.Close()
	return err
}
