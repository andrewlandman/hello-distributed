package main

import "net"
import "fmt"
import "bufio"
import "io/ioutil"
import "net/http"
import "log"
import "time"
import "encoding/json"
import "bytes"
import "flag"

func getIP() string {
    url := "https://api.ipify.org?format=text"
    resp, err := http.Get(url)
    if err != nil {
        log.Fatalln(err)
    }
    defer resp.Body.Close()
    ip, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatalln(err)
    }
    return string(ip)
}

func getIPList(master_ip string) []string {
    resp, err := http.Get("http://" + master_ip + ":8888")
    if err != nil {
        log.Fatalln(err)
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatalln(err)
    }
    var ip_list []string
    err = json.Unmarshal(body, &ip_list)
    if err != nil {
        log.Fatalln(err)
    }
    return ip_list
}

func main() {
    optMaster := flag.String("master", "", "master server ip address")
    flag.Parse()
    master_ip := *optMaster
    if master_ip == "" {
        fmt.Println("Usage: go run node --master <master server ip address>")
        return
    }

    // get public ip address
    ip := getIP()

    // send node ip address to hardcoded master server
    data, err := json.Marshal(ip)
    if err != nil {
        log.Fatalln(err)
    }
    bytes := bytes.NewBuffer(data)
    _, err = http.Post("http://" + master_ip + ":8888/ip", "application/json", bytes)
    if err != nil {
        log.Fatalln(err)
    }

    go func() {
        for {
            // recieve node list from master server
            ip_list := getIPList(master_ip)

            // send outbound message to all nodes
            for _, ip := range ip_list {
                conn_out, err := net.DialTimeout("tcp", ip + ":7777", time.Duration(5) * time.Second)
                if err != nil {
                    fmt.Println("Failed to connect to " + ip)
                } else {
                    message := "Hi " + ip
                    fmt.Println("Sent: " + message)
                    fmt.Fprintf(conn_out, message)
                    conn_out.Close()
                }
                time.Sleep(5 * time.Second)
            }
        }
    }()

    // listen on all interfaces
    ln, _ := net.Listen("tcp", ":7777")

    for {
        // read inbound messages
        conn_in, _ := ln.Accept()
        message, _ := bufio.NewReader(conn_in).ReadString('\n')
        fmt.Println("Received: " + message)
    }
}
