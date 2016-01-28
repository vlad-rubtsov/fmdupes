package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	tagg "github.com/wtolson/go-taglib"
	gcfg "gopkg.in/gcfg.v1"
)

// Types
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

// Global
var mp3List []Mp3Song
var xxx map[SongType][]string
var cntMP3Files int

func GetMp3Data(filename string) (Mp3Song, error) {
	mp3File, err := tagg.Read(filename)
	if err != nil {
		fmt.Println("Open: unable to open file: ", err)
		return Mp3Song{}, err
	}
	defer mp3File.Close()

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

func CountDirWalk(path string, fi os.FileInfo, err error) error {
	if fi.IsDir() {
		return nil
	}
	if filepath.Ext(path) != ".mp3" {
		return nil
	}
	cntMP3Files++
	return nil
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

	// count all music files
	for _, dir := range dirs {
		err := filepath.Walk(dir, CountDirWalk)
		if err != nil {
			fmt.Errorf("DirWalk error: %v", err)
		}
	}
	fmt.Printf("Found %d music files\n", cntMP3Files)

	for _, dir := range dirs {
		//files, _ := ioutil.ReadDir(dir)
		//fmt.Println("count of dirs and files: ", len(files))

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
	fmt.Printf("Found %d consequences\n", len(xxx))
	in := ""
	exit := false
	for key, val := range xxx {
		if exit {
			break
		}
		count := len(val)
		if count > 2 {
			fmt.Printf("%v, %d duplicates\n", key, count)
			//fmt.Println(val)
			for id, path := range val {
				i := id + 1
				fmt.Printf("[%d] %s\n", i, path)
			}
			fmt.Printf("Set [1-%d] to delete: ", len(val))
			fmt.Scanln(&in)
			switch in {
			case "0":
				break
			case "":
				break
			case "q":
				exit = true
				break
			case "all":
				// TODO: delete all files
				fmt.Println("Not realized yet")
				break
			default:
				ids := strings.Split(in, ",")
				for _, is := range ids {
					ii, _ := strconv.Atoi(is)
					filepath := val[ii-1]
					f, _ := os.Open(filepath)
					fi, _ := f.Stat()
					size := fi.Size()
					f.Close()

					var kilobytes float64
					kilobytes = (float64)(size / 1024)

					var megabytes float64
					megabytes = (float64)(kilobytes / 1024) // cast to type float64

					fmt.Printf("Delete %s, size: %.3f MB\n", filepath, megabytes)
					err := os.Remove(filepath)
					if err != nil {
						fmt.Printf("Error delete file: %s\n", err)
					}
				}
			}
			fmt.Println()
		}
	}
}

// TODO: show count of duplicate
// TODO: show all delete size