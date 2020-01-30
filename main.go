package main

import (
	"flag"
	"fmt"
	"github.com/lon9/arxiv"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"io/ioutil"
	"strings"
)

func main() {
	var (
		dbPath     string
		genresPath string
	)
	flag.StringVar(&dbPath, "d", "arxiv.db", "Database path")
	flag.StringVar(&genresPath, "g", "genres.txt", "Genres path")
	flag.Parse()
	db, err := gorm.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	if db.HasTable(&arxiv.Paper{}) {
		db.DropTable(&arxiv.Paper{})
	}
	if db.HasTable(&arxiv.Author{}) {
		db.DropTable(&arxiv.Author{})
	}
	db.CreateTable(&arxiv.Author{})
	db.CreateTable(&arxiv.Paper{})

	b, err := ioutil.ReadFile(genresPath)
	if err != nil {
		panic(err)
	}
	genres := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	a := arxiv.NewCrawler(genres)

	papersCh, errCh, doneCh := a.StartCrawl()

L1:
	for {
		select {
		case papers := <-papersCh:
			fmt.Println("Paper received")
			for i := range papers {
				db.Create(&papers[i])
			}
		case err := <-errCh:
			fmt.Println(err)
		case <-doneCh:
			close(papersCh)
			close(errCh)
			close(doneCh)
			break L1
		}
	}
}
