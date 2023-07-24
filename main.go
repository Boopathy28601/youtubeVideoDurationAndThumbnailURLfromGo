package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	dbhost     = "localhost"
	dbport     = 5432
	dbuser     = "YazhBoopathy"
	dbpassword = "***"
	dbname     = "yazhapp"
)

type YazhApi struct {
	db *sql.DB
}

type Output struct {
	Items []Items `json:"items"`
}

type Items struct {
	ContentDetails ContentDetails `json:"contentDetails"`
	Snippet        Snippet        `json:"snippet"`
}

type Snippet struct {
	Thumbnails Thumbnails `json:"thumbnails"`
}

type Thumbnails struct {
	Default Default `json:"default"`
}

type ContentDetails struct {
	Duration string `json:"duration"`
}

type Default struct {
	Url string `json:"url"`
}

func initdb() (*sql.DB, error) {
	psqlcon := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbhost,
		dbport,
		dbuser,
		dbpassword, dbname)
	db, err := sql.Open("postgres", psqlcon)
	if err != nil {
		log.Error().Msgf("db connection failed: %v", err)
		return nil, err
	}
	return db, nil
}
func main() {
	var err error
	yapi := YazhApi{}
	yapi.db, err = initdb()
	if err != nil {
		log.Fatal().Msgf("db init failed: %s", err)
	}
	log.Logger = log.With().Caller().Logger().Output(
		zerolog.ConsoleWriter{Out: os.Stderr})

	var sourceUrl string
	// var newDuration string
	var url string
	for i := 1; i <= 100; i++ {
		row := yapi.db.QueryRow("select sourceurl from song where id = $1", i)
		err := row.Scan(&sourceUrl)
		if err == sql.ErrNoRows {
			return
		}
		if err != nil {
			log.Error().Msgf("reading record failed from public.user: %s", err)
			return
		}
		// print(sourceUrl)
		videoId := sourceUrl[32:]
		// print(videoId)
		youtubeDetails, err := http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/videos?part=snippet,contentDetails&id="+videoId+"&key=AIzaSyAYGSn655r3AdsD63zqIHdP6mXz3RsR0g4", nil)
		if err != nil {
			log.Fatal().Msgf("Start failed %s \n", err)
			return
		}

		c := http.Client{}

		resp, err := c.Do(youtubeDetails)
		if err != nil {
			log.Fatal().Msgf("Start failed %s \n", err)
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		var output Output
		json.Unmarshal(body, &output)
		var duration string
		for _, value := range output.Items {
			duration = value.ContentDetails.Duration
			url = value.Snippet.Thumbnails.Default.Url
			fmt.Println(url)
			// fmt.Println(duration)
		}
		min := duration[2:3]
		sec := duration[4:]
		if sec == "" {
			sec = "00"
			sec = fmt.Sprintf("%sS", sec)
		}
		if len(sec) == 2 {
			sec = fmt.Sprintf("0%s", sec)
		}
		formattedDuration := fmt.Sprintf("%s:%s", min, sec)
		newDuration := formattedDuration[:4]
		fmt.Println(newDuration)

		_, er := yapi.db.Exec("UPDATE song SET duration = $1 where id=$2", newDuration, i)
		if er != nil {
			log.Error().Msgf("adding record failed in song table: %s", err)
			return
		}
	}
}
