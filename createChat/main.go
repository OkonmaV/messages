package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/beevik/guid"
	"github.com/big-larry/mgo"
	"github.com/big-larry/mgo/bson"
	utils "github.com/big-larry/suckutils"
)

type configs struct {
	mongoSession *mgo.Session
	mongoColl    *mgo.Collection
}
type ChatInfo struct {
	Id    string   `bson:"_id"`
	Users []string `bson:"users"`
	Name  string   `bson:"name"`
	Type  string   `bson:"type"`
}

type ChatCreatorAuthResp struct {
	ownerHash string `json:"hash"`
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
	chatType := r.Form["type"][0]

	switch chatType {
	case "self":
		//TODO
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
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// --- END REQUEST TO AUTH ---

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

		// --- HASH IN FORM IN MAIN REQUEST ---

		ownerHash := r.Form["hash"][0]

		// --- END OF KNOWING USER`S HASH ---
		uuid := guid.New()
		err = cfg.mongoColl.Insert(&ChatInfo{
			Id:    uuid.String(),
			Users: []string{ownerHash, r.Form["to"][0]},
			Name:  r.Form["name"][0],
			Type:  chatType})
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write([]byte(uuid.String())) // КУДА WRITE? КУДА WRITE ТО БЛТЬ?

	case "ls":

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
			w.WriteHeader(http.StatusUnauthorized)
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

		// --- HASH IN FORM IN MAIN REQUEST ---

		ownerHash := r.Form["hash"][0]

		// END OF KNOWING USER`S HASH AND CHECKING HIS RIGHTS

		uniq, resChat, err := cfg.IsThisChatUnique(ownerHash, r.Form["to"][0])
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if uniq {
			uuid := guid.New()
			resChat = &ChatInfo{
				Id:    uuid.String(),
				Users: []string{ownerHash, r.Form["to"][0]},
				Type:  chatType}
			err := cfg.mongoColl.Insert(resChat)
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		w.Write([]byte(resChat.Id)) // КУДА WRITE? КУДА WRITE ТО БЛТЬ?

	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func main() {
	mngSession, err := mgo.Dial("127.0.0.1")

	if err != nil {
		fmt.Println(err) //TODO
		return
	}
	defer mngSession.Close()

	mngCollection := mngSession.DB("main").C("chats")

	cfg := *&configs{mongoSession: mngSession, mongoColl: mngCollection}

	http.HandleFunc("/", cfg.handler)
	log.Fatal(http.ListenAndServe(":8091", nil))
}

func (cfg *configs) IsThisChatUnique(u1 string, u2 string) (bool, *ChatInfo, error) {
	res := &ChatInfo{}
	//var res []*ChatInfo maybe this
	err := cfg.mongoColl.Find(bson.M{"users": []string{u1, u2}}).One(res)
	if err != nil {
		return false, nil, err
	}
	err = cfg.mongoColl.Find(bson.M{"users": []string{u2, u1}}).One(res)
	if err != nil {
		return false, nil, err
	}
	return res == nil, res, nil
}
