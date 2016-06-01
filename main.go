package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/Rompei/arxiv"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"io"
	"os"
)

func main() {
	var (
		dbPath string
	)
	flag.StringVar(&dbPath, "d", "arxiv.db", "Database path")
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
	a := arxiv.NewCrawler()

	reader := bufio.NewReaderSize(os.Stdin, 4096)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		a.AddGenre(string(line))
	}

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
