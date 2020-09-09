package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	"github.com/dgrijalva/jwt-go"
)

type User struct {
	Name  string `json:"name"`
	Token string
}

func getUserData(w http.ResponseWriter, r *http.Request) User {
	//To allocate slice for request body
	length, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// return
	}

	//Read body data to parse json
	body := make([]byte, length)
	length, err = r.Body.Read(body)
	if err != nil && err != io.EOF {
		w.WriteHeader(http.StatusInternalServerError)
		// return
	}

	//parse json
	var user User
	err = json.Unmarshal(body[:length], &user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// return
	}
	return user
}

func insertData(user User) {
	db, err := sql.Open("mysql", "root:@/techtrain-mission-gameapi")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	ins, err := db.Prepare("INSERT INTO user(name, token) VALUES(?,?)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = ins.Exec(user.Name, user.Token)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("inserted", user.Name, user.Token)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	user := getUserData(w, r)

	// 鍵となる文字列(多分なんでもいい)
	secret := "secret"

	// Token を作成
	// jwt -> JSON Web Token - JSON をセキュアにやり取りするための仕様
	// jwtの構造 -> {Base64 encoded Header}.{Base64 encoded Payload}.{Signature}
	// HS254 -> 証明生成用(https://ja.wikipedia.org/wiki/JSON_Web_Token)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": user.Name,
		"iss":  "__init__", // JWT の発行者が入る(文字列(__init__)は任意)
	})

	//Dumpを吐く
	// spew.Dump(token)

	tokenString, _ := token.SignedString([]byte(secret))
	user.Token = tokenString

	insertData(user)
	fmt.Print(user)

}
