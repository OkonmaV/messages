package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/big-larry/mgo"
	"github.com/big-larry/mgo/bson"
	utils "github.com/big-larry/suckutils"
)

type configs struct {
	mongoSession *mgo.Session
	mongoColl    *mgo.Collection
}
type ChatInfo struct {
	Id    bson.ObjectId `bson:"_id"`
	Users []string      `bson:"users"`
	Name  string        `bson:"name"`
	Type  int           `bson:"type"`
}

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
	chatType := r.Form["type"][0]

	switch chatType {
	case "self":

		respAuth, err := http.Get(utils.ConcatTwo("http://www.URL-TO-AUTH.com/?jwt=", cookie.Value))
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

		ownerHash := r.Form["hash"][0]

		upsertData := bson.M{"$setOnInsert": bson.M{"users": ownerHash, "type": 0}}
		upsertSelector := bson.M{"type": 0, "users": ownerHash}
		insertResult, err := cfg.mongoColl.Upsert(upsertSelector, upsertData)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if insertResult.UpsertedId == nil {
			findResult := &ChatInfo{}
			err = cfg.mongoColl.Find(bson.M{"type": 0, "users": ownerHash}).One(findResult)
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if findResult != nil {
				w.Write([]byte(findResult.Id.Hex())) // КУДА WRITE?
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
		}
		chatId, ok := insertResult.UpsertedId.(bson.ObjectId)
		if ok {
			w.Write([]byte(chatId.Hex())) // КУДА WRITE?
			return
		}

		w.WriteHeader(http.StatusInternalServerError)

	case "group":

		// --- START REQUEST TO AUTH ---

		// --- USER'S COOKIE IN REQ COOKIES ---

		/*reqAuth, err := http.NewRequest("GET", "http://www.URL-TO-AUTH.com/", nil) //TODO: get auth like chelovek
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		reqAuth.AddCookie(cookie)

		client := &http.Client{}
		respAuth, err := client.Do(reqAuth)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}*/

		// --- REQUEST TO AUTH WITH USER'S JWT IN QUERY ---

		respAuth, err := http.Get(utils.ConcatTwo("http://www.URL-TO-AUTH.com/?jwt=", cookie.Value))
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
		// --- END REQUEST TO AUTH ---

		// --- HASH IN BODY OF RESPONCE ---

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

		// --- HASH IN FORM IN MAIN REQUEST ---

		ownerHash := r.Form["hash"][0]

		// --- END OF KNOWING USER`S HASH ---

		chatName := r.Form["name"][0]

		if chatName == "" {
			chatName = "Group chat"
		}

		upsertData := bson.M{"$setOnInsert": bson.M{"users": ownerHash, "type": 2}}
		upsertSelector := bson.M{"type": 2, "users.0": ownerHash, "name": chatName}
		insertResult, err := cfg.mongoColl.Upsert(upsertSelector, upsertData)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if insertResult.UpsertedId == nil {
			findResult := &ChatInfo{}
			err = cfg.mongoColl.Find(bson.M{"type": 0, "users.0": ownerHash, "name": chatName}).One(findResult)
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if findResult != nil {
				w.Write([]byte(findResult.Id.Hex())) // КУДА WRITE?
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
		}
		chatId, ok := insertResult.UpsertedId.(bson.ObjectId)
		if ok {
			w.Write([]byte(chatId.Hex())) // КУДА WRITE?
			return
		}

		w.WriteHeader(http.StatusInternalServerError)

	case "ls":

		secondUserHash := r.Form["to"][0]
		if secondUserHash == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// --- START REQUEST TO AUTH ---

		// --- REQUEST TO AUTH WITH KOKI IN COOKIE ---

		/*reqAuth, err := http.NewRequest("GET", "http://www.URL-TO-AUTH.com/", nil) //TODO: get auth like chelovek
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		reqAuth.AddCookie(cookie)

		client := &http.Client{}
		respAuth, err := client.Do(reqAuth)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}*/

		// --- REQUEST TO AUTH WITH USER'S JWT IN QUERY ---

		respAuth, err := http.Get(utils.ConcatFour("http://www.URL-TO-AUTH.com/?jwt=", cookie.Value, "&to=", r.Form["to"][0]))
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
		// --- END REQUEST TO AUTH ---

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

		ownerHash := r.Form["hash"][0]

		// -----------

		upsertData := bson.M{"$setOnInsert": bson.M{"users": []string{ownerHash, secondUserHash}, "type": 1}}
		upsertSelector := bson.M{"type": 1, "users": bson.M{"$all": []bson.M{{"$elemMatch": bson.M{"$eq": ownerHash}}, {"$elemMatch": bson.M{"$eq": secondUserHash}}}}}
		insertResult, err := cfg.mongoColl.Upsert(upsertSelector, upsertData)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if insertResult.UpsertedId == nil {
			findResult := &ChatInfo{}
			err = cfg.mongoColl.Find(bson.M{"type": 1, "users": bson.M{"$all": []string{ownerHash, secondUserHash}}}).One(findResult)
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if findResult != nil {
				w.Write([]byte(findResult.Id.Hex())) // КУДА WRITE?
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
		}

		chatId, ok := insertResult.UpsertedId.(bson.ObjectId)
		if ok {
			w.Write([]byte(chatId.Hex())) // КУДА WRITE?
			return
		}

		w.WriteHeader(http.StatusInternalServerError)

	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
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
	log.Fatal(http.ListenAndServe(":8091", nil))
}
