package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	//"io"
	"log"
	"net/http"

	"github.com/skip2/go-qrcode"
	//"os"
)

var scoreboard = make(map[string]int)

func gencode(domain_name string, qrdata string) {
	_ = qrcode.WriteFile(domain_name+"/"+qrdata, qrcode.Medium, 256, "qr/"+qrdata+".png")
}

func main() {
	fmt.Println("[server.go] Initializing hash computations")
	qrdatas := make([]string, 0)                     // stores the hashes / PNG names, init at 0
	h := sha256.New()                                // create a SHA256 hash computer
	iv, _ := rand.Int(rand.Reader, big.NewInt(1028)) // securely randomized init vector : prevents cheating based on "qr0, qr1, qr2"
	h.Write(iv.Bytes())                              // init the hashes with the IV
	fmt.Println("[server.go] Generating code hashes")
	for i := 0; i < 10; i++ { // here we have 10 QR but that's highly customizable
		digits := fmt.Sprintf("%d", i)
		qrrawhash := h.Sum([]byte(digits))
		h.Write(qrrawhash)
		qrhexhash := hex.EncodeToString(qrrawhash)
		qrdatas = append(qrdatas, string(qrhexhash))
	}
	fmt.Println("[server.go] Generating images and handlers")
	for _, elem := range qrdatas {
		fmt.Println(elem)
		gencode("http://localhost:6969", elem)
		http.HandleFunc("/"+elem, RESTHandler)
	}
	http.HandleFunc("/registerform", RESTregisterHandler)
	http.HandleFunc("/registersuccess", RESTnameHandler)
	fmt.Println("[server.go] Starting REST server")
	err := http.ListenAndServe(":6969", nil)
	log.Fatal("[server.go] ListenAndServe: ", err)
}

func authFromToken(token *http.Cookie, err error) string {
	if err == http.ErrNoCookie {
		return "Unregistered user"
	}
	if err != nil {
		return "UNKNOWN_SERVER_ERROR"
	}
	return token.Value // TODO
}

func RESTregisterHandler(w http.ResponseWriter, r *http.Request) {
	// Will require a full discord connexion later on
	const data = `<!DOCTYPE html>
<html>
<head></head>
<body>
<form action="/registersuccess" method="post">
Nom: <input type="text" name="name"/> <input type="submit"/></form>
</body>
</html>`
	if r.Method != "HEAD" && r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, data)
}

func RESTnameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	name := r.Form.Get("name")
	cookie := http.Cookie{
		Name:  "PSU_treasure_quest_token",
		Value: name,
	}
	http.SetCookie(w, &cookie)
	fmt.Fprintf(w, "Compte bien enregistré")
}

func RESTHandler(w http.ResponseWriter, r *http.Request) {
	QRhash := r.URL.EscapedPath() // has a trailing slash on front
	QRhash = QRhash[1:]           // removes trailing slash
	token, err := r.Cookie("PSU_treasure_quest_token")
	name := authFromToken(token, err)
	const data = `<!DOCTYPE html>
<html>
<head></head>
<body>
<p>Bravo ! Tu as trouvé un QRCode PSU !</p>
<p>ID du QRCode : %s</p>
<p>Utilisateur : %s</p>
</body>
</html>`
	if r.Method != "HEAD" && r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	filled_data := fmt.Sprintf(data, QRhash, name)
	scoreboard[name] = scoreboard[name] + 1
	fmt.Fprintf(w, filled_data)
}
