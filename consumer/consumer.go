package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	_ "go.mongodb.org/mongo-driver/bson/bsoncodec"
	_ "go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	mongoClient *mongo.Client
	kafkaConn   *kafka.Reader
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

func main() {
	connCtx, connCancel := context.WithTimeout(context.Background(), 3*time.Second)
	var err error
	mongoClient, err = mongo.Connect(connCtx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Panicf("Failed to establish mongo connection: %v", err)
	}
	connCancel()

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 3*time.Second)
	if mongoClient.Ping(pingCtx, readpref.Primary()); err != nil {
		log.Panicf("Failed to connect to mongo: %v", err)
	}
	pingCancel()

	//get collection for inserts
	coll := mongoClient.Database("members").Collection("medications")

	kafkaConn = kafka.NewReader(kafka.ReaderConfig{
		Topic:   "medications",
		Brokers: []string{"kafka:9092"},
	})

	var inserts int
	log.Info("Reading message stream...")
	for {
		msg, err := kafkaConn.ReadMessage(context.Background())
		if err != nil {
			log.Panicf("Failed to read message: %v", err)
		}
		var mem Member
		if err := json.Unmarshal(msg.Value, &mem); err != nil {
			log.Errorf("Failed to unmarshal message: %v", err)
		}
		insertCtx, insertCancel := context.WithTimeout(context.Background(), 2*time.Second)
		if _, err := coll.InsertOne(insertCtx, mem); err != nil {
			log.Errorf("Failed to insert into mongo: %v", err)
		} else {
			inserts++
			log.Infof("Inserts thus far: %d", inserts)
		}
		insertCancel()
	}
}
