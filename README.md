

# CrimsonQ
![GitHub release (latest by date)](https://img.shields.io/github/v/release/ywadi/crimsonq)
![GitHub](https://img.shields.io/github/license/ywadi/crimsonq)
![GitHub](https://img.shields.io/badge/Built%20On-Golang-lightgrey?logo=go)
[![Sockerhub](https://img.shields.io/badge/DockerHub-crimsonq-blue?logo=docker)](https://hub.docker.com/r/yousefjwadi/crimsonq)


## A Multi Consumer per Message with persistence and Queue Stages.
![crimsonq](https://github.com/ywadi/crimsonq/raw/main/assets/logo.png)

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

Below is an animation explaining the concept for visual ease 

![crimsonq](https://github.com/ywadi/crimsonq/raw/main/assets/anim.gif)

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
    HTTP:
	  enabled: true //Enable or disable HTTP REST interface
	  port: 8080 //Port for the HTTP to listen on 
      ip_whitelist: '*' //IP white list for all OR below 
      ip_whitelist: //This is an alternitave to "*" to be strict on IPs
	      - 127.0.0.1
	      - 198.162.0.12	  
	  username: "crimsonQ" //Basic Auth to connect by username
	  password: "crimsonQ!" //Password for basic auth username 
	RESP:
	  port: 9001 //Port for where the RESP interface will work 
	  password: "crimsonQ!" //RESP password to connect 
      ip_whitelist: '*' //IP white list for all OR below 
      ip_whitelist: //This is an alternitave to "*" to be strict on IPs
	      - 127.0.0.1
	      - 198.162.0.12	  
      heartbeat_seconds: 30 //How many seconds to send consumers a Pub about whats in pending 


## CrimsonQ connecting and controlling 
To connect to a CrimsonQ server all you need to do is use redis-cli or any other redis client. You can then pass the CrimsonQ commands through the redis client over RESP protocol and execute the commands. You can use those commands to also build your own client library.
You can see the list of commands by connecting through the client and sending __command__ that will return the command list. The list is below in alphabetical order with explanations

Connect to default settings after running the CrimsonQ server like this 

    redis-cli -p 9001 -a crimsonQ!

### Connecting through RESP (redis protocol) 
You can execute the following commands over the RESP protocol, for testing use redis-cli.
<table border="1" cellpadding="1" cellspacing="1" style="width:500px;">
	<tbody>
		<tr>
			<td><b>Command</b></td>
			<td><b>Arguments<b/></td>
			<td><b>Return<b/></td>
			<td><b>Notes<b/></td>
		</tr>
		<tr>
			<td>AUTH</td>
			<td><i>password [string]</i></td>
			<td> </td>
			<td>&nbsp;</td>
		</tr>
		<tr>
			<td>COMMAND</td>
			<td><i>no arguments</i></td>
			<td>[String Array] of commands and argumetns in the format ["cmd_name [arg1] [arg2] [arg3]"] </td>
			<td>&nbsp;</td>
		</tr>
		<tr>
			<td>CONSUMER.CREATE</td>
			<td><ol><li>consumerId [string]</li><li>topics [string]</li><li>concurrency [int]</li></ol></td>
			<td></td>
			<td>Topics are passed over as a string with comma separated. eg: "/path1/p2,/topic1/t2" <br/> </td>
		</tr> 
		<tr>
			<td>CONSUMER.DESTROY</td>
			<td>consumerId [string]</td>
			<td></td>
			<td>Destroies the consumer and all the data.</td>
		</tr>
		<tr>
			<td>CONSUMER.EXISTS</td>
			<td>consumerId [string]</td>
			<td>string "true" or "false"</td>
			<td>Checks if consumer exists or not.</td>
		</tr>
		<tr>
			<td>CONSUMER.FLUSH.FAIL</td>
			<td>consumerId [string]</td>
			<td></td>
			<td>Deletes all failed messages for consumerId</td>
		</tr>
		<tr>
			<td>CONSUMER.FLUSH.COMPLETE</td>
			<td>consumerId [string]</td>
			<td></td>
			<td>Deletes all complete messages for consumerId</td>
		</tr>
		<tr>
			<td>CONSUMER.LIST</td>
			<td></td>
			<td>[String Array] of all consumers</td>
			<td>Returns a list of all the consumers that have queues. </td>
		</tr>
		<tr>
			<td>CONSUMER.TOPICS.GET</td>
			<td>consumerId [string]</td>
			<td>[String Array]</td>
			<td>Returns a list of topics that a consumer is subscribed to get messages from. </td>
		</tr>
		<tr>
			<td>CONSUMER.TOPICS.SET</td>
			<td><ol><li>consumerId [string]</li><li>topics [string]: comma separate in string</li></ol></td>
			<td></td>
			<td>Sets the list of topics that a consumer is subscribed to get messages on. </td>
		</tr>
		<tr>
			<td>MSG.COMPLETE</td>
			<td><ol><li>consumerId [string]</li><li>messageKey[string]</li></ol></td>
			<td>[String Array]</td>
			<td>Returns a list of topics that a consumer is subscribed to get messages from. </td>
		</tr>
		<tr>
			<td>MSG.COUNTS</td>
			<td>consumerId [string]</td>
			<td>[String Array] of message counts grouped by status</td>
			<td>The returned array has strings with counts in format ["<b>status:10</b>"]</td>
		</tr>
		<tr>
			<td>MSG.DEL</td>
			<td><ol><li>consumerId [string]</li><li>status [string]</li><li>messageKey[string]</li></ol></td>
			<td></td>
			<td>Deletes a message from a consumer queue based on its status and key</td>
		</tr>
		<tr>
			<td>MSG.FAIL</td>
			<td><ol><li>consumerId [string]</li><li>messageKey[string]</li><li>error[string]</li></ol></td>
			<td></td>
			<td>Mark and active or delayed message as failed, and will move it to failed status. An error message can be sent and stored in the message for refference.</td>
		</tr>
		<tr>
			<td>MSG.KEYS</td>
			<td>consumerId [string]</td>
			<td>[String Array] [<b>Status:MessageKey]</b></td>
			<td>Returns a list of all Messages with status and keys</td>
		</tr>
		<tr>
			<td>MSG.LIST.JSON</td>
			<td><ol><li>consumerId [string]</li><li>status[string]</li></ol></td>
			<td>[JSON ] : Messages</td>
			<td>Returns a JSON will all messages and information for a status</td>
		</tr>
		<tr>
			<td>MSG.PULL</td>
			<td>consumerId [string]</td>
			<td>[JSON]: Message</td>
			<td>Pulls a new message from the queue in JSON. <i>You should then call msg.fail or msg.complete after the pull, if it takes too long to process it will move to delayed, then failed automatically</i></td>
		</tr>
		<tr>
			<td>MSG.PUSH.CONSUMER</td>
			<td><ol><li>consumerId [string]</li><li>messagePayload [string]</li></ol></td>
			<td></td>
			<td>Push a message to a single consumer by its consumerId. The payload will be in string format.</td>
		</tr>
		<tr>
			<td>MSG.PUSH.TOPIC</td>
			<td><ol><li>topic [string]</li><li>messagePayload [string]</li></ol></td>
			<td></td>
			<td>Pushes the message to all consumers that subscribe to a topic match. <i>See MQTT Topic Matching. </i></td>
		</tr>
		<tr>
			<td>MSG.RETRY</td>
			<td><ol><li>consumerId [string]</li><li>messageKey [string]</li></ol></td>
			<td></td>
			<td>Re-Queue a message that failed, placing it on top of the queue to be re-pulled.</td>
		</tr>
		<tr>
			<td>MSG.RETRYALL</td>
			<td>consumerId [string]</td>
			<td></td>
			<td>Re-Queue all the failed messages, putting them back on top of queue</td>
		</tr>
		<tr>
			<td>PING</td>
			<td>message[string]</td>
			<td>[String] "Pong! {message}" </td>
			<td></td>
		</tr>
		<tr>
			<td>Quit</td>
			<td></td>
			<td></td>
			<td>Connection to the server is gracefully dropped.</td>
		</tr>
		<tr>
			<td>INFO</td>
			<td></td>
			<td></td>
			<td>Returns server information</td>
		</tr>
		<tr>
			<td>CONSUMER.CONCURRENCYOK</td>
			<td>consumerId [string]</td>
			<td></td>
			<td>Checks if the number of active messages are still less than the concurrency threshold provided to the consumer.</td>
		</tr>
	</tbody>
</table>

### Connecting using HTTP Rest API 
Use the following routes to execute the commands, they are similar to the RESP but in HTTP format. The server utilizes the GET and POST methods as below;
The mention of variables in the table are as follows;
|Variable|Type|Usage|
|--|--|--|
consumerId|string|The consumer Id that will be used 
topics| string| A comma separated string of topics ex: path1/sub1,path2/sub2
messageId| string|The message Id 
errMsg| string| An error message that can be passed to a message on fail
topicString| string| A topic string that will be used to push messages to consumers 
messageString| string| The message body/payload to be added to message
status| string| The status of a message (pending, active, delayed, failed ,completed)
|concurrency| int| The allowed concurrency for the consumer if 0 it will be unlimited 

#### API Routes 
|Method|HTTP Route|Arguments|Returns|
|--|--|--|--|
|GET|/api/command||{JSON} List of comands
|GET|/api/consumer/concurrency/ok/:consumerId|consumerId|{JSON} Returns true or false depending on if the consumer has bandwidth to process messages
|GET|/api/consumer/exists/:consumerId|consumerId|JSON|Returns|{JSON} true or false if the consumer exists 
|GET|/api/consumer/list||{JSON} List of all consumers 
|GET|/api/consumer/topics/get/:consumerId|consumerId|{JSON} Returns the topics the consumer is listening on
|GET|/api/info||{JSON} CrimsonQ Server Information
|GET|/api/msg/counts/:consumerId|consumerId|{JSON} List Count of messages grouped by status
|GET|/api/msg/keys/:consumerId|consumerId|{JSON} List of all keys prefixed with status
|GET|/api/msg/list/json/:consumerId/:status|consumerId - status|{JSON} Gets all of the messages' body for a spesific consumerId and Status
|GET|/api/msg/pull/:consumerId|consumerId|{JSON} Pulls and returns a message with its body 
|GET|/api/ping/:messageString|messageString|{JSON} Returns Pong! + [messageString]
|POST|/api/consumer/concurrency/set|consumerId - concurrency|{JSON} ok or Error
|POST|/api/consumer/create|consumerId - topics - concurrency|{JSON} ok or Error
|POST|/api/consumer/destroy|consumerId|{JSON} ok or Error
|POST|/api/consumer/flush/complete|consumerId|{JSON} ok or Error
|POST|/api/consumer/flush/failed|consumerId|{JSON} ok or Error
|POST|/api/consumer/topics/set|consumerId - topics|{JSON} ok or Error
|POST|/api/msg/complete|consumerId - messageId|{JSON} ok or Error
|POST|/api/msg/del|consumerId - status - messageId|{JSON} ok or Error
|POST|/api/msg/fail|consumerId - messageId - errMsg|{JSON} ok or Error
|POST|/api/msg/push/consumer|consumerId - messageString|{JSON} ok or Error
|POST|/api/msg/push/topic|topicString - messageString|{JSON} ok or Error
|POST|/api/msg/retryall|consumerId|{JSON} ok or Error
|POST|/api/msg/retry|consumerId - messageId|{JSON} ok or Error|



## Performance
CrimsonQ Utilizes BadgerDB as its database layer, as well as RedCon as its RESP interface which both focus on high performance. Combine that with the concurrency of Golang. You get exceptional performance. 

The writes done on CrimsonQ using redis-bencmark to write into queues is highly performant and has very low latency. Making it ideal for high-bandwidth read/right scenarios. 

![Benchmark](https://github.com/ywadi/crimsonq/raw/main/assets/benchmark.png)

## Building a Client Library 
*TODO*
### List of Client Libraries 

|Platform|Link|Authors|
|--|--|--|
|Node.js|[crimsonqClient](https://github.com/ywadi/crimsonqClient)|Yousef Wadi, Ola Alkhateeb|

Please send up an issue to add your library here.

## CrimsonQ Admin UI 
We are actively looking for contributors to help us deliver the the Ui
