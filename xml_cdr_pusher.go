package main

import (
	"database/sql"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	xj "github.com/basgys/goxml2json"
	"gopkg.in/yaml.v2"

	_ "github.com/lib/pq"
)

type myConfig struct {
	Host     string `yaml:"Host"`
	Database string `yaml:"Database"`
	User     string `yaml:"User"`
	Password string `yaml:"Password"`
	SSL      string `yaml:"SSL"`
}

type myHandler struct {
	Config myConfig
}

func getConfig(name string) (*myConfig, error) {
	Config := myConfig{}
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return &Config, err
	}
	err = yaml.Unmarshal(data, &Config)
	log.Println(Config)
	if err != nil {
		return &Config, err
	}

	return &Config, nil
}

func (h *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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
		Duration                int      `xml:"variables>duration"`
		Billsec                 int      `xml:"variables>billsec"`
		HangupCause             string   `xml:"variables>hangup_cause"`
		HangupCauseQ850         int      `xml:"variables>hangup_cause_q850"`
		Accountcode             string   `xml:"variables>accountcode"`
		ReadCodec               string   `xml:"variables>read_codec"`
		WriteCodec              string   `xml:"variables>write_codec"`
		AlegUUID                string   `xml:"variables>uuid"`
		BlegUUID                string   `xml:"callflow>caller_profile>origination>origination_caller_profile>uuid"`
	}

	body, _ := ioutil.ReadAll(r.Body)
	decbody, _ := url.QueryUnescape(string(body))
	xmlraw := decbody

	xmlreader := strings.NewReader(xmlraw)
	json, jerr := xj.Convert(xmlreader)
	if jerr != nil {
		w.WriteHeader(500)
		w.Write([]byte("500 - " + http.StatusText(503) + "\nJSON conversion failed!"))
		log.Println("JSON conversion failed!", jerr)
		return
	}

	res := &Result{}
	err := xml.Unmarshal([]byte(decbody), res)
	if err != nil {
		w.WriteHeader(503)
		w.Write([]byte("503 - " + http.StatusText(503)))
		log.Println("Parsing XML failed: ", err)
		return
	}

	var StartStamp, AnswerStamp, EndStamp sql.NullString
	StartStamp.String, _ = url.QueryUnescape(res.StartStamp)
	AnswerStamp.String, _ = url.QueryUnescape(res.AnswerStamp)
	EndStamp.String, _ = url.QueryUnescape(res.EndStamp)

	if len(StartStamp.String) == 0 {
		StartStamp.Valid = false
	} else {
		StartStamp.Valid = true
	}
	if len(AnswerStamp.String) == 0 {
		AnswerStamp.Valid = false
	} else {
		AnswerStamp.Valid = true
	}
	if len(EndStamp.String) == 0 {
		EndStamp.Valid = false
	} else {
		EndStamp.Valid = true
	}

	connstring := "postgres://" + h.Config.User + ":" + h.Config.Password + "@" + h.Config.Host + "/" + h.Config.Database + "?sslmode=" + h.Config.SSL
	db, dberr := sql.Open("postgres", connstring)
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
		log.Print("DB Query Failed: ", qerr)
		return
	}

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

	if len(os.Args) < 2 {
		log.Fatal("Please provide a valid config file name as first argument!")
		os.Exit(-1)
	}
	arg := os.Args[1]
	config, err := getConfig(arg)
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

	handler := myHandler{}
	handler.Config = *config

	http.Handle("/", &handler)
	http.ListenAndServe(":8080", nil)
}
