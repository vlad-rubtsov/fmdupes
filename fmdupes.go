package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	id3 "github.com/mikkyang/id3-go"
	tagg "github.com/wtolson/go-taglib"
	gcfg "gopkg.in/gcfg.v1"
)

type Config struct {
	InputDir string
}

type configFile struct {
	Fmdupes Config
}

type Mp3Song struct {
	Filename string
	Artist   string
	Title    string
	Genre    string
	Path     string
	Size     int64
}

type SongType struct {
	Artist string
	Title  string
}

var mp3List []Mp3Song

var xxx map[SongType][]string

func GetMp3Data_Old(filename string) (Mp3Song, error) {
	mp3File, err := id3.Open(filename)
	if err != nil {
		fmt.Println("Open: unable to open file: ", err)
		return Mp3Song{}, err
	}
	defer mp3File.Close()

	// if mp3File.Title() == "Guachipupa" {
	// 	//fmt.Println(mp3File)
	// 	fmt.Println("padding: ", mp3File.Padding())
	// 	fmt.Println("size: ", mp3File.Size())
	// 	fmt.Println("version: ", mp3File.Version())
	// }

	fmt.Printf("f: %s, artist: %s, title: %s, v: %v\n",
		filename, mp3File.Artist(), mp3File.Title(), mp3File.Version())
	// fmt.Printf("file: %s, version: %v\n",
	// 	filename, mp3File.Version())

	return Mp3Song{
		Filename: filename,
		Artist:   mp3File.Artist(),
		Title:    mp3File.Title(),
		Genre:    mp3File.Genre(),
		Size:     0,
	}, nil
}

func GetMp3Data(filename string) (Mp3Song, error) {
	mp3File, err := tagg.Read(filename)
	if err != nil {
		fmt.Println("Open: unable to open file: ", err)
		return Mp3Song{}, err
	}
	defer mp3File.Close()

	// if mp3File.Title() == "Guachipupa" {
	// 	//fmt.Println(mp3File)
	// 	fmt.Println("padding: ", mp3File.Padding())
	// 	fmt.Println("size: ", mp3File.Size())
	// 	fmt.Println("version: ", mp3File.Version())
	// }

	//fmt.Printf("f: %s, artist: %s, title: %s\n",
	//	filename, mp3File.Artist(), mp3File.Title())
	// fmt.Printf("file: %s, version: %v\n",
	// 	filename, mp3File.Version())

	return Mp3Song{
		Filename: filename,
		Artist:   mp3File.Artist(),
		Title:    mp3File.Title(),
		Genre:    mp3File.Genre(),
		Size:     0,
	}, nil
}

func DirWalk(path string, fi os.FileInfo, err error) error {
	//fmt.Println("walk path: ", path)

	if fi.IsDir() {
		//fmt.Println("Search in ", path)
		//fmt.Printf("Process %d files\n", len(mp3List))
		return nil
	}

	if filepath.Ext(path) != ".mp3" {
		return nil
	}

	data, err := GetMp3Data(path)
	if err == nil {
		data.Size = fi.Size()
		data.Path = path
		mp3List = append(mp3List, data)

		// Debug
		// if data.Title == "Guachipupa" {
		// 	fmt.Println(data)
		// }
		// TODO: save data for know time, size and bitrate
		xxx[SongType{Artist: data.Artist, Title: data.Title}] =
			append(xxx[SongType{Artist: data.Artist, Title: data.Title}], path)
	}
	return nil
}

func loadConfig(cfgFile string) Config {
	var cfg configFile
	err := gcfg.ReadFileInto(&cfg, cfgFile)
	if err != nil {
		fmt.Errorf("Error reading config file: %s", err)
	}
	return cfg.Fmdupes
}

func main() {
	xxx = make(map[SongType][]string)

	// for future
	// delete := flag.Bool("d", false, "prompt user for files to preserve and delete all")
	// size := flag.Bool("S", false, "show size of duplicate files")

	// flag.Usage = func() {
	// 	fmt.Fprintf(os.Stderr, "Usage: %s OPTIONS dirs\n", os.Args[0])
	// 	flag.PrintDefaults()
	// 	os.Exit(2)
	// }
	// flag.Parse()

	cfg := loadConfig("fmdupes.conf")

	// Walk in dirs
	//
	dirs := strings.Split(cfg.InputDir, ",")
	//fmt.Println("dirs: ", dirs)

	for _, dir := range dirs {
		err := filepath.Walk(dir, DirWalk)
		if err != nil {
			fmt.Errorf("DirWalk error: %v", err)
		}
	}

	fmt.Println("Result:")
	fmt.Println("-------")
	fmt.Printf("Read %d songs from directory %s\n", len(mp3List), cfg.InputDir)

	//fmt.Println("xxx: ", xxx)

	// TODO: sort by count = len(val)
	for key, val := range xxx {
		count := len(val)
		if count > 2 {
			fmt.Printf("%v, count: %d\n", key, count)
			//fmt.Println(val)
			for id, path := range val {
				fmt.Printf("[%d] %s\n", id, path)
			}
			fmt.Printf("Set [0-%d, all] to delete:\n", len(val))
			fmt.Println()
		}
	}

}
