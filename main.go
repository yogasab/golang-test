package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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
	CreatedAt   string             `json:"createdAt"`
	Formatted   FormattedCreatedAt `json:"formatted"`
}

type FormattedCreatedAt struct {
	CreatedAt string `json:"createdAt"`
}

type Column struct {
	ParentName  string `json:"parentName,omitempty"`
	ParentValue string `json:"parentValue,omitempty"`
}

type Row []Column

type Table []Row

func main() {
	// layoutFormat := "04-02-2021 10:22 WIB"

	// 2015-09-02 00:00:00 +0700 WIB
	http.HandleFunc("/", parseTrackingHandler)
	log.Fatal(http.ListenAndServe(":5000", nil))
}

func convertToArray(reader io.ReadCloser) ([]Table, error) {
	tables := []Table{}

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return tables, err
	}

	doc.Find("table").Each(func(_ int, tableSelection *goquery.Selection) {
		table := Table{}

		tableSelection.Find("tr").Each(func(_ int, trSelection *goquery.Selection) {
			row := Row{}

			trSelection.Find("th").Each(func(thi int, thSelection *goquery.Selection) {
				row = append(row, Column{
					ParentName: strings.TrimSpace(thSelection.Text()),
				})
			})

			trSelection.Find("td").Each(func(tdi int, tdSelection *goquery.Selection) {
				if len(row) == 0 || len(row) == tdi {
					row = append(row, Column{
						ParentValue: strings.TrimSpace(tdSelection.Text()),
					})
				} else {
					row[tdi].ParentValue = strings.TrimSpace(tdSelection.Text())
				}
			})

			table = append(table, row)
		})

		tables = append(tables, table)
	})

	return tables, nil
}

func dataFormatter() *ResponseData {
	res, err := http.Get(reqURL)
	if err != nil {
		log.Fatalln("Failed to get data from request url: ", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Parse HTML to text
	tables, err := convertToArray(res.Body)
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
			var layoutFormat, value string
			var date time.Time
			// convert date to WIB
			layoutFormat = "02-01-2006 15:04 MST"
			value = fmt.Sprintf("%s WIB", histories[i-1].ParentValue)
			date, _ = time.Parse(layoutFormat, value)
			//
			metaData = MetaData{
				// Description is on odd index
				Description: histories[i].ParentValue,
				// Created is on even index
				CreatedAt: date.String(),
				Formatted: FormattedCreatedAt{
					CreatedAt: histories[i-1].ParentValue,
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
	return &responseData
}

func parseTrackingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data := dataFormatter()
	switch r.Method {
	case "GET":
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
	case "POST":
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message": "POST method requested"}`))
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Can't find method requested"}`))
	}
}
