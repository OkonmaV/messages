package main

import (
	"lib"
	"net/url"

	"thin-peak/logs/logger"

	"github.com/big-larry/mgo"
	"github.com/big-larry/mgo/bson"
	"github.com/big-larry/suckhttp"
	"github.com/rs/xid"
)

type CreateChat struct {
	mgoSession *mgo.Session
	mgoColl    *mgo.Collection
}
type chatInfo struct {
	//Id bson.ObjectId `bson:"_id"`
	Id    string   `bson:"_id"`
	Users []string `bson:"users"`
	Name  string   `bson:"name"`
	Type  int      `bson:"type"`
}

func NewCreateChat(mgoAddr string, mgoColl string) (*CreateChat, error) {

	mgoSession, err := mgo.Dial(mgoAddr)
	if err != nil {
		logger.Error("Mongo conn", err)
		return nil, err
	}

	mgoCollection := mgoSession.DB("main").C(mgoColl)

	return &CreateChat{mgoSession: mgoSession, mgoColl: mgoCollection}, nil

}

func (conf *CreateChat) Close() error {
	conf.mgoSession.Close()
	return nil
}

func getChatRandId() string {
	return xid.New().String()
}

func (conf *CreateChat) Handle(r *suckhttp.Request, l *logger.Logger) (*suckhttp.Response, error) {

	cookie := lib.GetCookie(r.GetHeader(suckhttp.Cookie), "koki")
	if cookie == nil {
		return suckhttp.NewResponse(401, "Unauthorized"), nil
	}

	formValues, err := url.ParseQuery(string(r.Body))
	if err != nil {
		return suckhttp.NewResponse(400, "Bad request"), err
	}

	ownerHash := formValues.Get("userId") // TODO: ОТКУДА ТО ПОЛУЧАЕМ ХЭШ ¯\_(ツ)_/¯
	if len(ownerHash) != 32 {
		return suckhttp.NewResponse(400, "Bad request"), nil
	}

	chatType := formValues.Get("type")
	switch chatType {
	case "1": // сам с собой

		insertId := getChatRandId()

		change := mgo.Change{
			Update:    bson.M{"$setOnInsert": bson.M{"_id": insertId, "users": ownerHash, "type": 1}},
			Upsert:    true,
			ReturnNew: false,
			Remove:    false,
		}
		selector := &bson.M{"type": 1, "users": ownerHash}

		foundChat := &chatInfo{}
		_, err := conf.mgoColl.Find(selector).Apply(change, &foundChat)
		if err != nil {
			return nil, err
		}

		if foundChat.Id == "" {
			foundChat.Id = insertId
			// resp := suckhttp.NewResponse(200, "OK")
			// resp.SetBody([]byte(insertId)) // TODO: КУДА WRITE? Редирект?
			// return resp, nil
		}

		resp := suckhttp.NewResponse(200, "OK")
		resp.SetBody([]byte(foundChat.Id)) // TODO: КУДА WRITE? Редирект?
		return resp, nil

	case "3": // групповой

		chatName := formValues.Get("name")
		if chatName == "" {
			chatName = "Group chat"
		}

		insertId := getChatRandId()

		change := mgo.Change{
			Update:    bson.M{"$setOnInsert": bson.M{"_id": insertId, "users": []string{ownerHash}, "type": 3}},
			Upsert:    true,
			ReturnNew: false,
			Remove:    false,
		}
		selector := &bson.M{"type": 3, "users.0": ownerHash, "name": chatName}

		foundChat := &chatInfo{}

		_, err := conf.mgoColl.Find(selector).Apply(change, &foundChat)
		if err != nil {
			return nil, err
		}

		if foundChat.Id == "" {
			foundChat.Id = insertId
		}

		responce := suckhttp.NewResponse(200, "OK")
		responce.SetBody([]byte(foundChat.Id)) // TODO: КУДА WRITE?
		return responce, nil

	case "2": // между двумя

		secondUserHash := formValues.Get("seconduser")
		if len(secondUserHash) != 32 {
			return suckhttp.NewResponse(400, "Bad request"), nil
		}

		// КУДА ТО ТАМ АВТОРИЗУЕМСЯ ¯\_(ツ)_/¯

		insertId := getChatRandId()
		change := mgo.Change{
			Update:    bson.M{"$setOnInsert": bson.M{"_id": insertId, "users": []string{ownerHash, secondUserHash}, "type": 2}},
			Upsert:    true,
			ReturnNew: false,
			Remove:    false,
		}
		selector := &bson.M{"type": 2, "$or": []bson.M{{"users": []string{ownerHash, secondUserHash}}, {"users": []string{secondUserHash, ownerHash}}}}

		foundChat := &chatInfo{}
		_, err := conf.mgoColl.Find(selector).Apply(change, &foundChat)
		if err != nil {
			return nil, err
		}

		if foundChat.Id == "" {
			foundChat.Id = insertId
		}

		responce := suckhttp.NewResponse(200, "OK")
		responce.SetBody([]byte(foundChat.Id)) // TODO: КУДА WRITE?
		return responce, nil

	default:
		return suckhttp.NewResponse(400, "Bad request"), nil
	}
}
