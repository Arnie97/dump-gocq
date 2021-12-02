package main

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Mrs4s/MiraiGo/message"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func dumpGoCQ(dbPath, jsonPath string) {
	gob.Register(message.Sender{})

	db, err := leveldb.OpenFile(dbPath, &opt.Options{
		WriteBuffer: 128 * opt.KiB,
	})
	if err != nil {
		checkError(err)
		return
	}

	jsonFile, err := os.Create(os.Args[2])
	if err != nil {
		checkError(err)
		return
	}

	for _, parsedDB := range parsedDatabases {
		ret := make([]map[string]interface{}, 0, len(parsedDB.keys))
		for _, k := range parsedDB.keys {
			k := []byte(k)
			v, err := db.Get(k, nil)
			checkError(err)

			var g map[string]interface{}
			if err = gob.NewDecoder(bytes.NewReader(v)).Decode(&g); err != nil {
				// legacy gzipped messages
				if gzReader, gzErr := gzip.NewReader(bytes.NewReader(v)); gzErr == nil {
					err = gob.NewDecoder(gzReader).Decode(&g)
				}
				checkError(err)
			}

			if g["group"] != 619047717 {
				continue
			}
			if t, ok := g["time"].(float64); ok && t < 1638182816 {
				continue
			}
			if s, ok := g["message"].(string); ok && s[:5] != "[CQ:i" {
				continue
			}
			ret = append(ret, map[string]interface{}{
				"message":  g["message"],
				"group_id": g["group"],
				"user_id":  257195552,
				"self_id":  257195552,
				"sender": map[string]interface{}{
					"user_id": 257195552,
					"title":   "",
				},
			})
		}
		encoder := json.NewEncoder(jsonFile)
		encoder.SetIndent("", "\t")
		checkError(encoder.Encode(ret))

		fmt.Printf("%d messages written to %#v\n", len(ret), jsonPath)
	}
}

var (
	rootPath        string
	timezone        string
	quiet           bool = true
	cleanOutput     bool = false
	searchResult    []string
	parsedDatabases []ParsedDB
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("usage: dump-gocq <leveldb path> <json path>")
		os.Exit(1)
	}

	rootPath = os.Args[1]
	jsonPath := os.Args[2]
	if dbExists, err := fileExists(rootPath); !dbExists || err != nil {
		checkError(err)
		fmt.Printf(`The path "%s" doesn't exist`, rootPath)
		return
	}

	start := time.Now()

	searchForDBs()
	readDBs()
	dumpGoCQ(rootPath, jsonPath)

	elapsed := time.Now().Sub(start)
	fmt.Printf("Completed in %v\n", elapsed)
}
