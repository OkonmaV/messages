package main

import (
	"net/url"
	"strings"
	"time"

	"github.com/big-larry/mgo"
	"github.com/big-larry/mgo/bson"
	"github.com/big-larry/suckhttp"
	"github.com/big-larry/suckutils"
	"github.com/thin-peak/httpservice"
	"github.com/thin-peak/logger"
)

const thisServiceName = httpservice.ServiceName("CreateChat")

type ChatInfo struct {
	Id    []byte   `bson:"_id"`
	Users []string `bson:"users"`
	Name  string   `bson:"name"`
	Type  int      `bson:"type"`
}

func (handler *CreateChatHandler) Handle(r *suckhttp.Request) (*suckhttp.Response, error) {

	var token string
	cookie := r.GetHeader(suckhttp.Cookie)
	cookie = strings.ReplaceAll(cookie, " ", "")
	cookieValues := strings.Split(cookie, ";")
	for _, cv := range cookieValues {
		i := strings.Index(cv, "=")
		if name := cv[:i]; name == "token" {
			token = cv[i+1:]
			break
		}
	}
	// десериализация кук?
	println(token)

	queryValues, err := url.ParseQuery(r.Uri.RawQuery)
	if err != nil {
		return nil, err
	}
	chatType := queryValues.Get("type")

	switch chatType {
	case "self":

		// КУДА ТО ТАМ АВТОРИЗУЕМСЯ ¯\_(ツ)_/¯

		ownerHash := queryValues.Get("u1") // TODO: ОТКУДА ТО ПОЛУЧАЕМ ХЭШ ¯\_(ツ)_/¯

		uid := suckutils.GetRandUID(314)
		change := mgo.Change{
			Update:    bson.M{"$setOnInsert": bson.M{"_id": uid, "users": ownerHash, "type": 0}},
			Upsert:    true,
			ReturnNew: false,
			Remove:    false,
		}
		selector := &bson.M{"type": 0, "users": ownerHash}
		foundChat := &ChatInfo{}

		insertedChat, err := handler.mongoColl.Find(selector).Apply(change, foundChat)
		if err != nil {
			return nil, err
		}

		if foundChat.Id == nil {
			insertedChatId, ok := insertedChat.UpsertedId.([]byte)
			if ok && insertedChatId != nil {
				responce := suckhttp.NewResponse(200, "OK")
				responce.SetBody(insertedChatId) // TODO: КУДА WRITE?
				return responce, nil
			} else {
				return nil, nil // ??
			}
		}

		responce := suckhttp.NewResponse(200, "OK")
		responce.SetBody(foundChat.Id) // TODO: КУДА WRITE?
		return responce, nil

	case "group":

		// КУДА ТО ТАМ АВТОРИЗУЕМСЯ ¯\_(ツ)_/¯

		ownerHash := queryValues.Get("u1") // TODO: ОТКУДА ТО ПОЛУЧАЕМ ХЭШ ¯\_(ツ)_/¯

		chatName := queryValues.Get("name")

		if chatName == "" {
			chatName = "Group chat"
		}

		uid := suckutils.GetRandUID(314)
		change := mgo.Change{
			Update:    bson.M{"$setOnInsert": bson.M{"_id": uid, "users": ownerHash, "type": 2}},
			Upsert:    true,
			ReturnNew: false,
			Remove:    false,
		}
		selector := &bson.M{"type": 2, "users.0": ownerHash, "name": chatName}
		foundChat := &ChatInfo{}

		insertedChat, err := handler.mongoColl.Find(selector).Apply(change, foundChat)
		if err != nil {
			return nil, err
		}

		if foundChat.Id == nil {
			insertedChatId, ok := insertedChat.UpsertedId.([]byte)
			if ok && insertedChatId != nil {
				responce := suckhttp.NewResponse(200, "OK")
				responce.SetBody(insertedChatId) // TODO: КУДА WRITE?
				return responce, nil
			} else {
				return nil, nil // ??
			}
		}

		responce := suckhttp.NewResponse(200, "OK")
		responce.SetBody(foundChat.Id) // TODO: КУДА WRITE?
		return responce, nil

	case "ls":

		secondUserHash := queryValues.Get("u2")
		if secondUserHash == "" {
			return suckhttp.NewResponse(400, "Miss second user"), nil
		}

		// КУДА ТО ТАМ АВТОРИЗУЕМСЯ ¯\_(ツ)_/¯

		ownerHash := queryValues.Get("u1") // TODO: ОТКУДА ТО ПОЛУЧАЕМ ХЭШ ¯\_(ツ)_/¯

		uid := suckutils.GetRandUID(314)
		change := mgo.Change{
			Update:    bson.M{"$setOnInsert": bson.M{"_id": uid, "users": []string{ownerHash, secondUserHash}, "type": 1}},
			Upsert:    true,
			ReturnNew: false,
			Remove:    false,
		}
		selector := &bson.M{"type": 1, "$or": []bson.M{{"users": []string{ownerHash, secondUserHash}}, {"users": []string{secondUserHash, ownerHash}}}}
		foundChat := &ChatInfo{}

		insertedChat, err := handler.mongoColl.Find(selector).Apply(change, foundChat)
		if err != nil {
			return nil, err
		}

		if foundChat.Id == nil {
			insertedChatId, ok := insertedChat.UpsertedId.([]byte)
			if ok && insertedChatId != nil {
				responce := suckhttp.NewResponse(200, "OK")
				responce.SetBody(insertedChatId) // TODO: КУДА WRITE?
				return responce, nil
			} else {
				return nil, nil // ??
			}
		}

		responce := suckhttp.NewResponse(200, "OK")
		responce.SetBody(foundChat.Id) // TODO: КУДА WRITE?
		return responce, nil

	default:
		return suckhttp.NewResponse(400, "Type error"), nil
	}
}

func main() {
	ctx, cancel := httpservice.CreateContextWithInterruptSignal()
	logger.SetupLogger(ctx, time.Second*2, []logger.LogWriter{logger.NewConsoleLogWriter(logger.DebugLevel)})
	defer func() {
		cancel()
		<-logger.AllLogsFlushed
	}()

	handler, err := NewCreateChatHandler("127.0.0.1")
	if err != nil {
		logger.Error("Mongo connection", err)
		return
	}

	logger.Error("HTTP service", httpservice.ServeHTTPService(ctx, ":8090", handler)) // TODO: отхардкодить порт?
}

type CreateChatHandler struct {
	mongoColl *mgo.Collection
}

func (handler *CreateChatHandler) Close() {
	handler.mongoColl.Database.Session.Close()
}

func NewCreateChatHandler(connectionString string) (*CreateChatHandler, error) { // порнография?
	mongoSession, err := mgo.Dial(connectionString)

	if err != nil {
		return nil, err
	}

	return &CreateChatHandler{mongoColl: mongoSession.DB("main").C("chats")}, nil
}
