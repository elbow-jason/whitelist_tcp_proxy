package main

import (
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

const fileSep = "\n"
const filename = "whitelist.txt"

// Whitelist .
type Whitelist struct {
	addrs []string
	mutex sync.Mutex
}

func newWhitelist() Whitelist {
	return Whitelist{
		addrs: []string{},
		mutex: sync.Mutex{},
	}
}

func (w *Whitelist) save() {
	data := []byte(strings.Join(w.addrs, fileSep))
	w.mutex.Lock()
	defer w.mutex.Unlock()
	err := ioutil.WriteFile(filename, data, 0777)
	if err != nil {
		panic(err)
	}
}

func (w *Whitelist) load() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.addrs = readWhitelistFromFile()
}

func (w *Whitelist) ipAddresses() []string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return append([]string{}, w.addrs...)
}

func (w *Whitelist) isWhitelisted(ip string) bool {
	ipAddrs := w.ipAddresses()
	for _, whitelisted := range ipAddrs {
		if ip == whitelisted {
			return true
		}
	}
	return false
}

var whitelist = newWhitelist()

func readWhitelistFromFile() []string {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return strings.Split(string(dat), "\n")
}

func forward(conn net.Conn) {
	client, err := net.Dial("tcp", os.Args[2])
	if err != nil {
		log.Fatalf("Dial failed: %v", err)
	}
	log.Printf("Connected %v to server at %v", conn.RemoteAddr(), client.RemoteAddr())
	go func() {
		defer client.Close()
		defer conn.Close()
		io.Copy(client, conn)
	}()
	go func() {
		defer client.Close()
		defer conn.Close()
		io.Copy(conn, client)
	}()
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage %s listen:port forward:port\n", os.Args[0])
		return
	}

	whitelist.load()

	listener, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		log.Fatalf("Failed to setup listener: %v", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("ERROR: failed to accept listener: %v", err)
		}
		ipString := parseIP(conn)
		if whitelist.isWhitelisted(ipString) {
			log.Printf("Accepted connection %v\n", conn.RemoteAddr())
			go forward(conn)
		} else {
			log.Printf("Rejected connection %v\n", conn.RemoteAddr())
			conn.Close()
		}
	}
}

func parseIP(conn net.Conn) string {
	return strings.Split(conn.RemoteAddr().String(), ":")[0]
}
