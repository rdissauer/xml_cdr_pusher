package main

import (
	"encoding/xml"
	"log"
	"net/http"
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

	decoder := xml.NewDecoder(r.Body)
	res := &Result{}
	err := decoder.Decode(res)

	//log.Println(err)
	log.Println(*res)

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
