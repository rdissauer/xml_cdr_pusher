package main

import (
	"database/sql"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	_ "github.com/lib/pq"
)

type myHandler struct{}

func (t *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	type Result struct {
		XMLName                 xml.Name `xml:"cdr"`
		Direction               string   `xml:"variables>direction"`
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
		Duration                string   `xml:"variables>duration"`
		Billsec                 string   `xml:"variables>billsec"`
		HangupCause             string   `xml:"variables>hangup_cause"`
		HangupCauseQ850         string   `xml:"variables>hangup_cause_q850"`
		Accountcode             string   `xml:"variables>accountcode"`
		ReadCodec               string   `xml:"variables>read_codec"`
		WriteCodec              string   `xml:"variables>write_codec"`
		AlegUUID                string   `xml:"variables>uuid"`
		BlegUUID                string   `xml:"callflow>caller_profile>origination>origination_caller_profile>uuid"`
	}

	path := r.URL.Path[1:]
	log.Println(path)

	body, _ := ioutil.ReadAll(r.Body)
	decbody, _ := url.QueryUnescape(string(body))
	//log.Println(decbody)

	//decoder := xml.NewDecoder(r.Body)
	res := &Result{}
	//err := decoder.Decode(res)

	err := xml.Unmarshal([]byte(decbody), res)

	//log.Println(err)
	log.Println(*res)

	db, dberr := sql.Open("postgres", "user=test dbname=test password=qakyrGVgyh3Q7ZWwyEgj sslmode=verify-full")
	if dberr != nil {
		log.Fatal(dberr)
	}

	qerr := db.QueryRow(`INSERT INTO xml_cdr (direction, caller_id_name, caller_id_number, called_destination_number, destination_number, context,
		                                        core_uuid, switchname, start_stamp, answer_stamp, end_stamp, duration, billsec,
		                                        hangup_cause, hangup_cause_q850, accountcode, read_codec, write_codec, uuid, bleg, json)
													VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`,
		res.Direction, res.CallerIDName, res.CallerIDNumber, res.CalledDestinationNumber, res.DestinationNumber, res.Context,
		res.CoreUUID, res.Switchname, res.StartStamp, res.AnswerStamp, res.EndStamp, res.Duration, res.Billsec,
		res.HangupCause, res.HangupCauseQ850, res.Accountcode, res.ReadCodec, res.WriteCodec, res.AlegUUID, res.BlegUUID)

	if qerr != nil {
		log.Println(err)
	}

	if err == nil {
		w.WriteHeader(200)
		w.Write([]byte("200 - " + http.StatusText(200)))
	} else {
		w.WriteHeader(503)
		w.Write([]byte("503 - " + http.StatusText(503)))
	}
}

func main() {

	http.Handle("/", new(myHandler))
	http.ListenAndServe(":8080", nil)
}
