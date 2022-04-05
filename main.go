// package main

// import (
// 	"database/sql"
// 	"encoding/json"
// 	"log"
// 	"net/http"

// 	_ "github.com/go-sql-driver/mysql"
// )

// type City struct {
// 	Id         int
// 	Name       string
// 	Population int
// }

// type Book struct {
// 	Title  string `json:"title"`
// 	Author string `json:"author"`
// 	Pages  int    `json:"pages"`
// }

// func Hello(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "text/html")
// 	w.Write([]byte("<h1 style='color:steelblue'>Hello</h1>"))
// }

// func getBook(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	book := Book{
// 		Title:  "The Gunslinger",
// 		Author: "Stephen King",
// 		Pages:  304}

// 	json.NewEncoder(w).Encode(book)

// }

// func main() {

// 	http.HandleFunc("/hello", Hello)
// 	http.HandleFunc("/book", getBook)

// 	db, err := sql.Open("mysql", "root:floatsky65@tcp(127.0.0.1:3306)/sys")
// 	defer db.Close()

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	log.Fatal(http.ListenAndServe(":5100", nil))
// }

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Urls struct {
	Url      string `json:"url"`
	ExpireAt string `json:"expireAt"`
}

type ResponseUrls struct {
	Id       string `json:"id"`
	ShortUrl string `json:"shortUrl"`
}

type UrlRows struct {
	Id          int
	OriginalUrl string
	ShortenUrl  string
	Date        string
}

func shortenUrl(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/urls" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "POST":
		var u Urls
		err := json.NewDecoder(r.Body).Decode(&u)
		fmt.Println(u.ExpireAt[0:9])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		//fmt.Fprintf(w, "Urls: %+v", u)
		res, err := db.Query("SELECT * FROM urls")
		if err != nil {
			log.Fatal(err)
		}
		id := 0
		for res.Next() {
			var url UrlRows
			err := res.Scan(&url.Id, &url.OriginalUrl, &url.ShortenUrl, &url.Date)
			if id < url.Id {
				id = url.Id
			}

			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%v\n", url)
		}
		newUrl := transTo62(int64(id))
		//sql := "INSERT INTO urls(originalUrl, shortenUrl, date) VALUES ('https://www.mysqltutorial.org/mysql-add-column/', 'gary/dsadaf','2021/04/30')"
		sql := fmt.Sprintf("INSERT INTO urls(originalUrl, shortenUrl, date) VALUES ('%s', 'http://localhost/%s','%s')", u.Url, newUrl, u.ExpireAt[0:10])
		res2, err := db.Exec(sql)
		if err != nil {
			panic(err.Error())
		}

		lastId, err := res2.LastInsertId()

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("The last inserted row id: %d\n", lastId)
		w.Header().Set("Content-Type", "application/json")
		urlResponse := ResponseUrls{
			Id:       newUrl,
			ShortUrl: "http://localhost/" + newUrl}

		json.NewEncoder(w).Encode(urlResponse)

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func redirect(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	sqlLine := fmt.Sprintf("SELECT * FROM urls where shortenUrl = 'http://localhost%s'", r.URL.Path)
	fmt.Println(sqlLine)
	var url UrlRows
	err := db.QueryRow(sqlLine).Scan(&url.Id, &url.OriginalUrl, &url.ShortenUrl, &url.Date)
	switch {
	case err == sql.ErrNoRows:
		log.Printf("Not found.")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "custom 404")
	case err != nil:
		log.Fatal(err)
	default:
		//do stuff
		fmt.Println(url.OriginalUrl)
		http.Redirect(w, r, url.OriginalUrl, http.StatusFound)
	}
}

func main() {
	http.HandleFunc("/api/v1/urls", shortenUrl)
	http.HandleFunc("/", redirect)
	var err error
	db, err = sql.Open("mysql", "root:floatsky65@tcp(127.0.0.1:3306)/sys")
	defer db.Close()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Starting server for testing HTTP POST...\n")
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatal(err)
	}
}

// 將十進位制轉換為62進位制   0-9a-zA-Z 六十二進位制
func transTo62(id int64) string {
	// 1 -- > 1
	// 10-- > a
	// 61-- > Z
	charset := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var shortUrl []byte
	for {
		var result byte
		number := id % 62
		result = charset[number]
		var tmp []byte
		tmp = append(tmp, result)
		shortUrl = append(tmp, shortUrl...)
		id = id / 62
		if id == 0 {
			break
		}
	}
	fmt.Println(string(shortUrl))
	return string(shortUrl)
}

//curl -X POST -H "Content-Type:application/json" http://localhost/api/vi/urls -d '{"url":"https://ithelp.ithome.com.tw/articles/10254833","expireAt":"2021-02-08T09:20:4171"}'
