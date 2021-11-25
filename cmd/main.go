package main

import (
	"chatapp/internal"
	"chatapp/model"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

const (
	secretKey         = "eyJhbGciOiJIUzUxMiJ9.eyJJc3N1ZXIiOiJJc3N1ZXIiLCJVc2VybmFtZSI6IkROSUVBRCJ9.bf6d8Y8CKTrRKSJ3iv-Tp5nqe7p3G4o3ot5MI31BISfbMkOEs8QDnfgaEh7coTOVoNvNP0qsAKomEmoA3mH-sA"
	mongoDbConnection = "mongodb://localhost:27017"
)

var (
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan model.Message)
	upgrader  = websocket.Upgrader{}
	mh        *internal.MongoHandler
)

func main() {
	r := mux.NewRouter()
	mh = internal.NewHandler(mongoDbConnection)

	r.PathPrefix("/signin").Handler(http.StripPrefix("/signin", http.FileServer(http.Dir("../public/signin")))).Methods("GET")
	r.HandleFunc("/signin", signIn).Methods("POST")

	r.PathPrefix("/signup").Handler(http.StripPrefix("/signup", http.FileServer(http.Dir("../public/signup")))).Methods("GET")
	r.HandleFunc("/signup", signUp).Methods("POST")

	subroute := r.PathPrefix("/ws").Subrouter()
	subroute.HandleFunc("", handleConnections)
	subroute.Use(MiddlewareValidateUser)

	r.PathPrefix("/").Handler(MiddlewareValidateUser(http.FileServer(http.Dir("../public/main"))))

	go handleMessage()

	log.Println("http server started on :8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer ws.Close()

	clients[ws] = true
	for {
		var msg model.Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		broadcast <- msg
	}
}

func handleMessage() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func signUp(w http.ResponseWriter, r *http.Request) {
	var user model.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// checks if email is already register or not
	err = mh.GetOne(&user, bson.M{"email": user.Email})
	if err == nil {
		http.Error(w, "This email has been used", http.StatusBadRequest)
		return
	}

	user.Password, err = GeneratehashPassword(user.Password)
	if err != nil {
		log.Fatalf("error in password hash: %v", err)
	}

	// insert user details in database
	_, err = mh.AddOne(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func GeneratehashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateJWT(useName, email string) (string, error) {
	mySigningKey := []byte(secretKey)
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["usename"] = useName
	claims["email"] = email
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		fmt.Errorf("Something Went Wrong with the token: %v", err)
		return "", err
	}
	return tokenString, nil
}

func signIn(w http.ResponseWriter, r *http.Request) {
	var userLogin model.Authentication
	err := json.NewDecoder(r.Body).Decode(&userLogin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	check := false
	var storedUser model.User
	if userLogin.Email == "" {
		http.Error(w, "Empty email field", http.StatusBadRequest)
		return
	} else {
		err = mh.GetOne(&storedUser, bson.M{"email": userLogin.Email})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	check = CheckPasswordHash(userLogin.Password, storedUser.Password)
	if !check {
		err = errors.New("err: user is not stored in users")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	validToken, err := GenerateJWT(storedUser.Username, storedUser.Email)
	if err != nil {
		err = errors.New("err: can not generate jwt token")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	var token model.UserInfo
	token.Email = storedUser.Email
	token.Username = storedUser.Username
	w.Header().Set("Content-Type", "application/json")
	cookie := &http.Cookie{
		Name:  "jwt-token",
		Value: validToken,
		Path:  "/",
	}
	http.SetCookie(w, cookie)
	json.NewEncoder(w).Encode(token)
}

func MiddlewareValidateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var signInToken string = ""
		for _, cookie := range r.Cookies() {
			if cookie.Name == "jwt-token" {
				signInToken = cookie.Value
			}
		}

		if signInToken == "" {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		mySigningKey := []byte(secretKey)
		token, err := jwt.Parse(signInToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("There was an error in parsing")
			}
			return mySigningKey, nil
		})
		if err != nil {
			http.Error(w, "your token has been expired", http.StatusBadRequest)
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			r.Header.Set("username", claims["usename"].(string))
			r.Header.Set("email", claims["email"].(string))
			next.ServeHTTP(w, r)
			return

		}
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
	})
}
