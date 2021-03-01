package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type configs struct {
	mongoConn *mongo.Client
	mongoColl *mongo.Collection
}
type ChatInfo struct {
	Id    string   `bson:"_id"`
	Users []string `bson:"users"`
	Name  string   `bson:"name"`
	Type  string   `bson:"type"`
}

var ctx = context.Background()

func (cfg *configs) handler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	var users []string
	users = r.Form["u"]
	name := r.Form["n"][0] //todo ЧИТАТЬ НАЗВАНИЕ КАК ЧЕЛОВЕК

	chatInfo := &ChatInfo{ /*Id: guid.New().String(),*/ Name: name, Users: users}
	if len(users) > 2 {
		chatInfo.Type = "1"
	} else {
		chatInfo.Type = "0"
	}

	//fmt.Println("WRITING TO MNG")
	_, err := cfg.mongoColl.InsertOne(ctx, chatInfo)
	if err != nil {
		fmt.Println(err) //todo
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//fmt.Println("WRITED TO MNG")
	w.WriteHeader(http.StatusOK)
}

func main() {
	connMng, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		fmt.Println(err) //todo
		return
	}
	defer func() {
		if connMng.Disconnect(ctx) != nil {
			fmt.Println(err)
		}
	}()

	collectionMng := connMng.Database("main").Collection("chats")

	cfg := *&configs{mongoConn: connMng, mongoColl: collectionMng}

	http.HandleFunc("/", cfg.handler)
	log.Fatal(http.ListenAndServe(":8092", nil))
}
