package storage

import (
    "os"
    "io/ioutil"
    "log"
    "bufio"
    "net/textproto"
    "strconv"
)

var STORAGE_PATH = os.TempDir() + "/maildev"

type Mail struct{
    Id string
    Headers textproto.MIMEHeader
    Body string
}

type MailMeta struct{
    From string
    To string
    Subject string
    Date string
    Id int64
}

type Message struct{
    Title string
    Message string
}



// Create a temp directory for storing our mails
func SetUpTempDir(){
    _, err := os.Stat(STORAGE_PATH)
    if(!os.IsExist(err)){
        os.Mkdir(STORAGE_PATH, 0700)
    }
}

// List the 50 latest mails
func ListMails() ([50]MailMeta){
    files, err := ioutil.ReadDir(STORAGE_PATH)
    if(err != nil){
        log.Fatal(err)
    }

    // We will return a max number of 50 mails
    mails := [50]MailMeta{}

    i := 0
    for j := len(files) - 1; j >= 0; j-- {
        fileInfo := files[j]
        path := STORAGE_PATH + "/" + fileInfo.Name()
        file, _ := os.Open(path)

        reader := bufio.NewReader(file)
        tp := textproto.NewReader(reader)
        headers,_ := tp.ReadMIMEHeader()
        mailId, err := strconv.ParseInt(fileInfo.Name()[:13], 10, 64)

        if(err != nil){
            log.Fatal(err)
        }

        mails[i] = MailMeta{
            headers.Get("From"),
            headers.Get("To"),
            headers.Get("Subject"),
            headers.Get("Date"),
            mailId,
        }
        i+=1

        if(i == 50){
            break
        }
    }

    return mails
}

func StoreMail(mailId string, raw string){

    // Open a new file for storing our mail
    // Use the current unix nano time as unique identifier
    file, _ := os.Create(STORAGE_PATH + "/" + mailId  + ".mail");

    // Write the raw email to the file stream
    file.Write([]byte(raw))

    file.Sync()
}

func RetreiveMail(mailId string) (Mail){

    file, err := os.Open(STORAGE_PATH + "/" + mailId  + ".mail")
    if(err != nil){
        return Mail{mailId, nil, "Sorry, this mail couldn't be found"}
    }else{
        // Write the file to the browser
        reader := bufio.NewReader(file)
        tp := textproto.NewReader(reader)

        headers,_ := tp.ReadMIMEHeader()
        body,_ := tp.ReadLine()

        return Mail{mailId, headers, body}
    }
}