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
	mgoSession      *mgo.Session
	mgoColl         *mgo.Collection
	clickhouseConn  *clickhouse.Conn
	clickhouseTable string
}
type chatInfo struct {
	Id    string   `bson:"_id"`
	Users []string `bson:"users"`
	Name  string   `bson:"name"`
	Type  int      `bson:"type"`
}

func NewSendMessage(mgoAddr string, mgoColl string, clickhouseAddr string, clickhouseTable string) (*SendMessage, error) {

	mgoSession, err := mgo.Dial(mgoAddr)
	if err != nil {
		logger.Error("Mongo conn", err)
		return nil, err
	}
	logger.Info("DB", "Mongo connected!")
	mgoCollection := mgoSession.DB("main").C(mgoColl)

	chConn := clickhouse.NewConn(clickhouseAddr, clickhouse.NewHttpTransport())
	//"CREATE TABLE IF NOT EXISTS main.chats (time DateTime,chatID UUID,user String,text String) ENGINE = MergeTree() ORDER BY tuple()"
	err = chConn.Ping()
	if err != nil {
		logger.Error("Clickhouse conn ping", err)
		return nil, err
	}
	logger.Info("DB", "Clickhouse connected!")
	return &SendMessage{mgoSession: mgoSession, mgoColl: mgoCollection, clickhouseConn: chConn, clickhouseTable: clickhouseTable}, nil

}

func (conf *SendMessage) Close() error {
	conf.mgoSession.Close()
	return nil
}

func (conf *SendMessage) Handle(r *suckhttp.Request, l *logger.Logger) (*suckhttp.Response, error) {

	formValues, err := url.ParseQuery(string(r.Body))
	if err != nil {
		return suckhttp.NewResponse(400, "Bad Request"), err
	}

	text := formValues.Get("text")

	if text == "" {
		return suckhttp.NewResponse(400, "Bad Request"), err
	}

	cookie := lib.GetCookie(r.GetHeader(suckhttp.Cookie), "koki")
	if cookie == nil {
		return suckhttp.NewResponse(401, "Unauthorized"), nil
	}
	// TODO: AUTH

	userHash := formValues.Get("from") //TODO: delete this
	if userHash == "" {
		return suckhttp.NewResponse(400, "Bad Request"), nil
	}

	chat := formValues.Get("chat")
	if chat == "" {
		return suckhttp.NewResponse(400, "Bad Request"), nil
	}

	res := &chatInfo{}
	err = conf.mgoColl.Find(bson.M{"Id": chat}).One(&res) //
	if err != nil {
		if err == mgo.ErrNotFound {
			return suckhttp.NewResponse(400, "Bad Request"), err
		}
		return nil, err
	}

	query, err := clickhouse.BuildInsert(conf.clickhouseTable,
		clickhouse.Columns{"time", "chatID", "user", "text"},
		clickhouse.Row{time.Now(), chat, userHash, text})
	if err != nil {
		return nil, err
	}
	err = query.Exec(conf.clickhouseConn)
	if err != nil {
		return nil, err
	}

	return suckhttp.NewResponse(200, "OK"), nil
}
