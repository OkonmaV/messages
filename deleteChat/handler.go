package main

import (
	"lib"
	"net/url"
	"thin-peak/logs/logger"

	"github.com/big-larry/mgo"
	"github.com/big-larry/suckhttp"
	"go.mongodb.org/mongo-driver/bson"
)

type DeleteChat struct {
	MgoConn *mgo.Session
	MgoColl *mgo.Collection
}

func NewDeleteChat(mgoAddr string, mgoColl string) (*DeleteChat, error) {

	mgoSession, err := mgo.Dial(mgoAddr)
	if err != nil {
		logger.Error("Mongo conn", err)
		return nil, err
	}
	mgoCollection := mgoSession.DB("main").C(mgoColl)
	return &DeleteChat{MgoConn: mgoSession, MgoColl: mgoCollection}, nil
}

func (conf *DeleteChat) Close() error {
	conf.MgoConn.Close()
	return nil
}

func (conf *DeleteChat) Handle(r *suckhttp.Request, l *logger.Logger) (*suckhttp.Response, error) {

	cookie := lib.GetCookie(r.GetHeader(suckhttp.Cookie), "koki")
	if cookie == nil {
		return suckhttp.NewResponse(401, "Unauthorized"), nil
	}

	// TODO: AUTH?????

	formValues, err := url.ParseQuery(string(r.Body))
	if err != nil {
		return suckhttp.NewResponse(400, "Bad request"), err
	}

	chatId := formValues.Get("chatid")
	chatType := formValues.Get("type")
	if chatId == "" || chatType == "" {
		return suckhttp.NewResponse(400, "Bad request"), nil
	}

	// TODO: откуда хэш?
	userHash := formValues.Get("userhash")
	if userHash == "" {
		return suckhttp.NewResponse(400, "Bad request"), nil
	}
	//

	selector := &bson.M{}

	switch chatType {
	case "1":
		selector = &bson.M{"_id": chatId, "type": 1, "users": userHash}
	case "2":
		selector = &bson.M{"_id": chatId, "type": 2, "$or": []bson.M{{"users.0": userHash}, {"users.1": userHash}}}
	case "3":
		selector = &bson.M{"_id": chatId, "type": 3, "users.0": userHash}
	default:
		return suckhttp.NewResponse(400, "Bad request"), nil
	}
	change := mgo.Change{
		Update:    bson.M{"$mul": bson.M{"type": -1}, "$set": bson.M{"deleted.by": userHash}, "&currentDate": bson.M{"deleted.date": true}},
		Upsert:    false,
		ReturnNew: true,
		Remove:    false,
	}

	var foo interface{}

	info, err := conf.MgoColl.Find(selector).Apply(change, &foo)
	if err != nil {
		return nil, err
	}
	if info.Updated != 1 {
		return suckhttp.NewResponse(403, "Forbidden"), nil
	}

	return suckhttp.NewResponse(200, "OK"), nil
}
