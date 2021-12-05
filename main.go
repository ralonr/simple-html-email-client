package main

import (
	"bytes"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/jordan-wright/email"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"os"
)

var tpl *template.Template

type EmailType struct {
	Email   string
	Pass    string
	SMTP    string
	Host    string
	Subject string
}

type EmailMessage struct {
	First string
	Email []string
	Cc    []string
	Bcc   []string
	Pass  string
}

var eConfig EmailType

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
	err := godotenv.Load("env.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	eConfig.Email = os.Getenv("EMAIL")
	eConfig.Pass = os.Getenv("PASS")
	eConfig.SMTP = os.Getenv("SMTP")
	eConfig.Host = os.Getenv("HOST")
	eConfig.Subject = os.Getenv("SUBJECT")
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/send", send)
	fmt.Println("listening on port:8080")
	_ = http.ListenAndServe(":8080", nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	_ = tpl.ExecuteTemplate(w, "index.gohtml", nil)
}

func send(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	toEmail := r.FormValue("email")
	fname := r.FormValue("firstname")
	cc := r.FormValue("Cc")
	bcc := r.FormValue("Bcc")
	password := r.FormValue("pass")
	to := []string{toEmail}
	c := []string{cc}
	bc := []string{bcc}

	data := EmailMessage{
		First: fname,
		Email: to,
		Cc:    c,
		Bcc:   bc,
		Pass:  password,
	}

	t, err := template.ParseFiles("./templates/LMSCredentials.gohtml")
	if err != nil {
		log.Fatalln(err)
	}

	var tplBuf bytes.Buffer
	if err := t.Execute(&tplBuf, data); err != nil {
		log.Fatalln(err)
	}

	templateBytes := tplBuf.Bytes()
	err = SendNewMail(eConfig, data, &templateBytes)
	if err != nil {
		log.Fatalln(err)
		return
	}

	_ = tpl.ExecuteTemplate(w, "send.gohtml", data)
	return
}

// SendNewMail sends email to the user
func SendNewMail(emConfig EmailType, dataEmail EmailMessage, htmlIncoming *[]byte) (err error) {
	//err = e.SendWithStartTLS("smtp.office365.com:587", smtp.PlainAuth(auth.Identity, auth.UserName, auth.Password, auth.Host),)

	e := email.NewEmail()
	e.Subject = emConfig.Subject
	e.HTML = *htmlIncoming
	e.From = emConfig.Email
	e.To = dataEmail.Email
	if len(dataEmail.Cc) > 0 {
		e.Cc = dataEmail.Cc
	}

	if len(dataEmail.Bcc) > 0 {
		e.Bcc = dataEmail.Bcc
	}
	//e.Send("smtp.gmail.com:587", smtp.PlainAuth("", "", "", "smtp.gmail.com"))
	err = e.Send(emConfig.SMTP, smtp.PlainAuth("", emConfig.Email, emConfig.Pass, emConfig.Host))
	if err != nil {
		return err
	}
	return nil
}
