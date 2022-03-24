
# Kafka Go! 

This is a short article to show you just how easy and simply it is to get up and running with the managed [Apache Kafka](https://developer.aiven.io/docs/products/kafka/index.html) service provided by [Aiven](https://aiven.io/). Aiven offers a fully managed, open source cloud data platform that takes the headache out of managed data infrastructure and lets you focus best on what you do. Of course, getting up and running is the easy part and so with this example I also want to demonstrate how using the building blocks provided by Aiven you can not only get started quickly, but get started with a service that is ready to run in a production deployment.

In our example today we are going to mock a collection of 100 different temperature sensors, sending data to our topic every 5 seconds. Once that is up and running, we will connect our observability components and see the effect of running these services on the underlying service. 

Let's get started. 

## Create your Aiven Account
If you haven't done this already, you will need to [get started with Aiven](https://console.aiven.io/signup.html). 

![Aiven Free Trial](/images/1-signup.png) 

Once you have signed up, your free 30 day trial is up and running and you are good to go.

## Login and Create Your First Service
The Aiven console is pretty self-explanatory when you get started. On the left hand side you will see options that navigate you to different parts of the configuration, the one we are interested in today is the Services option. It should be the default landing when you log in.

![Aiven Services Home](/images/2-home.png) 

You won't see any services if you have just created your account, so lets go ahead and create our very own new Kafka service! Click the big red "Create a new service" button to navigate to the service creation options. When creating a service Aiven has an incredible amount of flexibility over size, performance, cloud provider and even geographical location. For me, today I want to setup 
1. Kafka  
2. on Google Cloud 
3. in the United States. 

So my configuration will reflect this. Notice that as you change this configuration the estimated monthly price is adjusted as well. 

![Creating Kafka Service](/images/3-kafka-settings.png) 

I have just accepted the defaults for setting up this service, but you should scroll down to see the options around plans offered, disk space and to adjust the name of the service if you feel like it. 

Go ahead and click the "Create Service" button. This will take you back to your services home where you can see that the Aiven hamsters are busy building and deploying your Kafka service, it will only take a moment or two. Now might be an excellent time for a cup of tea if you are so inclined. Once you come back, you should see your service up and running. 

![Kafka Running](/images/4-kafkaRunning.png) 

While we are here, flip on the Apache Kafka REST API switch. This will help us check if everything is working in a minute or two. 

Also, navigate to the Topics tab and let's create a topic for us to send some messages to. I simply leave the defaults in the Advanced Configuration section. 

![Kafka Topics](/images/6-kafkaTopics.png) 

## Write (or git clone) the source code
This particular example is using Golang, for no other reason that this was what I felt like using at the time. With Aiven being so simple to use, I felt like I wanted to make something a challenge (full disclosure, I am a Java developer!).

If you aren't interested in writing your own, checkout the code [on Githb](https://github.com/troysellers/gokafka).
Once you have this, you might want to skip ahead to [run the code](#run-and-send-some-messages)

Otherwise, here are some of the more interesting parts of the code. 

### Connection to Kafka
It's just a habbit this days, but I always like to grab my connection strings from my system environment. 

Make sure to import the gotdotenv package 
```go
import (
    ...
    
    github.com/joho/gotenv
    ...
)
```

Then, when the code is initialising load the properties from the .env file
```go
func init () {
	if e := godotenv.Load(); e != nil {
		log.Printf("%v", e)
	}
}
```

The following environment variables are referenced (this is my .env file) 
```
SERVICE_URI=kafka-2ff59bfb-troy-6e20.aivencloud.com:26399
TOPIC:aiven-kafka
KEY_PATH=auth/service.key
CERT_PATH=auth/service.cert
CA_CERT_PATH=auth/ca.pem
```
You will find these things all in the service dashboard of your running Aiven service which you can access by clicking on the newly running Kafka service in the Service home page. 

![Kafka Connection](/images/5-kafkaConnection.png)

Download the Access Key(KEY_PATH), Access Certificate(CERT_PATH) and CA Certificate (CA_CERT_PATH) and store these somewhere the running code can access them. For me, I simply placed them in a directory called "auth" in the root of the project. Of course I don't need to tell you this, but don't check these into your source control :) 

### The Kafka Library
Who wants to write a Kafka library from scratch? Not I thats for sure! 
I choose the [kafka-go](https://github.com/segmentio/kafka-go) module. Why? Well, I spent some time working for Segment and after seeing what they do with event streams, I figured this one must be fairly well battle tested. But hey, you do you :) 

An excellent example of using some of the other Kafka Go modules can be found [here](https://help.aiven.io/en/articles/5344122-go-examples-for-testing-aiven-for-apache-kafka)

### Using the Credentials

The credentials are encapsulated in a struct and loaded by calling a function from main. 
```go
type credentials struct {
	ServiceCertPath string
	ServiceKeyPath  string
	CaCertPath      string
	CaCert          []byte
	KeyPair         tls.Certificate
	X509CertPool    *x509.CertPool
}

func loadCredentials() (*credentials, error) {

	c := &credentials{}
	c.ServiceCertPath = os.Getenv("CERT_PATH")
	c.ServiceKeyPath = os.Getenv("KEY_PATH")
	c.CaCertPath = os.Getenv("CA_CERT_PATH")

	var err error
	c.KeyPair, err = tls.LoadX509KeyPair(c.ServiceCertPath, c.ServiceKeyPath)
	if err != nil {
		return nil, err
	}

	c.CaCert, err = ioutil.ReadFile(c.CaCertPath)
	if err != nil {
		log.Printf("unable to find path %s\n", c.CaCertPath)
		return nil, err
	}

	c.X509CertPool = x509.NewCertPool()
	ok := c.X509CertPool.AppendCertsFromPEM(c.CaCert)
	if !ok {
		log.Fatalf("Failed to parse the CA Certificate file at : %s", c.CaCertPath)
	}

	return c, nil
}
```

###  Create and Send the Messages

In the main() function I create 100 different go routines to simulate 100 different temp sensors sending one message every 5 seconds. Each sensor gets it's own name passed as a function parameter.

```go
func main() {
	c, err := loadCredentials()
	if err != nil {
		log.Fatal(err)
	}
	// run some concurrent processes here to simulate 100 different temp sensors
	var wg sync.WaitGroup
	for i := 1; i <= 100; i++ {
		wg.Add(1)
		go run(c, fmt.Sprintf("sensor-%d", i), &wg)
	}
	wg.Wait()
}
```

Now for each individual sensor, we can loop indefinitely and send a message every five seconds. 
```go
func run(c *credentials, sensor string, wg *sync.WaitGroup) {

	defer wg.Done()

	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
		TLS: &tls.Config{
			Certificates: []tls.Certificate{c.KeyPair},
			RootCAs:      c.X509CertPool,
		},
	}

	producer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{os.Getenv("SERVICE_URI")},
		Topic:   os.Getenv("TOPIC"),
		Dialer:  dialer,
	})
	// TODO : Handle the error this produces
	defer producer.Close()
	// yes, this sensor will run forever!!
	var err error
	for {
		if err = writeMsg(producer, sensor); err != nil {
			log.Fatalf("failed attempting to write message \n %v", err)
		}
		time.Sleep(5 * time.Second)
	}
}
```

And finally, the writing of the random data

```go
/*
	sends a temperature reading to the kafka topic
	temp will be random between 0 - 100
*/
func writeMsg(p *kafka.Writer, sensor string) error {

	t := time.Now()

	// create the message, just some random data for now
	msg := &kafkaMessage{
		Timestamp:   t.Format("2006-01-02T15:04:05-0700"),
		Sensor:      sensor,
		Temperature: rand.Float64() * 100}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	// write the messages to our Kafka topic
	err = p.WriteMessages(context.Background(), kafka.Message{Key: []byte(uuid.New().String()), Value: bytes})

	if err != nil {
		return err
	}

	log.Printf("Sent message\n%v\n", string(bytes))
	return nil
}
```

## Run and Send Some Messages
So you now have the code it is time to run. 

From the root of the project you either built or downloaded run using go run. You should see the following log messages if you have connected and are sending messages. Remember, this needs a CTRL-C to stop. 
```
>go run gokafka.go
2022/03/24 20:46:00 Sent message
{"timestamp":"2022-03-24T20:45:54+1100","sensor":"sensor-25","temperature":57.06732760710226}
2022/03/24 20:46:00 Sent message
{"timestamp":"2022-03-24T20:45:54+1100","sensor":"sensor-10","temperature":9.696951891448457}
2022/03/24 20:46:00 Sent message
{"timestamp":"2022-03-24T20:45:54+1100","sensor":"sensor-15","temperature":36.0871416856906}
2022/03/24 20:46:00 Sent message
{"timestamp":"2022-03-24T20:45:54+1100","sensor":"sensor-80","temperature":21.426387258237494}
2022/03/24 20:46:00 Sent message
{"timestamp":"2022-03-24T20:45:54+1100","sensor":"sensor-48","temperature":71.09071952999952}
2022/03/24 20:46:00 Sent message
{"timestamp":"2022-03-24T20:45:54+1100","sensor":"sensor-87","temperature":49.31419977048804}
```

To view this in the Aiven Kafka Service console, open the Service, go to the Topic and click on the Topic you have sent the messages to. Clicking on "Messages" will allow you to see what has been received by Kafka. (If you don't see this, did you enable the API as mentioned [above?](#connection-to-kafka))

![Kafka Messages](/images/7-kafkaMessages.png)

Leave your message producer running while we setup our observability tooling, this will give us some data to observe!

## Observability
So we can see some messages have arrived but that doesn't really give us much visiblity into the health of our Kafka deployment. It definitely doesn't give us any chance to be monitoring the state and getting alerts, so instead of having one of the team click refresh on the messages button 24-7, lets set up some observability tooling to help us sleep at night. 

Aiven provides [Service Integrations](https://help.aiven.io/en/articles/1456441-getting-started-with-service-integrations) that we are going to use to help here. The idea is we can setup a connection of Kafka metrics => InfluxDB => Grafana to handle all this for us. It's just as simple as the initial configuration of Kafka was! 

### Create an InfluxDB Service.
We first want a time series data store to capture the metrics from Kafka. Back from the Services home page you can create another new service, this time selecting InfluxDB. It's probably best that you locate this in the same cloud and region as your Kafka service. 


![Create InfluxDB](/images/8-influxDb.png)

### Create a Grafana Service
While we wait for our InfluxDB to be provisioned, lets repeat the steps and create a Grafana service.
Again, I accept the sane defaults and locate in the same cloud and region as everything else. 

![Create Grafana](/images/9-grafana.png)

### Connect It All 
Now we need to connect the three services. From the services home open your Kafka service and look for the Service Integrations section (you may have to scroll down) and click the "Manage Integrations" button. You will see a window that lists all the available integrations for Kafka and also that you have none enabled at the moment. Click "Use Integration" on the metrics service. 

![Kafka Integrations](/images/10-kafkaIntegrations.png)

Select the existing InfluxDB service and continue, then close the Service Integrations window. 

Now, from the Services home, open the InfluxDB service and look for Service Integrations again. This time, select the "Dashboards" integration.  Click "Use Integration" and select the Grafana service you previously created.

![Influx Integrations Grafana](/images/11-influxIntegrations.png)

### Go View Your Dashboards

Finally, time to open up Grafan and have a look at all thos metrics we have just enabled. Go to the Services home again and this time, click on the Grafana service.

Here you will see the Service URI that will take you to Grafana, along with the credentials you will need to login to the service. 

![Grafana Service](/images/12-grafanaService.png)

Once logged into Grafana, from the left hand menu go to browse the dashboards and you will see the automatically generated Aiven Kafka dashboard. 

![Grafana Dashboard](/images/13-grafana.png)

Congratulations! In the time it takes to make and eat a sandwich, you deployed a Kafka cluster, created a message producer and setup a complete observability solution to make sure this was production ready for your business.

If you are interested in a bit more depth as to what is on this dashboard, check out this [blog post](https://aiven.io/blog/gain-visibility-with-kafka-dashboards)

For more information about [Aiven Kafka](https://developer.aiven.io/docs/products/kafka/index.html), [Aiven for Grafana](https://developer.aiven.io/docs/products/grafana/index.html) or anything particulary buring questions about [Aiven in general](https://help.aiven.io/en/) click the links! :) Or, reach out to your [local Aiven team](https://aiven.io/contact) as we are always happy to help. 
