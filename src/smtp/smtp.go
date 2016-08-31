package smtp

import (
    "net"
    "log"
    "bufio"
    "strings"
    "io"
    "storage"
    "strconv"
    "time"
)

const MAX_SIZE = 102400

type SmtpServer struct {
    listener net.Listener
    MailChan chan storage.Mail
}

func NewServer(listenInterface string) *SmtpServer{
    listen, err := net.Listen("tcp", listenInterface)

    if (err != nil) {
        log.Fatal(err)
    }
    server := new(SmtpServer)
    server.listener = listen

    return server
}

// Will accept client connections and will delicate it to a go routine
func (s *SmtpServer) Start(){
    for{
        conn, err := s.listener.Accept()
        if(err != nil){
            log.Println(err)
            continue
        }
        go s.handleClient(conn)
    }
}

// Handles all the communication with the client
// This includes the handling of the receiving mail
func (s *SmtpServer) handleClient(client net.Conn){
    // When we are done, close the connection
    defer client.Close()

    // Create or reader and writer streams
    reader := bufio.NewReader(client)
    writer := bufio.NewWriter(client)

    data := ""

    _, err := sendResponse(writer, "220 maildev smtp server\r\n")
    if(err != nil){
        log.Println(err)
    }

    for{
        cmd, err := receiveMessage(reader)

        // Apparently we are done
        if(err == io.EOF){
            break;
        }

        if(err != nil){
            log.Println(err)
            break
        }

        // TODO: It's probably a good idea to keep track of the communication state
        switch {

        // Handling HELO
        case strings.Index(cmd, "HELO") == 0:
            sendResponse(writer, "250 Maildev")
            break

        // Handling EHLO
        case strings.Index(cmd, "EHLO") == 0:
            sendResponse(writer, "250-maildev")
            sendResponse(writer, "250-PIPELINING")
            sendResponse(writer, "250-SIZE " + strconv.FormatInt(MAX_SIZE, 10))
            sendResponse(writer, "250 HELP")
            break

        // Handling mail from
        case strings.Index(cmd, "MAIL FROM:") == 0:
            // TODO: Use this "from" instead of the header version?
            sendResponse(writer, "250 Ok")
            break

        // Handling rcpt to
        case strings.Index(cmd, "RCPT TO:") == 0:
            // TODO: Use this "to" instead of the header version?
            sendResponse(writer, "250 Ok")
            break

        // Handling mail data
        case strings.Index(cmd, "DATA") == 0:
            sendResponse(writer, "354 End data with <CR><LF>.<CR><LF>")

            data, err = readMail(reader)
            if(err != nil && err != io.EOF){
                log.Println("Failed reading mail: " + err.Error())
            }
            mailId := strconv.FormatInt( time.Now().UnixNano() / int64(time.Millisecond), 10)
            storage.StoreMail(mailId, data)

            mail := new(storage.Mail)
            mail.Id = mailId
            mail.Body = data

            s.MailChan <- *mail

            sendResponse(writer, "250 Ok")
            break

        case strings.Index(cmd, "QUIT") == 0:
            sendResponse(writer, "221 Bye")
            break

        default:
            sendResponse(writer, "500 Unrecognized command")
            println("Did not understood command: " + cmd)
            break
        }

    }
}

func sendResponse(writer *bufio.Writer, response string) (num int, err error){
    num, err = writer.WriteString(response + "\r\n")

    if(err != nil){
        return num, err
    }

    err = writer.Flush()

    return num, err
}

func receiveMessage(reader *bufio.Reader) (message string, err error){
    message, err = reader.ReadString('\n')
    if(err != nil){
        return "", err
    }

    message = strings.TrimSpace(message)

    return message, err
}

// Reads the whole mail message including the headers
// TODO: complex messages will be a problem, like messages which include attachments
func readMail(reader * bufio.Reader) (data string, err error){
    for{
        part, err := reader.ReadString('\n')

        part = strings.TrimSpace(part) + "\n"

        if(err != nil || part == ".\n"){
            return data, err
        }

        data += part
    }

    return data, err
}