package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func Ping() string {
	return "+PONG\r\n"
}

func Echo(i int, commands []string) (string, int) {
	var response string
	if i < len(commands)-1 {
		response = fmt.Sprintf("$%d\r\n%s\r\n", len(commands[i+1]), commands[i+1])
		i++
	} else {
		response = "-ERR wrong number of arguments for 'echo' command\r\n"
	}
	return response, i
}

func Set(i int, commands []string) (string, int) {
	var response string
	if i < len(commands)-2 {
		key := commands[i+1]
		value := commands[i+2]
		KeyValuePairs[key] = value
		fmt.Printf("SET %s %s\n", key, value)

		if i+3 < len(commands) && strings.ToUpper(commands[i+3]) == "PX" {
			expiry, err := strconv.Atoi(commands[i+4])
			if err != nil {
				response = "-ERR invalid expiry\r\n"
				return response, i
			}

			KeyExpiryTime[key] = time.Now().UnixNano() + int64(expiry)*int64(time.Millisecond)

			/* Active Expiry: Working Idea
			go func(k string, exp int) {
				time.Sleep(time.Duration(exp) * time.Millisecond)
				delete(KeyValuePairs, k)
				fmt.Printf("Key %s expired\n", k)
			}(key, expiry)
			*/
			i += 4
		} else {
			i += 2
		}
		response = "+OK\r\n"
	} else {
		response = "-ERR wrong number of arguments for 'set' command\r\n"
	}
	return response, i
}

func Get(i int, commands []string) (string, int) {
	var response string
	if i < len(commands)-1 {
		key := commands[i+1]
		fmt.Printf("GET %s\n", key)
		if value, ok := KeyValuePairs[key]; ok {
			if expiry, found := KeyExpiryTime[key]; found && expiry <= time.Now().UnixNano() {
				// Key has expired, delete it
				delete(KeyValuePairs, key)
				delete(KeyExpiryTime, key)
				response = EmptyResponse()
			} else {
				response = fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
			}
			i++
		} else {
			response = EmptyResponse()
			i++
		}
	} else {
		response = "-ERR wrong number of arguments for 'get' command\r\n"
	}
	return response, i
}

func Info(i int, commands []string) (string, int) {
	var response string
	if i < len(commands)-1 && strings.ToUpper(commands[i+1]) == "REPLICATION" {
		if isReplica {
			response = "$10\r\nrole:slave\r\n"
		} else {
			raw := fmt.Sprintf("role:master\nmaster_replid:%s\nmaster_repl_offset:%d", masterReplID, masterReplOffset)
			response = fmt.Sprintf("$%d\r\n%s\r\n", len(raw), raw)
		}
		i++
	} else {
		response = "$13\r\n# Replication\r\n"
	}
	return response, i
}

// HandleREPLCONF handles the REPLCONF command for replica configuration
func HandleREPLCONF(index int, commands []string) (string, int) {
	var response string
	if index < len(commands)-1 {
		subCommand := strings.ToUpper(commands[index+1])
		switch subCommand {
		case "LISTENING-PORT":
			if index < len(commands)-2 {
				_, err := strconv.Atoi(commands[index+2])
				if err != nil {
					response = "-ERR invalid listening port\r\n"
				} else {
					response = "+OK\r\n"
					index += 2
				}
			} else {
				response = "-ERR wrong number of arguments for 'REPLCONF' command\r\n"
			}
		case "CAPA":
			if index < len(commands)-2 {
				capability := strings.ToUpper(commands[index+2])
				if capability == "PSYNC2" {
					response = "+OK\r\n"
				} else {
					response = "-ERR unsupported capability\r\n"
				}
				index += 2
			} else {
				response = "-ERR wrong number of arguments for 'REPLCONF' command\r\n"
			}
		default:
			response = "-ERR unknown subcommand for 'REPLCONF' command\r\n"
		}
	} else {
		response = "-ERR wrong number of arguments for 'REPLCONF' command\r\n"
	}
	return response, index
}
