package main

import (
  "fmt"
  "net/http"
  "smtp"
    "storage"
    "log"
)

func fireUpHttpServer(){
    web := http.NewServeMux()

    defineHandlers(web)
    // TODO: port number should be configurable
    err := http.ListenAndServe(":8025", web)
    if(err != nil){
        log.Fatal(err)
    }
}

func fireUpSmtpServer(mailChan chan storage.Mail){
    // TODO: port number should be configurable
    mail := smtp.NewServer(":2525")
    mail.MailChan = mailChan
    mail.Start()
}

func broadcastSender(){
    for{
        mail:= <- mailChan
        // Inform the users who are connected to the web interface that we received a new e-mail
        BroadCast(mail.Id)
    }
}

// Channel for all the received mail
var mailChan = make(chan storage.Mail)

func main() {
    fmt.Println("Starting the Mail Dev daemon")

    // Make sure we have a temp directory to read the mails from
    storage.SetUpTempDir()

    go fireUpSmtpServer(mailChan)
    go broadcastSender()

    // The HTTP server that serves our api and UI
    fireUpHttpServer()
}
