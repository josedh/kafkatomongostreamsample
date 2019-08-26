package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
)

type Member struct {
	FirstName       string  `json:"firstName" bson:"firstName"`
	LastName        string  `json:"lastName"  bson:"lastName"`
	PhoneNumber     string  `json:"phoneNumber" bson:"phoneNumber"`
	BirthDate       string  `json:"birthDate" bson:"birthDate"`
	SSN             string  `json:"SSN" bson:"ssn"`
	PatientID       string  `json:"patientID" bson:"patientID"`
	Address         Address `json:"address" bson:"address"`
	Doctor          Doctor  `json:"doctor" bson:"doctor"`
	AppointmentDate string  `json:"appointmentDate" bson:"appointmentDate"`
}

type Address struct {
	Street string `json:"street" bson:"street"`
	City   string `json:"city" bson:"city"`
	Zip    int    `json:"zip" bson:"zip"`
	State  string `json:"state" bson:"state"`
}

type Doctor struct {
	FirstName string  `json:"firstName" bson:"firstName"`
	LastName  string  `json:"lastName" bson:"lastName"`
	Phone     string  `json:"phoneNumber" bson:"phoneNumber"`
	Address   Address `json:"address" bson:"address"`
}

var (
	kafkaWriter *kafka.Writer
	async       bool
)

func init() {
	// Seed the random num gen
	rand.Seed(time.Now().Unix())
}

func main() {
	flag.BoolVar(&async, "async", false, "Used to control async writes to kafka")
	flag.Parse()

	// The boolean value is for whether we want to send async or not
	connKafka(async)

	var (
		fakeMembers *os.File
		doctors     []byte
		err         error
	)
	// Read the file
	if fakeMembers, err = os.Open("./FakeNameGenerator.com_8963e431.csv"); err != nil {
		log.Panicf("Failed to open fake ppl file: %v", err)
	}
	defer fakeMembers.Close()

	reader := csv.NewReader(fakeMembers)
	if doctors, err = ioutil.ReadFile("./doctors.json"); err != nil {
		log.Panicf("Failed to open fake docs file: %v", err)
	}

	var doctorsSlice []Doctor
	if err := json.Unmarshal(doctors, &doctorsSlice); err != nil {
		log.Panicf("Failed to unmarshall docs: %v", err)
	}

	phs := []string{"Walgreens", "CVS", "Harris Teeter", "RiteAid"}

	var count int
	var msgsSent int
	for {
		line, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Errorf("ERROR: %v\n", err)
			break
		}

		// skip first line
		if count == 0 {
			count++
			continue
		}
		count++

		var msg kafka.Message
		if msg, err = getKafkaMessageFromLine(line, phs, doctorsSlice); err != nil {
			log.Errorf("Failed to create valid kafka message from line: %v", err)
		}

		// NOTE: if the variable 'async' above is set to true, this error check WILL NOT
		// work, as the method call is nonblocking. This will significantly increase the
		// speed at which messages are sent, but errors on send will be ignored.
		if err := kafkaWriter.WriteMessages(context.Background(), msg); err != nil {
			log.Errorf("Failed to send message: %v", err)
		}

		msgsSent++
		log.Infof("Total messages sent: %d", count)
	}

	log.Info("KAFKA LOAD DONE!")
}

func getKafkaMessageFromLine(line, pharmacies []string, docsSlice []Doctor) (kafka.Message, error) {
	// lets convert the zip code string to an int val
	var (
		msg        kafka.Message
		marshalled []byte
		zip        int
		err        error
	)
	if zip, err = strconv.Atoi(line[11]); err != nil {
		return msg, err
	}

	// create a sexy looking json message
	if marshalled, err = marshallData(line, docsSlice, zip); err != nil {
		log.Panicf("Failed to json marshal data: %v", err)
	}
	return kafka.Message{
		Key:   []byte(pharmacies[rand.Intn(len(pharmacies))]), // select one of the pharmacies at random
		Value: []byte(marshalled),                             // send a message + curr time
	}, nil

}

func getDate() string {
	t := time.Now()
	var list []string
	list = append(list,
		t.Format("2006-01-02"),
		t.Add(-24*time.Hour).Format("2006-01-02"),
		t.Add(24*time.Hour).Format("2006-01-02"),
		t.Add(48*time.Hour).Format("2006-01-02"),
		t.Add(72*time.Hour).Format("2006-01-02"),
	)

	return list[rand.Intn(len(list))]
}

func connKafka(async bool) {
	kafkaWriter = kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"kafka:9092"},
		Topic:   "medications",
		Async:   async,
	})
}

func marshallData(line []string, docs []Doctor, zip int) ([]byte, error) {
	return json.Marshal(Member{
		FirstName:   line[4],
		LastName:    line[5],
		PhoneNumber: line[14],
		BirthDate:   line[17],
		SSN:         line[19],
		PatientID:   line[0],
		Address: Address{
			Street: line[7],
			City:   line[8],
			Zip:    zip,
			State:  line[9],
		},
		Doctor:          docs[rand.Intn(len(docs))],
		AppointmentDate: getDate(),
	})
}
