package main

// TODO: sort by count = len(val)
// TODO: save data for know time, size and bitrate
// TODO: process flags
// TODO: fix error when press 'q'
// TODO: move delete file code to function

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cheggaaa/pb"
	tag "github.com/wtolson/go-taglib"
	gcfg "gopkg.in/gcfg.v1"
)

// Types
type Config struct {
	InputDir        string
	CountDuplicates int `gcfg:"count"`
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
var bar *pb.ProgressBar

func GetMp3Data(filename string) (Mp3Song, error) {
	mp3File, err := tag.Read(filename)
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

		xxx[SongType{Artist: data.Artist, Title: data.Title}] =
			append(xxx[SongType{Artist: data.Artist, Title: data.Title}], path)
		bar.Increment()
	}
	return nil
}

func loadConfig(cfgFile string) Config {
	var cfg configFile
	err := gcfg.ReadFileInto(&cfg, cfgFile)
	if err != nil {
		fmt.Errorf("Error reading config file: %s", err)
	}
	if cfg.Fmdupes.CountDuplicates == 0 {
		cfg.Fmdupes.CountDuplicates = 2
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

	// count all music files
	for _, dir := range dirs {
		err := filepath.Walk(dir, CountDirWalk)
		if err != nil {
			fmt.Errorf("DirWalk error: %v", err)
		}
	}
	fmt.Printf("Found %d music files in directory %s\n", cntMP3Files, cfg.InputDir)

	bar = pb.New(cntMP3Files)
	bar.Start()

	for _, dir := range dirs {
		err := filepath.Walk(dir, DirWalk)
		if err != nil {
			fmt.Errorf("DirWalk error: %v", err)
		}
	}
	bar.FinishPrint("Result:")
	fmt.Println("-------")

	cntDuplicates := 0
	for _, val := range xxx {
		if len(val) > 1 {
			cntDuplicates++
		}
	}
	fmt.Printf("Found %d duplicates\n", cntDuplicates)

	exit := false
	cntCons := 0
	var deleteAllSize int64
	var cntDeletedFiles int
	for key, val := range xxx {
		if exit {
			break
		}
		count := len(val)
		if count > 1 {
			cntCons++
		}
		if count >= cfg.CountDuplicates {
			fmt.Printf("%d. %v, %d duplicates\n", cntCons, key, count)
			for id, path := range val {
				i := id + 1
				fmt.Printf("[%d] %s\n", i, path)
			}
			fmt.Printf("Set [1-%d, all] to delete: ", len(val))
			in := ""
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
				fmt.Println("Need to test")
				for _, filepath := range val {
					f, _ := os.Open(filepath)
					fi, _ := f.Stat()
					size := fi.Size()
					f.Close()

					var kilobytes float64
					kilobytes = (float64)(size / 1024)
					var megabytes float64
					megabytes = (float64)(kilobytes / 1024)

					fmt.Printf("Delete %s, size: %.3f MB\n", filepath, megabytes)
					err := os.Remove(filepath)
					if err != nil {
						fmt.Printf("Error delete file: %s\n", err)
					} else {
						deleteAllSize += size
						cntDeletedFiles++
					}
				}
				break
			default:
				ids := strings.Split(in, ",")
				for _, is := range ids {
					ii, err := strconv.Atoi(is)
					if err == nil && ii <= len(val) {
						filepath := val[ii-1]
						f, _ := os.Open(filepath)
						fi, _ := f.Stat()
						size := fi.Size()
						f.Close()

						var kilobytes float64
						kilobytes = (float64)(size / 1024)
						var megabytes float64
						megabytes = (float64)(kilobytes / 1024)

						fmt.Printf("Delete %s, size: %.3f MB\n", filepath, megabytes)
						err := os.Remove(filepath)
						if err != nil {
							fmt.Printf("Error delete file: %s\n", err)
						} else {
							deleteAllSize += size
							cntDeletedFiles++
						}
					}
				}
			}
			fmt.Println()
		} // if
	} // for
	fmt.Printf("Delete %d files, all size: %.3f MB\n",
		cntDeletedFiles,
		((float64)(deleteAllSize) / 1024 / 1024))
}
