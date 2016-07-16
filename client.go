package main

import (
	"bufio"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strings"
	"time"
)

type Nothing struct{}

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		log.Fatalf("Usage: %s <Username> OR %s <username> <serveraddress>", os.Args[0])
	}

	username := os.Args[1]
	address := ":3410"
	if len(os.Args) == 3 {
		address = os.Args[2]
	}
	if strings.HasPrefix(address, ":") {
		address = "localhost" + address
	}

	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Fatalf("Error connecting to server at %s: %v", address, err)
	}
	junk := new(Nothing)
	err = client.Call("Server.Register", &username, &junk)
	if err != nil {
		log.Fatalf("Couldn't register your username: %v", err)
	}

	go func() {
		for {
			messages := []string{}
			err = client.Call("Server.CheckMessages", &username, &messages)
			if err != nil {
				log.Fatalf("Couldn't check your messages: %v", err)
			}
			for _, message := range messages {
				fmt.Println(message)
			}

			time.Sleep(time.Second)
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter a Command:")
		text, _ := reader.ReadString('\n')
		words := strings.Fields(text)
		if len(words) == 0 {
			fmt.Println("")
		} else if words[0] == "tell" {
			to := words[1]
			message := ""
			for i := 2; i < len(words); i++ {
				message += words[i]
				message += " "
			}
			request := []string{username, to, message}
			err = client.Call("Server.Tell", request, &junk)
			if err != nil {
				log.Fatalf("Couldn't tell because: %v", err)
			}
		} else if words[0] == "say" {
            message := ""
            for i := 1; i < len(words); i++ {
				message += words[i]
				message += " "
			}
            request := []string{username, message}
			err = client.Call("Server.Say", request, &junk)
            if err != nil {
				log.Fatalf("Couldn't say because: %v", err)
			}
		} else if words[0] == "list" {
            messages := []string{}
            err = client.Call("Server.List", &junk, &messages)
            if err != nil {
				log.Fatalf("Couldn't list because: %v", err)
			}
            for _, message := range messages {
				fmt.Print(message)
			}

		} else if words[0] == "quit" {
            err = client.Call("Server.Logout", username, &junk)
            if err != nil {
				log.Fatalf("Couldn't quit because: %v", err)
			}
            return
		} else if words[0] == "help" {
			fmt.Println("All the possible commands are:")
			fmt.Println("tell <user> some message: This sends “some message” to a specific user.")
			fmt.Println("say some other message: This sends “some other message” to all users.")
			fmt.Println("list: This lists all users currently logged in.")
			fmt.Println("quit: this logs you out.")
			fmt.Println("shutdown: this shuts down the server.")
		} else if words[0] == "shutdown" {
            err = client.Call("Server.Shutdown", &junk, &junk)
            if err != nil {
				log.Fatalf("Couldn't shutdown because: %v", err)
			}
            return
		} else {
			fmt.Println("Not a valid command, try 'help'")
		}
	}

}
