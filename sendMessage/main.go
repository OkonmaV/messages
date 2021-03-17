package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/big-larry/mgo"
	"github.com/big-larry/mgo/bson"
	utils "github.com/big-larry/suckutils"
	"github.com/roistat/go-clickhouse"
)

type configs struct {
	mongoSession   *mgo.Session
	mongoColl      *mgo.Collection
	clickhouseConn *clickhouse.Conn
}
type ChatInfo struct {
	Id    bson.ObjectId `bson:"_id"`
	Users []string      `bson:"users"`
	Name  string        `bson:"name"`
	Type  int           `bson:"type"`
}

func (cfg *configs) handler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(r.PostForm["text"][0]) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("koki")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	respAuth, err := http.Get(utils.ConcatFour("http://www.URL-TO-AUTH.com/?jwt=", cookie.Value, "&to=", r.PostForm["to"][0])) // TODO: что если групповой
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer func() {
		err := respAuth.Body.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}()

	if respAuth.StatusCode != 200 {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// --- HASH IN STRING IN BODY OF RESPONCE ---

	/*	bytes, err := ioutil.ReadAll(respAuth.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		ownerHash := string(bytes)
		ownerHash = ownerHash*/

	// --- HASH IN JSON IN BODY OF RESPONCE ---

	/*	respAuthBodyDecoded := &ChatCreatorAuthResp{}
		err := json.NewDecoder(respAuth.Body).Decode(respAuthBodyDecoded)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}*/

	// ---

	userHash := r.PostForm["hash"][0] //TODO: delete this

	res := &ChatInfo{}
	err = cfg.mongoColl.Find(bson.M{"Id": r.PostForm["chat"][0]}).One(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if res == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	query, err := clickhouse.BuildInsert("main.chats",
		clickhouse.Columns{"time", "chatID", "user", "text"},
		clickhouse.Row{time.Now(), r.PostForm["chat"][0], userHash, r.PostForm["text"][0]})

	err = query.Exec(cfg.clickhouseConn)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	mngSession, err := mgo.Dial("127.0.0.1")

	if err != nil {
		fmt.Println(err) //TODO
		return
	}
	defer mngSession.Close()

	mngCollection := mngSession.DB("main").C("chats")

	chConn := clickhouse.NewConn("localhost:8123", clickhouse.NewHttpTransport())
	//"CREATE TABLE IF NOT EXISTS main.chats (time DateTime,chatID UUID,user String,text String) ENGINE = MergeTree() ORDER BY tuple()"

	cfg := *&configs{mongoSession: mngSession, mongoColl: mngCollection, clickhouseConn: chConn}

	http.HandleFunc("/", cfg.handler)
	log.Fatal(http.ListenAndServe(":8092", nil))
}
