package main

import (
    "net/http"
    "github.com/gorilla/websocket"
    "os"
    "bufio"
    "net/textproto"
    "encoding/json"
    "storage"
    "log"
)

func defineHandlers(mux *http.ServeMux){
    // Serves info about a e-mail in JSON format
    mux.HandleFunc("/api/mail/", handler)

    // Serves a e-mail message that is viewed in a iframe
    mux.HandleFunc("/api/mail/raw/", mailRawHandler)

    // Handles our websocket
    mux.HandleFunc("/websocket", websocketHandler)

    // Serve our static files
    mux.Handle("/", http.FileServer(http.Dir("public")))
}

func handler(writer http.ResponseWriter, request *http.Request) {
    writer.Header().Set("Content-type", "application/json")
    mailId := request.URL.Path[10:]

    // We will list all the mails (from, to, subject)
    if(mailId == ""){

        mails := storage.ListMails()

        json, _ := json.Marshal(mails)
        writer.Write(json)

        return

        // We got a mailId so lets try to find it and of course return ing
    }else{
        mail := storage.RetreiveMail(mailId)
        jsonMail, _ := json.Marshal(mail)
        writer.Write(jsonMail)
    }

}

var upgrader = websocket.Upgrader{}
var connections = make(map[*websocket.Conn]bool)

func websocketHandler(writer http.ResponseWriter, request *http.Request){
    connection, err := upgrader.Upgrade(writer, request, nil)
    connections[connection] = true
    if err != nil {
        log.Println("upgrade:", err)
        return
    }
}

func mailRawHandler(writer http.ResponseWriter, request *http.Request){
    mailId := request.URL.Path[14:]
    file, err := os.Open("/tmp/maildev/" + mailId  + ".mail")
    if(err != nil){
        writer.WriteHeader(http.StatusNotFound)
        writer.Write([]byte("Sorry, could not found this e-mail"))
    }else{
        reader := bufio.NewReader(file)
        tp := textproto.NewReader(reader)
        headers,_ := tp.ReadMIMEHeader()

        contentType := headers.Get("Content-Type")
        if(contentType != ""){
            writer.Header().Set("Content-type", contentType)
        }

        input := bufio.NewScanner(reader)

        for input.Scan() {
            line := input.Text()
            data := []byte(line + "\n")
            writer.Write(data)

        }
    }
}

// Sends the message to all active websocket clients
func BroadCast(message string){
    for conn := range connections{
        err := conn.WriteMessage(websocket.TextMessage, []byte(message))
        if(err != nil){
            log.Println("Could not write message: " + err.Error())
            delete(connections, conn)
        }
    }
}