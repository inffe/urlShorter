package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type urlStorage map[string]string

var DB *sql.DB

func main() {

	/*
		1) Флаг -d активирует работу с базой данных, результаты запросов POST будут храниться и в памяти, и в базе данных.
		При запросах GET значения берутся из базы данных;
		2) В памяти результаты хранятся в map в формате shortenedUrl:originalUrl.
	*/

	if len(os.Args) == 2 && os.Args[1] == "-d" {
		connStr := "user=postgres password=mypass dbname=urlsdb sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			panic(err)
		}
		DB = db
		defer DB.Close()
	}

	data := urlStorage{}

	http.HandleFunc("/", data.urlHandler)

	err := http.ListenAndServe("localhost:8080", nil)

	if err != nil {
		log.Fatal(err)
	}

}

func (data urlStorage) urlHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:

		fmt.Println("GET Request")

		shortenedUrl := r.URL.String()[1:]

		var originalUrl string

		if len(os.Args) == 2 && os.Args[1] == "-d" {
			err := DB.QueryRow("select originalurl from urls where shortenedurl = $1", shortenedUrl).Scan(&originalUrl)
			if err != nil {
				log.Println(err)
			}
		} else {
			originalUrl = data[shortenedUrl]
		}

		if originalUrl != "" {
			fmt.Fprint(w, originalUrl)
		} else {
			fmt.Fprint(w, "Unvalid URL-adress")
		}

	case http.MethodPost:

		fmt.Println("POST Request")

		originalUrl, err := io.ReadAll(r.Body)

		if err != nil {
			log.Println(err)
			fmt.Fprint(w, "Error occurred")
		} else {

			shortenedUrl := data.hash(string(originalUrl))

			data[shortenedUrl] = string(originalUrl)

			if len(os.Args) == 2 && os.Args[1] == "-d" {
				_, err := DB.Exec("insert into urls (shortenedUrl, originalUrl) values ($1, $2)", shortenedUrl, originalUrl)
				if err != nil {
					log.Println(err)
				}
			}
			fmt.Fprint(w, "http://localhost:8080/"+shortenedUrl)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

/*
	Хешируем исходный URL-адрес с помощью хэш-функции SHA-256, затем берем из получившейся строки 7 первых символов и
	полученную подстроку проверяем на совпадения. Если совпадений нет, то принимаем результат за скоращенную ссылку, если совпадение есть
	берем следующие 7 символов и проверяем вновь.
*/

func (data urlStorage) hash(originalUrl string) string {

	h := sha256.New()
	h.Write([]byte(originalUrl))
	shortenedUrl := hex.EncodeToString(h.Sum(nil))

	// проверка на уникальность
	finalResult := ""
	step := 7

	for i := 0; i < len(shortenedUrl); i += step {
		end := i + step

		/*
			Если shortenedUrl существует, то проверяем следующие 7 символов.
			Данная проверка позволяет избежать события, когда разные originalUrl при хешировании имеют одинаковые 7 символов.
		*/

		if _, ok := data[shortenedUrl[i:end]]; ok {

			/*
				Если shortenedUrl существует и его data[shortenedUrl[i:end]] равна рассматриваемому originalUrl, то повторный запрос
				POST с такой же originalUrl не будет повторно сокращен, но уже с другой подстрокой получившегося хеша.
			*/

			if data[shortenedUrl[i:end]] == originalUrl {
				finalResult = shortenedUrl[i:end]
				break
			}
			continue
		} else {
			finalResult = shortenedUrl[i:end]
			break
		}
	}
	return finalResult
}
