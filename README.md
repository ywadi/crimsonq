

# CrimsonQ 
![GitHub release (latest by date)](https://img.shields.io/github/v/release/ywadi/crimsonq)
![GitHub](https://img.shields.io/github/license/ywadi/crimsonq)
![GitHub](https://img.shields.io/badge/Built%20On-Golang-lightgrey?logo=go)
[![Sockerhub](https://img.shields.io/badge/DockerHub-crimsonq-blue?logo=docker)](https://hub.docker.com/r/yousefjwadi/crimsonq)


## A Multi Consumer per Message with persistence and Queue Stages.
![crimsonq](https://github.com/ywadi/crimsonq/raw/main/assets/logo.png)

**[Website](http://www.crimsonq.io) ♦️ [Documentation](https://docs.crimsonq.io/) ♦️ [Docker-Hub](https://hub.docker.com/r/yousefjwadi/crimsonq)** 

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

Full documentation can be found at [docs.crimsonq.io](https://docs.crimsonq.io/)
