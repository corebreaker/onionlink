package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/terminal"
)

func main() {
	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	config := &ssh.ServerConfig{
		PasswordCallback: func(conn *ssh.ServerConn, user, pass string) bool {
			return user == "testuser" && pass == "tiger"
		},
	}

	pemBytes, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		log.Fatal("Failed to load private key:", err)
	}
	if err = config.SetRSAPrivateKey(pemBytes); err != nil {
		log.Fatal("Failed to parse private key:", err)
	}

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	conn, err := ssh.Listen("tcp", "0.0.0.0:2022", config)
	if err != nil {
		log.Fatal("failed to listen for connection")
	}
	for {
		// A ServerConn multiplexes several channels, which must
		// themselves be Accepted.
		log.Println("accept")
		sConn, err := conn.Accept()
		if err != nil {
			log.Println("failed to accept incoming connection")
			continue
		}
		if err := sConn.Handshake(); err != nil {
			log.Println("failed to handshake")
			continue
		}
		go handleServerConn(sConn)
	}
}

func handleServerConn(sConn *ssh.ServerConn) {
	defer sConn.Close()
	for {
		// Accept reads from the connection, demultiplexes packets
		// to their corresponding channels and returns when a new
		// channel request is seen. Some goroutine must always be
		// calling Accept; otherwise no messages will be forwarded
		// to the channels.
		ch, err := sConn.Accept()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Println("handleServerConn Accept:", err)
			break
		}
		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell, the type is
		// "session" and ServerShell may be used to present a simple
		// terminal interface.
		if ch.ChannelType() != "session" {
			ch.Reject(ssh.UnknownChannelType, "unknown channel type")
			break
		}
		go handleChannel(ch)
	}
}

func handleChannel(ch ssh.Channel) {
	term := terminal.NewTerminal(ch, "> ")
	serverTerm := &ssh.ServerTerminal{
		Term:    term,
		Channel: ch,
	}
	ch.Accept()
	defer ch.Close()
	for {
		line, err := serverTerm.ReadLine()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Println("handleChannel readLine err:", err)
			continue
		}
		fmt.Println(line)
	}
}
