
# Kafka Go! 

This is a short article to show you just how easy and simply it is to get up and running with the managed [Apache Kafka](https://developer.aiven.io/docs/products/kafka/index.html) service provided by [Aiven](https://aiven.io/). Aiven offers a fully managed, open source cloud data platform that takes the headache out of managed data infrastructure and lets you focus best on what you do. 

In our example today we are going to start up a simple service that send messages to a topic, then we are going to look at some tooling that allows us to observe the Kafka deployment in action. This is a temperature sensor, reading a temp every 5 seconds and sending this event to our service, a fairly trivial example.

Let's get started. 

## Create your Aiven Account

## Add your first Kafka Service

## Grab (or write your own) the source code
This particular example is using Golang, for no other reason that this was what the author felt like using at the time. With Aiven being so simple to use, I felt like I wanted to make something a challenge (full disclosure, I am a Java developer!).

If you aren't interested in writing your own, checkout the code [on Githb](https://github.com/troysellers/gokafka)

Otherwise, lets get started. 

### Housekeeping
It's just a habbit this days, but I always like to grab my connection strings from my system environment. 

Make sure to import the gotdotenv package 
```
import (
    ...
    
    github.com/subosito/gotenv
    ...
)
```

Then, when the code is initialising load the properties from the .env file
```
func init () {
    if err := gotdotenv.Load(); err != nil {
        log.Fatal(err)
    }
}
```

### The Kafka Library
Who wants to write a Kafka library from scratch? Not I thats for sure! 
There is some good information in the [Motiviations](https://github.com/segmentio/kafka-go) section for using the kafka-go library maitained by the awesome team at Segment which I will leave as an exercise for the reader, but this is library I chose to use. 

Connecting to Kafka
```
	p, err := strconv.Atoi(os.Getenv("PARTITION"))
	if err != nil {
		return err
	}
	ak := &aivenKafka{
		topic:     "aiven-topic",
		partion:   0,
		host:      os.Getenv("HOST"),
		partition: p,
	}
```

