package main

import (
	"fmt"
	"net/http"
	"net/rpc"
	"os"
	"strings"
    "time"
    "sync"
)

type Nothing struct{}

type Server struct {
	users    map[string][]string
	shutdown bool
    mutex sync.Mutex
}

func (s *Server) Register(username *string, junk *Nothing) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    for user := range s.users {
        if user == *username {
            return fmt.Errorf("the username %s is already taken", user)
        }
    }
	s.users[*username] = make([]string, 0)
	for user := range s.users {
		s.users[user] = append(s.users[user], "\n"+*username+" has joined the chat")
	}
	return nil
}

func (s *Server) List(junk *Nothing, response *[]string) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
	for user := range s.users {
		*response = append(*response, user+"\n")
	}
	return nil
}

func (s *Server) CheckMessages(username *string, messages *[]string) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
	// for i := 0; i < len(s.users[*username]); i++ {
	// 	*messages = append(*messages, s.users[*username][i])
	// }

    *messages = s.users[*username]
	s.users[*username] = []string{}
	return nil
}

func (s *Server) Tell(request *[]string, junk *Nothing) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
	stuff := *request
	s.users[stuff[1]] = append(s.users[stuff[1]], "\n"+stuff[0]+" tells you: "+stuff[2])
	return nil
}

func (s *Server) Say(request *[]string, junk *Nothing) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
	stuff := *request
	for user := range s.users {
		s.users[user] = append(s.users[user], "\n"+stuff[0]+" says: "+stuff[1])
	}
	return nil
}

func (s *Server) Logout(username *string, junk *Nothing) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
	delete(s.users, *username)
	for user := range s.users {
		s.users[user] = append(s.users[user], "\n"+*username+" has logged out")
	}

	return nil
}

func (s *Server) Shutdown(junk *Nothing, nothing *Nothing) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
	s.shutdown = true
	return nil
}

func main() {
	port := "3410"
	if len(os.Args) == 2 {
		port = os.Args[1]
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	fmt.Printf("The port is %s\n", port)

	server := Server{users: make(map[string][]string)}
    go func() {
        for {
            if server.shutdown {
                os.Exit(1)
            }
            time.Sleep(time.Second)
        }
    }()
	rpc.Register(&server)
	rpc.HandleHTTP()

	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println(err.Error())
	}
}
