package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/big-larry/mgo"
	utils "github.com/big-larry/suckutils"
	"go.mongodb.org/mongo-driver/bson"
)

type configs struct {
	mongoSession *mgo.Session
	mongoColl    *mgo.Collection
}

var ctx = context.Background()

func (cfg *configs) handler(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("koki")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	chatId := r.Form["chat"][0]
	if len(chatId) != 24 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	respAuth, err := http.Get(utils.ConcatFour("http://www.URL-TO-AUTH.com/deletechat?jwt=", cookie.Value, "&chat=", chatId))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer func() {
		err := respAuth.Body.Close()
		if err != nil {
			fmt.Println(err) //todo
			return
		}
	}()

	if respAuth.StatusCode != 200 {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// --- RAW STRING IN BODY IN RESPONCE ---

	/*	bytes, err := ioutil.ReadAll(respAuth.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		ownerHash := string(bytes)
		ownerHash = ownerHash*/

	// --- JSON IN RESPONCE ---

	/*	respAuthBodyDecoded := &ChatCreatorAuthResp{}
		err := json.NewDecoder(respAuth.Body).Decode(respAuthBodyDecoded)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}*/

	// --- ¯\_(ツ)_/¯ ---

	userHash := r.Form["hash"][0]

	// -----------
	updateData := bson.M{"$set": bson.M{"deletedBy": userHash}, "$mul": bson.M{"type": -1}}
	err = cfg.mongoColl.UpdateId(chatId, updateData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}
func main() {
	mongoSession, err := mgo.Dial("127.0.0.1")

	if err != nil {
		fmt.Println(err) //TODO
		return
	}
	defer mongoSession.Close()

	mongoCollection := mongoSession.DB("main").C("chats")

	cfg := *&configs{mongoSession: mongoSession, mongoColl: mongoCollection}

	http.HandleFunc("/", cfg.handler)
	log.Fatal(http.ListenAndServe(":8093", nil))
}
