package helpers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	reqURL = "https://gist.githubusercontent.com/nubors/eecf5b8dc838d4e6cc9de9f7b5db236f/raw/d34e1823906d3ab36ccc2e687fcafedf3eacfac9/jne-awb.html"
)

type ResponseData struct {
	Status Status `json:"status"`
	Data   Data   `json:"data"`
}

type Status struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Data struct {
	ReceivedBy string     `json:"receivedBy"`
	Histories  []MetaData `json:"histories"`
}

type MetaData struct {
	Description string             `json:"description"`
	CreatedAt   time.Time          `json:"createdAt"`
	Formatted   FormattedCreatedAt `json:"formatted"`
}

type FormattedCreatedAt struct {
	CreatedAt string `json:"createdAt"`
}

func ResponseFormatter() *ResponseData {
	res, err := http.Get(reqURL)
	if err != nil {
		log.Fatalln("Failed to get data from request url: ", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Parse HTML to text
	tables, err := ConvertFromHTML(res.Body)
	if err != nil {
		log.Fatalln("Failed to convert html to table: ", err)
	}

	// Loop to get the contains from slice of array
	var columns []Column
	for _, table := range tables {
		for _, row := range table {
			for _, column := range row {
				columns = append(columns, column)
			}
		}
	}

	// Loop to get the histories of tracking details
	var histories []Column
	for i, column := range columns {
		if column.ParentValue == "History" {
			histories = columns[i+1:]
		}
	}

	// Loop data to struct
	var responseData ResponseData
	var data Data
	var metaDatas []MetaData
	var receivedBy string

	for i := len(histories) - 1; i >= 0; i-- {
		if strings.Contains(histories[i].ParentValue, "PAK MURADI") {
			receivedBy = "PAK MURADI"
		}
		var metaData MetaData
		// if the index is odd
		if i%2 == 1 {
			//
			var layoutFormat, _, value string
			var date, _ time.Time
			// convert date to WIB
			value = fmt.Sprintf("%s WIB", histories[i-1].ParentValue)

			layoutFormat = "02-01-2006 15:04 MST"
			date, _ = time.Parse(layoutFormat, value)

			// Parse time to 04 Februari 2021, 10:22 WIB
			var dateS1 = date.Format("Monday 02, January 2006 15:04 MST")

			metaData = MetaData{
				// Description is on odd index
				Description: histories[i].ParentValue,
				// Created is on even index
				CreatedAt: date,
				Formatted: FormattedCreatedAt{
					CreatedAt: dateS1,
				},
			}
			metaDatas = append(metaDatas, metaData)
			data = Data{ReceivedBy: receivedBy, Histories: metaDatas}
		}
	}
	responseData = ResponseData{
		Status: Status{
			Code:    "060101",
			Message: "Delivery tracking detail fetched successfully",
		},
		Data: data,
	}

	SaveFileToDisk(responseData)

	return &responseData
}
