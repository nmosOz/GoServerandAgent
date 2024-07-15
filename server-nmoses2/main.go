package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"regexp"

	"github.com/gorilla/mux"


	//Amazon SDK imports
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Test struct {
	RequestTime string
	StatCode    int
}

type Stat struct {
	Table_name string
	Num_items  int
}

type Item struct {
	SpeciesCode string
	ComName     string
	SciName     string
	LocName     string
	ObsDt       string
	HowMany     int
}

func main() {
	r := mux.NewRouter()
	r.Use(RequestLoggerMiddleware(r))
	r.HandleFunc("/nmoses2/all", allHandler)
	r.HandleFunc("/nmoses2/search", searchHandler)
	r.HandleFunc("/nmoses2/status", testHandler).Methods("GET")
	r.HandleFunc("/", catchAllHandler)
	r.HandleFunc("/{path}", catchAllHandler)
	r.HandleFunc("/{path}/{path}", catchAllHandler)
	r.Handle("/", r)
	http.ListenAndServe(":8080", r)

}

func testHandler(w http.ResponseWriter, r *http.Request) {
	//Read the dynamodb table
	svc := dynamodb.New(session.New())
	info, err := svc.Scan(&dynamodb.ScanInput{
		TableName: aws.String("test-table-nmoses2"),
	})

	if err != nil {
		fmt.Println(err)
	}

	//itemCount := int(*info.Count)

	fmt.Println(int(*info.Count))
	fmt.Println(info)
	var stat Stat
	stat.Table_name = "test-table-nmoses2"
	//dt := time.Now()
	stat.Num_items = int(*info.Count)

	marshalled_struct, err := json.MarshalIndent(stat, "", "	")
	if err != nil {
		fmt.Println(err)
	}

	w.Write([]byte(marshalled_struct))
	//fmt.Println(test)
	return
}

// This handler will return a json representation of all the data in my dynamoDB table
func allHandler(w http.ResponseWriter, r *http.Request) {

	svc := dynamodb.New(session.New())
	info, err := svc.Scan(&dynamodb.ScanInput{
		TableName: aws.String("test-table-nmoses2"),
	})

	if err != nil {
		fmt.Println(err)
	}

	//itemCount := int(*info.Count)

	fmt.Println(int(*info.Count))
	fmt.Println(info)
	var stat Stat
	stat.Table_name = "test-table-nmoses2"
	//dt := time.Now()

	marshalled_struct, err := json.MarshalIndent(info, "", "	")
	if err != nil {
		fmt.Println(err)
	}

	w.Write([]byte(stat.Table_name))
	w.Write([]byte(marshalled_struct))

	return

}

// This handler will return a specific set of data based off of query parameters. (Region code in my case)
func searchHandler(w http.ResponseWriter, r *http.Request) {
	//Create new status response writer
	sw := NewStatusResponseWriter(w)
	//Get the query params
	query := r.URL.Query()

	//Print out query
	(fmt.Printf("\t%+v\n", query))

	//Here i am making sure it is CommonName
	param, present := query["CommonName"]

	//Check to make sure both common name is there AND the length is greater than 0
	if !present || len(param) == 0 {
		fmt.Println("Common name not present")
	}

	//Check length of parameters and the parameters themselves
	fmt.Println(len(param))
	fmt.Println(param)

	//Create new CommonName string for validation
	comName := param[0]
	fmt.Println(comName)

	//Create regexp object with my rules (only 1 or more alphanumeric characters)
	re := regexp.MustCompile("^[a-z A-Z]+$")
	//Validate the query parameter
	if len(comName) > 50 || len(comName) < 2 {
		//Send a 400 code and a short response as to why it was a bad request
		sw.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Common Name length too long or short"))
	} else if !re.MatchString(comName) {
		//Send a 400 code
		sw.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Common name incorrectly formatted"))
	} else {
		//query the dynamo db table
		//Initialize a new session
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))

		//Create dynamo db client
		svc := dynamodb.New(sess)

		//Grab the specified item from the table
		result, err := svc.GetItem(&dynamodb.GetItemInput{
			TableName: aws.String("test-table-nmoses2"),
			Key: map[string]*dynamodb.AttributeValue{
				"comName": {
					S: aws.String(comName),
				},
			},
		})
		//Check to see if there is an error, print it to the console if there is one
		if err != nil {
			log.Fatalf("Got error calling GetItem")
		}

		//Check to see if the resulting item is nil, if it is that means the item could not be found in the table
		if result.Item == nil {
			err_message := "Could not find" + comName
			log.Fatalf(err_message)
		}

		//Create struct to house item
		item := Item{}

		//Unmarshal the resulting item into the struct
		err = dynamodbattribute.UnmarshalMap(result.Item, &item)

		//Check to make sure it unmarshalled correctly
		if err != nil {
			log.Fatalf("Failed to unmarshal the record, err = %v", err)
		}

		//Send the info to the website
		fmt.Println(item)

		//
		marshalled_struct, err := json.MarshalIndent(item, "", "	")
		if err != nil {
			fmt.Println(err)
		}

		w.Write([]byte(marshalled_struct))
	}

}

func catchAllHandler(w http.ResponseWriter, r *http.Request) {
	//...
}

// Status response writer struct
type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// Function to return a pointer of a new status response writer
func NewStatusResponseWriter(w http.ResponseWriter) *statusResponseWriter {
	return &statusResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (sw *statusResponseWriter) WriteHeader(statusCode int) {
	sw.statusCode = statusCode
	sw.ResponseWriter.WriteHeader(statusCode)
}

// Middleware function
func RequestLoggerMiddleware(r *mux.Router) mux.MiddlewareFunc {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			//Set up new status response writer
			sw := NewStatusResponseWriter(w)

			//If the status code is GET make it a 405, else if its not the right path, change it to a 404
			if r.Method != "GET" {
				sw.WriteHeader(http.StatusMethodNotAllowed)
			} else if r.RequestURI != "/nmoses2/status" {
				sw.WriteHeader(http.StatusNotFound)
			}

			//Get info about request
			info := r.Method
			info += "\n" + r.Host + "\n" + r.RequestURI
			info += "\n" + strconv.Itoa(sw.statusCode)
			//Send info to loggly
			var tag string
			tag = "server-nmoses2"
			client := loggly.New(tag)
			err := client.EchoSend("info", info)
			fmt.Println("Error:", err)

			//Do stuff here
			log.Println("Middleware")
			log.Println("URI: ", r.RequestURI)
			log.Println("Source IP:", r.Host)

			//Call the next handler
			next.ServeHTTP(sw, r)
		})
	}

}
