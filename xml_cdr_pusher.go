package main

import (
	"database/sql"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"strconv"
	xj "github.com/basgys/goxml2json"

	_ "github.com/lib/pq"
)

type myHandler struct{}

func (t *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	type Result struct {
		XMLName                 xml.Name `xml:"cdr"`
		Direction               string   `xml:"variables>call_direction"`
		CallerIDName            string   `xml:"variables>caller_id_name"`
		CallerIDNumber          string   `xml:"variables>caller_id_number"`
		CalledDestinationNumber string   `xml:"callflow>caller_profile>destination_number"`
		DestinationNumber       string   `xml:"callflow>caller_profile>origination>origination_caller_profile>destination_number"`
		Context                 string   `xml:"callflow>caller_profile>context"`
		CoreUUID                string   `xml:"core-uuid,attr"`
		Switchname              string   `xml:"switchname,attr"`
		StartStamp              string   `xml:"variables>start_stamp"`
		AnswerStamp             string   `xml:"variables>answer_stamp"`
		EndStamp                string   `xml:"variables>end_stamp"`
		Duration                int   `xml:"variables>duration"`
		Billsec                 int   `xml:"variables>billsec"`
		HangupCause             string   `xml:"variables>hangup_cause"`
		HangupCauseQ850         int   `xml:"variables>hangup_cause_q850"`
		Accountcode             string   `xml:"variables>accountcode"`
		ReadCodec               string   `xml:"variables>read_codec"`
		WriteCodec              string   `xml:"variables>write_codec"`
		AlegUUID                string   `xml:"variables>uuid"`
		BlegUUID                string   `xml:"callflow>caller_profile>origination>origination_caller_profile>uuid"`
	}

	//path := r.URL.Path[1:]
	//log.Println(path)

	body, _ := ioutil.ReadAll(r.Body)
	decbody, _ := url.QueryUnescape(string(body))
	//log.Println(decbody)
	//decbody, _ = url.QueryUnescape(decbody)
	//log.Println(decbody)
	xmlraw := decbody	
	xmlreader := strings.NewReader(xmlraw)
	json, jerr := xj.Convert(xmlreader)
	if jerr != nil {
		w.WriteHeader(500)
                w.Write([]byte("500 - " + http.StatusText(503) + "\nJSON conversion failed!"))
		log.Println("JSON conversion failed!")
		return
	}
	//log.Println(json.String())

	//decoder := xml.NewDecoder(r.Body)
	res := &Result{}
	//err := decoder.Decode(res)

	err := xml.Unmarshal([]byte(decbody), res)

	if err != nil {
		w.WriteHeader(503)
		w.Write([]byte("503 - " + http.StatusText(503)))
		log.Print("Parsing XML failed: ")
		log.Println(err);
		return
	}

	var StartStamp, AnswerStamp, EndStamp sql.NullString

	StartStamp.String, _ = url.QueryUnescape(res.StartStamp)
	AnswerStamp.String, _ = url.QueryUnescape(res.AnswerStamp)
	EndStamp.String, _ = url.QueryUnescape(res.EndStamp)

	if len(StartStamp.String) == 0 {StartStamp.Valid = false } else {StartStamp.Valid = true }
	if len(AnswerStamp.String) == 0 {AnswerStamp.Valid = false } else {AnswerStamp.Valid = true }
	if len(EndStamp.String) == 0 {EndStamp.Valid = false } else {EndStamp.Valid = true }

	//log.Println(err)
	//log.Println(*res)

	db, dberr := sql.Open("postgres", "user=test dbname=test password=qakyrGVgyh3Q7ZWwyEgj sslmode=require")
	if dberr != nil {
		w.WriteHeader(500)
                w.Write([]byte("500 - " + http.StatusText(503) + "\nFailed to connect to DB!"))
		log.Fatal(dberr)
		return
	}
	defer db.Close()

	_, qerr := db.Query(`INSERT INTO xml_cdr (direction, caller_id_name, caller_id_number, called_destination_number, destination_number, context,
		                                        core_uuid, switchname, start_stamp, answer_stamp, end_stamp, duration, billsec,
		                                        hangup_cause, hangup_cause_q850, accountcode, read_codec, write_codec, uuid, bleg, json)
													VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`,
		res.Direction, res.CallerIDName, res.CallerIDNumber, res.CalledDestinationNumber, res.DestinationNumber, res.Context,
		res.CoreUUID, res.Switchname, StartStamp, AnswerStamp, EndStamp, res.Duration, res.Billsec,
		res.HangupCause, res.HangupCauseQ850, res.Accountcode, res.ReadCodec, res.WriteCodec, res.AlegUUID, res.BlegUUID, json.String())

	if qerr != nil {
		w.WriteHeader(500)
                w.Write([]byte("500 - " + http.StatusText(503) + "\nDB Query failed!"))
		log.Print("DB Query Failed: ")
		log.Println(qerr)
		return
		//log.Println(qerr.Error())
		//if qerr, ok := qerr.(*pq.Error); ok {
		//	fmt.Println("pq error:", qerr.Code.Name())
		//}
	}

	//log.Println(row)

	if err == nil {
		w.WriteHeader(200)
		w.Write([]byte("200 - " + http.StatusText(200)))
		log.Println("Success, Inserted: " + res.CallerIDName + " -> " + res.DestinationNumber + " -> " + res.Direction + " Accountcode: " + res.Accountcode + " Billsec: " + strconv.Itoa(res.Billsec))
	} else {
		w.WriteHeader(503)
		w.Write([]byte("503 - " + http.StatusText(503)))
		log.Println("Failed")
	}
}

func main() {

	http.Handle("/", new(myHandler))
	http.ListenAndServe(":8080", nil)
}
