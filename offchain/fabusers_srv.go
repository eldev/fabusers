/*
 
 You have to install
    1) mgo package (MongoDB driver for Golang)
        go get gopkg.in/mgo.v2
    2) mgo mgo.bson package (binary json)
           go get gopkg.in/mgo.v2/bson
    3) goji
          go get goji.io
 */

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
    "encoding/hex"

	"goji.io"
	"goji.io/pat"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"


	"./crypdata"
    "./onchain"
)

const (
    DB_NAME               = "fabusers"
    USERS_COLLECTION_NAME = "users"
)


func ErrorWithJSON(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "{message: %q}", message)
}

func ResponseWithJSON(w http.ResponseWriter, json []byte, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(json)
}

// the service handles requests that consist of JSON objects
// These JSON objects have to match to this struct 
type UserInfo struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`

	Privdata []string `json:"priv_data"`
}


// offchain db record corresponds to this struct
// db index field is Userhash field
type CipheredUserInfo struct {
    // Userhash is a hash value of all user info (Username + Email + Hashedpassword + Privdata)
	Userhash        string

	Username        string
	Email           string
	Hashedpassword  string
	Privdata        string
}


// the service loop function
func main() {
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	ensureIndex(session)

	// init crypdata package
	err = crypdata.Init()
	if err != nil {
		panic(err)
	}

	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/users"), allUsers(session))                      // ONLY for DEBUG!
	mux.HandleFunc(pat.Post("/users"), AddUser(session))
	mux.HandleFunc(pat.Get("/users/:username"), UserByUsername(session))
	mux.HandleFunc(pat.Get("/userhashes/:userhash"), userByUserhash(session)) // ONLY for DEBUG!
	mux.HandleFunc(pat.Put("/users/:username"), UpdateUser(session))

	http.ListenAndServe("localhost:8080", mux)
}



func ensureIndex(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()

	c := session.DB(DB_NAME).C(USERS_COLLECTION_NAME)

    // index of offchain database is userhash field
	index := mgo.Index {
		Key:        []string{"userhash"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err := c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}


// allUsers() receives all records (users info) in the offchain database
// NOTE: this function is ONLY for DEBUGGING purposes
func allUsers(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		c := session.DB(DB_NAME).C(USERS_COLLECTION_NAME)

		var users []CipheredUserInfo
		err := c.Find(bson.M{}).All(&users)
		if err != nil {
			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed get all users: ", err)
			return
		}

		respBody, err := json.MarshalIndent(users, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, respBody, http.StatusOK)
	}
}


// createCipheredUserinfo() is an auxiliary function
// it builds (copy some info, computes hash of password, encrypt private data)
// CipheredUserInfo struct from UserInfo
func createCipheredUserinfo(userInfo *UserInfo, cipheredUserInfo *CipheredUserInfo) error {
    // 1. build one string object from array of UserInfo.Privdata strings
	var userPrivData string = "["
	for ind, pdata := range userInfo.Privdata {
        if ind > 0 {
            // we should separate private data strings
            userPrivData += ", "
        }
		userPrivData += pdata
	}
    userPrivData += "]"

    // 2. encrypt private data
	ciphertext, err := crypdata.Encrypt([]byte(userPrivData))
	if err != nil {
		return err
	}

	cipheredUserInfo.Username = userInfo.Username
	cipheredUserInfo.Email = userInfo.Email
	cipheredUserInfo.Hashedpassword = crypdata.Hash(userInfo.Password)
	cipheredUserInfo.Privdata = hex.EncodeToString(ciphertext) // convert to string representation

    // Userhash field is a hash of all user info
	cipheredUserInfo.Userhash = crypdata.Hash(cipheredUserInfo.Username +
		cipheredUserInfo.Email +
		cipheredUserInfo.Hashedpassword +
		cipheredUserInfo.Privdata)

	return nil
}


// AddUser() takes new user info (as JSON object in the request),
// builds ciphered user info and saves this record to the offchain db
func AddUser(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

        // 1. Decode input json object
		var user UserInfo
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&user)
		if err != nil {
			ErrorWithJSON(w, "Incorrect body", http.StatusBadRequest)
			return
		}

		c := session.DB(DB_NAME).C(USERS_COLLECTION_NAME)

        // 2. Builds ciphered user info
		var cipheredUserInfo CipheredUserInfo
		err = createCipheredUserinfo(&user, &cipheredUserInfo)
		if err != nil {
			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed insert user: ", err)
			return
		}

		// 3. Register the new user in the onchain part (create ca-cert)
		err = onchain.RegisterUser(&cipheredUserInfo.Username)
        if err != nil {
            ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
            log.Println("Failed register user: ", err)
            return
        }

        // 4. Store offchain data for the new user
        err = c.Insert(cipheredUserInfo)
        if err != nil {
            if mgo.IsDup(err) {
                ErrorWithJSON(w, "User with this userhash already exists", http.StatusBadRequest)
                return
            }

            ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
            log.Println("Failed insert user: ", err)
            return
        }

		// 5. Add record (username + userhash) into onchain ledger
		err = onchain.AddUserInfoToLedger(&cipheredUserInfo.Username, &cipheredUserInfo.Userhash)
        if err != nil {
            ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
            log.Println("Failed add UserInfo to ledger: ", err)
            return
        }

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", r.URL.Path+"/"+cipheredUserInfo.Username)
		w.WriteHeader(http.StatusCreated)
	}
}


// UserByUsername() finds offchain database record with the specified userhash
// and decrypt its private data
func UserByUsername(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		var err error

        // 1. Extract username and password for this user
		username := pat.Param(r, "username")
		keys, ok := r.URL.Query()["password"]
		if !ok || len(keys) < 1 {
			log.Println("Url param doesn't have password")
			return
		}
		password := keys[0]

        // 2. Get userhash from onchain part (see onchain package)
		userhash, err := onchain.GetUserhash(&username)
		if err != nil {
			ErrorWithJSON(w, "get_userhash error", http.StatusInternalServerError)
			log.Println("Failed find user: ", err)
			return
		}

		c := session.DB(DB_NAME).C(USERS_COLLECTION_NAME)

        // 3. Find the offchain db record with this userhash
		var user CipheredUserInfo
		err = c.Find(bson.M{"userhash": userhash}).One(&user)
		if err != nil {
			ErrorWithJSON(w, "can't find userhash", http.StatusInternalServerError)
			log.Println("Failed find user: ", err)
			return
		}

		if user.Username == "" {
			ErrorWithJSON(w, "User is not found", http.StatusNotFound)
			return
		}

		// 4. If a hash of specified password matches the saved password hash,
        //    then service should decrypt private data.
        //    The user has access to his private data!
		if crypdata.Hash(password) == user.Hashedpassword {
			privDataDecodedBytes, error := hex.DecodeString(user.Privdata)
			plaintext, error := crypdata.Decrypt(privDataDecodedBytes)
			if error != nil {
				ErrorWithJSON(w, "Decrypt error", http.StatusInternalServerError)
				log.Println("Failed find user: ", error)
				return
			}
			user.Privdata = string(plaintext)
		}

		respBody, err := json.MarshalIndent(user, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, respBody, http.StatusOK)
	}
}



// userByUserhash() finds offchain database record with the specified userhash
// and decrypt its private data
// NOTE: this function is ONLY for DEBUGGING purposes
func userByUserhash(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        session := s.Copy()
        defer session.Close()

        // 1. Extract userhash and password for this user
        userhash := pat.Param(r, "userhash")
        keys, ok := r.URL.Query()["password"]
        if !ok || len(keys) < 1 {
            log.Println("Url param doesn't have password")
            return
        }
        password := keys[0]

        c := session.DB(DB_NAME).C(USERS_COLLECTION_NAME)

        // 2. Find user with the specified userhash
        var user CipheredUserInfo
        err := c.Find(bson.M{"userhash": userhash}).One(&user)
        if err != nil {
            ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
            log.Println("Failed find user: ", err)
            return
        }

        if user.Username == "" {
            ErrorWithJSON(w, "User is not found", http.StatusNotFound)
            return
        }

        // 3. If a hash of specified password matches the saved password hash,
        //    then service should decrypt private data
        if crypdata.Hash(password) == user.Hashedpassword {
            privDataDecodedBytes, error := hex.DecodeString(user.Privdata)
            plaintext, error := crypdata.Decrypt(privDataDecodedBytes)
            if error != nil {
                ErrorWithJSON(w, "Decrypt error", http.StatusInternalServerError)
                log.Println("Failed find user: ", error)
                return
            }
            user.Privdata = string(plaintext)
        }

        respBody, err := json.MarshalIndent(user, "", "  ")
        if err != nil {
            log.Fatal(err)
        }

        ResponseWithJSON(w, respBody, http.StatusOK)
    }
}


// UpdateUser() finds offchain database record with the specified userhash
// and decrypt its private data
func UpdateUser(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        session := s.Copy()
        defer session.Close()

        var err error

        // 1. Extract username and password for this user
        username := pat.Param(r, "username")
        keys, ok := r.URL.Query()["password"]
        if !ok || len(keys) < 1 {
            log.Println("Url param doesn't have password")
            return
        }
        password := keys[0]

        // 2. Find this user's userhash in onchain part (Hyperledger Fabric)
        userhash, err := onchain.GetUserhash(&username)
        if err != nil {
            ErrorWithJSON(w, "get_userhash error", http.StatusInternalServerError)
            log.Println("Failed find user: ", err)
            return
        }

        c := session.DB(DB_NAME).C(USERS_COLLECTION_NAME)

        // 3. Find the offchain db record with this userhash
        var cryptoUser CipheredUserInfo
        err = c.Find(bson.M{"userhash": userhash}).One(&cryptoUser)
        if err != nil {
            ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
            log.Println("Failed find user: ", err)
            return
        }

        if cryptoUser.Username == "" {
            ErrorWithJSON(w, "User is not found", http.StatusNotFound)
            return
        }

        // 4. TODO: ONLY admin can change this data
        // but for debugging purposes, at this moment user also has right
        // to change private data
        if crypdata.Hash(password) != cryptoUser.Hashedpassword {
            ErrorWithJSON(w, "password is wrong", http.StatusInternalServerError)
            log.Println("password is wrong")
            return
        }
        

        // 5. Take new user data (as json object)
        var user UserInfo
        decoder := json.NewDecoder(r.Body)
        err = decoder.Decode(&user)

        if err != nil {
            ErrorWithJSON(w, "Incorrect body", http.StatusBadRequest)
            return
        }

        // 6. Create new crypto data
        err = createCipheredUserinfo(&user, &cryptoUser)
        if err != nil {
            ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
            log.Println("Failed create crypto user: ", err)
            return
        }

        // 7. Update the ledger
        err = onchain.UpdateLedgerUserinfo(&cryptoUser.Username, &cryptoUser.Userhash)
        if err != nil {
            ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
            log.Println("Update ledger user info error: ", err)
            return
        }

        // 8. Update the offchain db
        err = c.Update(bson.M{"userhash": userhash}, &cryptoUser)
        if err != nil {
            ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
            log.Println("Update db error: ", err)
            return
        }

        w.WriteHeader(http.StatusNoContent)
    }
}

