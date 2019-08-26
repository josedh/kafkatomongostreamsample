# KAFKA!!!
### Noob starter pack
In order to run this tutorial you will need to make a few minor adjustments to your machine (mac).
First, make sure that docker is installed and running (for this demonstration, we not twerkin' with kubernetes).
Second, make sure Go is installed and your go path set up. Install go by visiting:
https://golang.org/dl/ and installing your applicable package.
Then, in your `~/.bash_profile`, add the following:
```text
export GOPATH=$HOME/go
```
```text
export PATH=$PATH:$(go env GOPATH)/bin
```
Second, pull the following publicly available docker images:
```
$ docker pull spotify/kafka` and `$ docker pull mongo
```
Third install mongo locally
```
$ brew install mongodb
```
Fourth, and optionally, install kafka and zookeeper locally so you can interact with the kafka container from your mac:
```
brew cask install java` && `brew install kafka
```

### Baby steps
We first need to spin up the kafka cluster. In order to do this, we run:
```text
docker run --rm --env ADVERTISED_HOST=localhost --env ADVERTISED_PORT=9092 -p 2181:2181 -p 9092:9092 --name kafka -h kafka spotify/kafka
```
NOTE: the beautify of the spotify/kafka image is that it includes a script to start both: kafka and its required dependency: zookeeper.
So it saves us the trouble of having to run and manage that garbage individually ourselves.
Since we are going to also show a "kafka to mongo" pipeline, also run:
```text
docker run -p 27017:27017 mongo
```
in order to spin up a mongo instance where we can insert documents.
Magically, our kafka cluster is now running, as well as mongo.
We now need to create a topic in the kafka cluster where we are going send/retrieve messages to/from
The two steps needed for that are:
 1. Run `docker run -it [kafka container id] /bin/bash` to get an interactive bash session in the kafka cluster container
 2. Once inside, run `$ $KAFKA_HOME/bin/kafka-topics.sh --create --topic medications --replication-factor 1 --partitions 1 --zookeeper localhost:2181`
 3. Exit by running ```$ exit```.
Alternatively, if you installed kafka in your mac, the two steps above can be skipped and instead you can run:
```text
kafka-topics --create --topic medications --replication-factor 1 --partitions 1 --zookeeper kafka:2181
```
You just created a topic in the cluster, we are now ready for pub/sub to kafka
### Mongo (now we are talking)
Now, we are going to fetch messages from kafka and then insert them into mongo, for this lets create a db inside the mongo container
```text
$ mongo 127.0.0.1:27017/admin
```
then
```text
$ use members
```
to create a new db called members then exit with:
```text
$ exit
```

Before we can actually communicate with kafka from our go code locally, we gotta modify our host file
in your command line type: `$ sudo vim /etc/hosts`.
In the resulting interactive vim session add a new entry as:
`127.0.0.1   localhost kafka`
Exit and save by typing: `:wq <enter>` in your vim session.

Now, `cd` into the consumer directory and **run:
```text
$ go run consumer.go
```
This command fires a go app that connects to the kafka "medications" topic and listens for messages.
When a new message gets pushed, this app will read it, and send it over to the "members" db in the mongo container
for storage in the "medications" collection.
You can open a new terminal session and run:
```text
$ mongo 127.0.0.1:27017/members
```
to enter mongo and then run:
```text
> db.getCollection('medications').count()
```
sporadically to see how many messages get saved to mongo

Finally, `cd` into the producer directory and run the consumer (and mongo store) application with:
```text
$ go run producer.go
```
This command will:
 1. Open a 9000 lines csv file with a list of "people" and their personal info
 2. Open a json file with a list of "doctors"
 3. Combine them to create a json representation of people and their doctors
 4. Send the combo over to kafka for streaming

You can see the data consumed from kafka and stored into mongo by running the following in the mongo terminal session:
```text
> db.getCollection('medications').find({})
```

** Note:
Use either go modules (go.mod) to manage your dependecies.
OR
Run `go get [dependecy path]` to manually install them to your go path (~/go/pkg) on the packages listed by the failure output of running `go run *`

### Profit!
