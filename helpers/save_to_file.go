package helpers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// func SaveFileToDisk(responseData formatter.ResponseData) {
func SaveFileToDisk(responseData ResponseData) {
	var file, _ = os.Stat("response.json")
	if file != nil {
		log.Println("File already exists")
		return
	}
	newFile, _ := json.MarshalIndent(responseData, "", " ")
	_ = ioutil.WriteFile("response.json", newFile, 0644)
	log.Println("File created successfully")
}
