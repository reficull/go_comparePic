package main

import (
    "net/http"
    "os"
    "strconv"
    "log"
    "./htpServer"
)
//test : ab -n 20000 -c 200 "127.0.0.1:8000/inc?name=i"
func main() {


    server := htpServer.Server{htpServer.StartProcessManager(map[string]float64{"i":0,"j":0})}
    http.HandleFunc("/uf", server.UF)

    portnum := 8882
    if len(os.Args) > 1 {
        portnum, _ = strconv.Atoi(os.Args[1])

    }
    log.Printf("compare picture service Going to listen on port %d\n", portnum)
    log.Fatal(http.ListenAndServe(":"+strconv.Itoa(portnum), nil))
}


