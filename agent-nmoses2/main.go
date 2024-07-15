// This package is used to test the Loggly package
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Sighting struct {
	SpeciesCode string `json:"speciesCode"`
	ComName     string `json:"comName"`
	SciName     string `json:"sciName"`
	LocName     string `json:"locName"`
	ObsDt       string `json:"obsDt"`
	HowMany     int    `json:"howMany"`
}

func main() {

	var tag string
	tag = "agent-nmoses2"

	interval := flag.Int("i", 60, "ticker interval")

	urls := [10]string{"US", "NZ", "AU", "CA", "GB", "SE", "IT", "JP", "HK", "RU"}

	//Check the cmd line for the flag
	flag.Parse()

	ticker := time.NewTicker(time.Minute * time.Duration(*interval))
	// Instantiate the client
	client := loggly.New(tag)
	fmt.Println("Interval:", *interval)
	i := 0

	var Sightings []Sighting

	for range ticker.C {

		//Set up the URL
		url := "https://api.ebird.org/v2/data/obs/" + urls[i] + "/recent?maxResults=5"

		//Send a get request to
		req, err := http.NewRequest("GET", url, nil)

		//Add the auth key to the header of the request
		req.Header.Add("x-ebirdapitoken", "7uud8o6kg1n1")

		//Check to make sure there was no error in the request
		if err != nil {
			panic(err)
		}

		//send request using client
		newClient := &http.Client{}
		resp, err := newClient.Do(req)

		//Check to make sure there was not error in the response
		if err != nil {
			log.Println("Error on response:", err)
		}

		//Read the body portion of the GET response
		body, err := io.ReadAll(resp.Body)

		//Check for any errors on readying the body
		if err != nil {
			log.Println("Error on reading:", err)
		}

		//Send the number of bytes as a string to loggly
		err = client.EchoSend("info", strconv.Itoa(len(body)))
		fmt.Println("err:", err)

		//Unmarsharl the GET request and store it in the Go Struct 'Sightings' array
		err = json.Unmarshal([]byte(body), &Sightings)

		//Check for errors, if there are any, print them to the screen
		if err != nil {
			log.Println("Error in unmarshal:", err)
		}

		//If you are at the end of the array, reset to 0, else increment the counter
		if i == 9 {
			i = 0
		} else {
			i++
		}

		//Print out the array of sightings to the screen
		//fmt.Printf("Info : \n%+v\n", Sightings)

		//Write a Go Struct to dyanamoDB
		//Init a session that the SDK will use to load
		//Credentials from the shared credentials file ~/.asw/credentials
		//and region from the shared config file /.aws/config
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))

		//Create dynamo db client
		svc := dynamodb.New(sess)

		//Specify name of the table
		tableName := "test-table-nmoses2"

		//Loop through each struct in the Sightings array and put each in the dynamo db table
		for j := 0; j < len(Sightings); j++ {

			//Map to an item that can be put in dynamoDB
			av, err := dynamodbattribute.MarshalMap(Sightings[j])
			fmt.Printf("AV:\t%+v\n0", av)

			//Check to see if there was an error with marshalling the sighting struct
			if err != nil {
				log.Fatalf("Got error marshalling new Sighting item: %s", err)
			}

			//Specify the input item and what table to put the item in
			input := &dynamodb.PutItemInput{
				//Item we can to put in
				Item: av,
				//Ensure the table name can be represented as a string
				TableName: aws.String(tableName),
			}

			//Put item, do not care about first return value
			_, err = svc.PutItem(input)

			//Check to see if there is an error in PutItem
			if err != nil {
				log.Fatalf("Got error calling PutItem: %s", err)
			}

			//Print out message if successfully added
			fmt.Println("Successfully Added")

			//Close the request
			resp.Body.Close()
		}
	}
}
