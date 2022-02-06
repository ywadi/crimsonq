
# CrimsonQ
[![buddy pipeline](https://app.buddy.works/ywadi85/crimsonq/pipelines/pipeline/373325/badge.svg?token=463c4f343893f85c5056a16ba6da1379079553b6b7a950b7ba9d643591fcb0d2 "buddy pipeline")](https://app.buddy.works/ywadi85/crimsonq/pipelines/pipeline/373325)
## A Multi Consumer per Message Queue with persistence and Queue Stages.
![crimsonq](https://github.com/ywadi/crimsonq/raw/main/assets/logo.png)

[![Build Status](https://travis-ci.org/joemccann/dillinger.svg?branch=master)](https://travis-ci.org/joemccann/dillinger)

 __Under Active Development__
 
Crimson Queue allows you to have multiple consumers listening on topics and receive a copy of a message that targets its topic. It provides multiple consumers with a message queue and complex message routing. Utilizes the same methodology as MQTT topic routing but with queue stages and persistence per consumer. This is under active development. It is being developed in Go and utilizes BadgerDB as the internal database layer to provide top notch performance, persistence and durability. It supports Redis RESP protocol, allowing any redis library the ability to connect to it and communicate with it also keeping the performance at, also can be utilized from redis-cli. 

The combination of BadgerDB and RESP protocol allows near 0.5M transactions a second on a single core with minimal ram requirements. It is optimized to be used in the cloud with VPS servers providing SSD storages. 
