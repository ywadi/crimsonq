

# CrimsonQ
[![buddy pipeline](https://app.buddy.works/ywadi85/crimsonq/pipelines/pipeline/373325/badge.svg?token=463c4f343893f85c5056a16ba6da1379079553b6b7a950b7ba9d643591fcb0d2 "buddy pipeline")](https://app.buddy.works/ywadi85/crimsonq/pipelines/pipeline/373325)
## A Multi Consumer per Message Queue with persistence and Queue Stages.
![crimsonq](https://github.com/ywadi/crimsonq/raw/main/assets/logo.png)

 __Early Release: v0.6__
 __Currently functional and usable, actively being developed further__
 
Crimson Queue allows you to have multiple consumers listening on topics and receive a copy of a message that targets its topic. It provides multiple consumers with a message queue and complex message routing. Utilizes the same methodology as MQTT topic routing but with queue stages and persistence per consumer. This is under active development. It is being developed in Go and utilizes [BadgerDB](https://github.com/dgraph-io/badger) as the internal database layer to provide top notch performance, persistence and durability. It supports Redis RESP protocol through [RedCon](https://github.com/tidwall/redcon), allowing any Redis library the ability to connect to it and communicate with it also keeping the performance at, also can be utilized from redis-cli. 
Currently the official client library is being developed for Node.Js on top of node-redis. You can easily develop a client utilizing your redis client of choice on any language you like. Share it with us and we will make sure we get it out there to the community. 
[Link to CrimsonQ Client Library for Nodejs: Under Development](https://github.com/ywadi/crimsonqClient)  

The combination of BadgerDB and RESP protocol allows near 7K message writes a second and higher on reads. It is optimized to be used in the cloud with VPS servers providing SSD storages. Has been stress tested with Docker as well, providing great results. 

## The Concept of CrimsonQ  
The main purpose for the creation of CrimsonQ was that there is no direct way to have a message queue system allow you to have multiple consumers for a single message. The concept of CrimsonQ was born on the idea that we needed a Pub/Sub like system to distribute messages but also need client offline persistence as well as the ability to have stages/statues for the messages (pending, active, delayed, completed, failed). SQS as an example is not going to work for 1 message to be used by multiple consumers, and settings something up for it is a hassle and hard to sustain Where MQTT and REDIS pub/sub provide a fire and forget approach with the message, meaning if your consumer misses a bunch of messages, they are gone.

CrimsonQ allows publishers to target consumers with messages. This is done by either pushing a message to a consumer directly or through an MQTT like topic. 

    Example: 
    3 Consumers are connected to CrimsonQ (ConsumerA, ConsumerB, ConsumerC) 
    - ConsumerA is subscribed to the topic /consumer/a 
    - ConsumerB is subscribed to the topic /consumer/b 
    - ConsumerC is subscribed to the topic /consumer/c
    
    You can now send a message to the consumers' queues by either directly 
    msg.push.consumer ConsumerA "My Message to A directly" 
    Or you can send the same message to multiple consumers to use it by using topic 
    msg.push.topic /consumer/# "My message to consumers"
    Where # in MQTT topics is equivilant to * as a wildcard 
    sending the message to all the subtopics of consumer 
You can find more on how mqtt topic matching works in the [link here](https://www.hivemq.com/blog/mqtt-essentials-part-5-mqtt-topics-best-practices/) 

## Deploying CrimsonQ 
You have 2 options, you either build it with Go or Build a docker image all ready to go. 
### Building and using Docker 
To use docker you can simply use docker-compose that is also added to the project root. This will run the defaut settings. See below for how to setup settings. 

    $crimsonqPath:> docker-compose build
    $crimsonqPath:> docker-compose up 
This will run the server with a docker volume where all data is stored as well as exposing port 9001 to the host machine which can be connected to using redis-cli to test. 

### Building Manually with GO 
To build you will need to go to the project folder and build it as any other go project. Suggested you build and run within the project folder so as to pickup the settings easily. 

    $crimsonqPath:> go build .
    $crimsonqPath:> ./crimsonq 
Make sure you have the right permissions for crimsonq to be able to right to the folder paths identified in the settings file.

## CrimsonQ Settings 
You can setup CrimsonQ to manage your server as needed. You can find the file in the root project as crimson.config. __Make sure you edit the settings before docker-compose build so as to move the settings to the docker container as you want them__
The settings yaml is as follows 

    crimson_settings:
      port: 9001  //The RESP IP address to connect to 
      password: "crimsonQ!" //The password to be used to authenticate 
      ip_whitelist: '*' //IP white list for all OR below 
      ip_whitelist: //This is an alternitave to "*" to be strict on IPs
	      - 127.0.0.1
	      - 198.162.0.12
      heartbeat_seconds: 30 //How many seconds to send consumers a Pub about whats in pending 
      system_db_path: /CrimsonQ/.crimsonQ/dbs //Path for crimsonq system db files will be stored
      data_rootpath: /CrimsonQ/.crimsonQ/dbs //Path for the consumer db data will be stored
      watchdog_seconds: 5 //Seconds for when the watchdog will check for expirations 
      active_before_delayed: 30 //How manys seconds a msg is in active before moved to delayed 
      delayed_before_failed: 30 //How manys seconds a msg is in delayed before moving to failed
      db_full_persist: false // If true all transactions are stored on disk
      disk_sync_seconds: 30 //If db_full_persist=false this will sync the disk every x seconds
      db_detect_conflicts: true //Allows DB to avoid conflicts but lower in performance
      consumer_inactive_destroy_hours: 48 //How many hours of no consumer msg.pull command before destroying consumer queue 
      log_file: /CrimsonQ/.crimsonQ/logs //The file with the log information 

## CrimsonQ connecting and controlling 
To connect to a CrimsonQ server all you need to do is use redis-cli or any other redis client. You can then pass the CrimsonQ commands through the redis client over RESP protocol and execute the commands. You can use those commands to also build your own client library.
You can see the list of commands by connecting through the client and sending __command__ that will return the command list. The list is below in alphabetical order with explanations

- auth [password] 
Used to authenticate to the server
- command []
Retrives a list of commands similar to this
- consumer.create [consumerId] [topics]
Create a new consumer queue, takes a unique id and an mqtt like topic list to recieve messages on 

>     consumer.create myAwesomeConsumer myTopic/Path/Here,myTopic/SecondPath/Here

- consumer.destroy [consumerId]
Destroys a consumer by its consumerId 

- consumer.exists [consumerId]
Checks if a consumer exists, this is useful to check before creating
- consumer.flush.complete [consumerId]
Deletes all messages for a consumer that are in the complete status 
- consumer.flush.failed [consumerId]
Deletes all messages for a consumer that are in the failed status
- consumer.list []
Gets a list of all consumer queues 
- consumer.topics.get [consumerId]
Gets the topics that a consumer is listening on 
- consumer.topics.set [consumerId] [topics]
Sets topics for a consumer to listen on, overrides previous 
- msg.complete [consumerId] [messageId]
Mark a msg as complete, the msg can only be complete if it is either active or pending 
- msg.counts [consumerId]
Get a message count grouped by status for a consumer 
- msg.del [consumerId] [status] [messageId]
Delete a message for a consumer by messageId
- msg.fail [consumerId] [messageId] [errMsg]
Send an active or delayed message to a fail status 
- msg.keys [consumerId]
Get all the message keys for a consumerId, the keys will be prefixed with status
- msg.list.json [consumerId] [status]
Get a list of all messages & details for a consumerId in a given status (ie: pending, active, delayed, completed, failed) in json 
- msg.pull [consumerId]
Pull a message from pending and send it to active, returns the key of the message 
- msg.push.consumer [consumerId] [messageString
Push a message to a consumer directly by its consumer id ex: msg.push.consumer myAwesomeQ "my message text"
- msg.push.topic [topicString] [messageString]
Push a message copy to multiple consumers based on if they match the topic its being sent to. 
- msg.retry [consumerId] [messageId]
Retry a failed message for a consumerId based on its messageId
- msg.retryall [consumerId]
Retry all failed messages for a consumerId
- ping [messageString]
Send a ping message and receive a pong with a message string concatenated to it.
- quit []
Disconnect and quit the RESP client 
- subscribe [consumerId]
Subscribe to a consumerId which receives either a string key of new messages added to the consumer's Q and receive a heartbeat with an array of all pending message ids that need action.

## Building a Client Library 
TODO


