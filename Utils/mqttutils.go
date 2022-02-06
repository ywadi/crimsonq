package Utils

import (
	"strings"
)

func match(route []string, topic []string) bool {
	if len(route) == 0 {
		return len(topic) == 0
	}

	if len(topic) == 0 {
		return route[0] == "#"
	}

	if route[0] == "#" {
		return true
	}

	if (route[0] == "+") || (route[0] == topic[0]) {
		return match(route[1:], topic[1:])
	}
	return false
}

func routeIncludesTopic(route, topic string) bool {
	return match(routeSplit(route), strings.Split(topic, "/"))
}

// removes $share and sharename when splitting the route to allow
// shared subscription routes to correctly match the topic
func routeSplit(route string) []string {
	var result []string
	if strings.HasPrefix(route, "$share") {
		result = strings.Split(route, "/")[2:]
	} else {
		result = strings.Split(route, "/")
	}
	return result
}

// match takes the topic string of the published message and does a basic compare to the
// string of the current Route, if they match it returns true
func MQTTMatch(r string, topic string) bool {
	return r == topic || routeIncludesTopic(r, topic)
}
