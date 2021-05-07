package main

import (
	"lib"
	"net/url"
	"thin-peak/logs/logger"
	"time"

	"github.com/big-larry/mgo"
	"github.com/big-larry/mgo/bson"
	"github.com/big-larry/suckhttp"
	"github.com/roistat/go-clickhouse"
)

type SendMessage struct {
	mgoSession     *mgo.Session
	mgoColl        *mgo.Collection
	clickhouseConn *clickhouse.Conn
}
type chatInfo struct {
	Id    bson.ObjectId `bson:"_id"`
	Users []string      `bson:"users"`
	Name  string        `bson:"name"`
	Type  int           `bson:"type"`
}

func NewSendMessage(mgoAddr string, mgoColl string, clickhouseAddr string) (*SendMessage, error) {

	mgoSession, err := mgo.Dial(mgoAddr)
	if err != nil {
		logger.Error("Mongo conn", err)
		return nil, err
	}

	mgoCollection := mgoSession.DB("main").C(mgoColl)

	chConn := clickhouse.NewConn(clickhouseAddr, clickhouse.NewHttpTransport())
	//"CREATE TABLE IF NOT EXISTS main.chats (time DateTime,chatID UUID,user String,text String) ENGINE = MergeTree() ORDER BY tuple()"
	err = chConn.Ping()
	if err != nil {
		logger.Error("Clickhouse connection ping", err)
		return nil, err
	}
	return &SendMessage{mgoSession: mgoSession, mgoColl: mgoCollection, clickhouseConn: chConn}, nil

}

func (conf *SendMessage) Close() error {
	conf.mgoSession.Close()
	return nil
}

func (conf *SendMessage) Handle(r *suckhttp.Request, l *logger.Logger) (w *suckhttp.Response, err error) {

	w = &suckhttp.Response{}

	formValues, err := url.ParseQuery(string(r.Body))
	if err != nil {
		w.SetStatusCode(400, "Bad Request")
		return
	}

	text := formValues.Get("text")

	if text == "" {
		w.SetStatusCode(400, "Bad Request")
		return
	}

	cookie := lib.GetCookie(r.GetHeader(suckhttp.Cookie), "koki")
	if cookie == nil {
		w.SetStatusCode(401, "Unauthorized")
		return
	}
	// TODO: AUTH

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

	userHash := formValues.Get("from") //TODO: delete this
	if userHash == "" {
		w.SetStatusCode(400, "Bad Request")
		return
	}
	chat := formValues.Get("chat")
	if chat == "" {
		w.SetStatusCode(400, "Bad Request")
		return
	}

	res := &chatInfo{}
	err = conf.mgoColl.Find(bson.M{"Id": chat}).One(res) //
	if err != nil {
		if err == mgo.ErrNotFound {
			w.SetStatusCode(400, "Bad Request")
			return
		}
		return nil, err
	}

	query, err := clickhouse.BuildInsert("main.chats",
		clickhouse.Columns{"time", "chatID", "user", "text"},
		clickhouse.Row{time.Now(), chat, userHash, text})
	if err != nil {
		return nil, err
	}
	err = query.Exec(conf.clickhouseConn)
	if err != nil {
		return nil, err
	}
	w.SetStatusCode(200, "OK")
	return
}
