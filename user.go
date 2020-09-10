package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	Name  string `json:"name"`
	Token string
}

// bodyからjson形式のnameデータを取得し、user型に格納して返す
func getUserData(w http.ResponseWriter, r *http.Request) User {

	length, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	body := make([]byte, length)
	length, err = r.Body.Read(body)
	if err != nil && err != io.EOF {
		w.WriteHeader(http.StatusInternalServerError)
	}

	var user User
	err = json.Unmarshal(body[:length], &user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	return user
}

// DBにuserの名前、トークンを格納
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

// ユーザーを作成し、tokenを付与
func createUser(w http.ResponseWriter, r *http.Request) {
	user := getUserData(w, r)

	// jwt版
	// secret := "secret"

	// token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
	// 	"name": user.Name,
	// 	"iss":  "__init__",
	// })

	// tokenString, _ := token.SignedString([]byte(secret))
	// user.Token = tokenString
	//

	// ランダム文字列版
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, 20)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	user.Token = string(b)
	//

	insertData(user)
	fmt.Print(user)
}

// ヘッダーからx-tokenを受け取り、DB照会して合致したデータの名前を返す
func getUser(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("x-token")
	db, err := sql.Open("mysql", "root:@/techtrain-mission-gameapi")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	var result string
	if err := db.QueryRow("SELECT name FROM user WHERE token=?;", token).Scan(&result); err != nil {
		log.Fatal(err)
	}
	fmt.Println("name:", result)
}

// ヘッダーからx-token、bodyから更新後のnameを取得
// x-tokenに合致したユーザーデータのnameを更新
func updateUser(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("x-token")
	db, err := sql.Open("mysql", "root:@/techtrain-mission-gameapi")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	var result string
	if err := db.QueryRow("SELECT name FROM user WHERE token=?;", token).Scan(&result); err != nil && err != sql.ErrNoRows {
		log.Fatal(err)
	}

	user := getUserData(w, r)
	user.Token = token

	// tokenがx-tokenに合致するデータの名前を更新
	upd, err := db.Prepare("UPDATE user set name=? WHERE token=?;")
	if err != nil {
		log.Fatal(err)
	}

	upd.Exec(user.Name, user.Token)

	fmt.Println("updated", user.Name, user.Token)
}
