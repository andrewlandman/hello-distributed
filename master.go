package main

import "fmt"
import "net/http"
import "log"
import "encoding/json"
import "sync"

var mu = &sync.RWMutex{}
var set = make(map[string]bool)

func postIP(w http.ResponseWriter, r *http.Request){
    log.Println("Hit: postIP")
    decoder := json.NewDecoder(r.Body)
    var ip string
    err := decoder.Decode(&ip)
    if err != nil {
        log.Fatalln(err)
    }
    fmt.Println(ip)

    mu.Lock()
    set[ip] = true
    mu.Unlock()
}

func getIPs(w http.ResponseWriter, r *http.Request){
    log.Println("Hit: getIPs")

    list := []string{}
    mu.RLock()
    for k := range set {
        list = append(list, k)
    }
    mu.RUnlock()

    json.NewEncoder(w).Encode(list)
}

func main() {
    http.HandleFunc("/", getIPs)
    http.HandleFunc("/ip", postIP)

    log.Fatalln(http.ListenAndServe(":8888", nil))
}
