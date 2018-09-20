package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/jmoiron/jsonq"
	"gopkg.in/cheggaaa/pb.v1"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	apiKeyFile = "../.tinify_api_key"
	apiURL     = "https://api.tinify.com/shrink"
	filesDir   = "./files"
	doneDir    = "compressed"
	apiLimit   = 500
	count      int
	doneCount  int
	fileCount  int
	origSize   int
	files      []os.FileInfo
)

func main() {
	if _, err := os.Stat(apiKeyFile); os.IsNotExist(err) {
		logError()
		fmt.Print("API Key not found. Please save API key at ")
		c := color.New(color.Bold)
		c.Println(apiKeyFile)
		os.Exit(1)
	}
	if _, err := os.Stat(doneDir); os.IsNotExist(err) {
		os.Mkdir(doneDir, 1)
	}
	logInfo()
	fmt.Println("Starting compression....")
	f, err := os.Open("./files")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fis, err := f.Readdir(-1)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	f.Close()
	for _, fi := range fis {
		if (filepath.Ext(fi.Name()) == ".JPG") ||
			(filepath.Ext(fi.Name()) == ".jpg") ||
			(filepath.Ext(fi.Name()) == ".JPEG") ||
			(filepath.Ext(fi.Name()) == ".jpeg") ||
			(filepath.Ext(fi.Name()) == ".PNG") ||
			(filepath.Ext(fi.Name()) == ".png") {
			files = append(files, fi)
		}
	}
	fileCount = len(files)
	if fileCount == 0 {
		logError()
		fmt.Println("No pictures found!")
		os.Exit(1)
	}
	for _, file := range files {
		processImage(file.Name())
		time.Sleep(100 * time.Millisecond)
	}
	logInfo()
	fmt.Println(doneCount, "files compressed!")
}

func logAny() {
	t := time.Now()
	c := color.New(color.FgCyan)
	c.Printf(" [%02d:%02d:%02d] ", t.Hour(), t.Minute(), t.Second())
}

func logInfo() {
	logAny()
	c := color.New(color.FgGreen, color.Bold)
	c.Print("[INFO] ")
}

func logWarning() {
	logAny()
	c := color.New(color.FgYellow, color.Bold)
	c.Print("[WARNING] ")
}

func logError() {
	logAny()
	c := color.New(color.FgRed, color.Bold)
	c.Print("[ERROR] ")
}

func exists(file string) bool {
	_, err := os.Stat(file)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func readAPIKey(file string) string {
	f, err := os.Open(file)
	if err != nil {
		f.Close()
		logError()
		fmt.Println(err)
		os.Exit(1)
	}
	scanner := bufio.NewScanner(f)
	var apiKey string
	for scanner.Scan() {
		apiKey = scanner.Text()
		break
	}
	return apiKey

}

func uploadImage(url string, filename string, filesize int) (string, []byte) {
	apiKey := readAPIKey(apiKeyFile)
	client := &http.Client{}
	f, _ := os.Open(filename)
	bar := pb.New(filesize).SetUnits(pb.U_BYTES)
	bar.Start()
	reader := bar.NewProxyReader(f)
	req, _ := http.NewRequest("POST", url, reader)
	req.SetBasicAuth("api", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		f.Close()
		logWarning()
		fmt.Println(err)
	}
	bar.Finish()
	f.Close()
	apiCount, err := strconv.Atoi(resp.Header.Get("Compression-Count"))
	if err != nil {
		logWarning()
		fmt.Println(err)
	}
	logInfo()
	fmt.Println("API Requests:", strconv.Itoa(apiCount))
	if apiCount > apiLimit {
		logError()
		fmt.Println("API Limit Reached!")
		os.Exit(1)
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logWarning()
		fmt.Println(err)
	}
	return resp.Status, content
}

func downloadImage(url string, filename string, filesize int) string {
	apiKey := readAPIKey(apiKeyFile)
	preserve := "{ \"preserve\": [\"location\", \"creation\"] }"
	f := strings.NewReader(preserve)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, f)
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("api", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		logWarning()
		fmt.Println(err)
	}
	data := resp.Body
	newFilesize, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err == nil {
		filesize = newFilesize
	}
	bar := pb.New(filesize).SetUnits(pb.U_BYTES)
	bar.Start()
	reader := bar.NewProxyReader(data)
	writer, err := os.Create(filename)
	if err != nil {
		logWarning()
		fmt.Println(err)
	}
	writer.Write(streamToByte(reader))
	bar.Finish()
	return resp.Status
}

func streamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}

func processImage(file string) {
	count++
	outputFilepath := doneDir + "/" + file
	outputSize := 0
	f, err := os.Stat(outputFilepath)
	if err == nil {
		outputSize = int(f.Size())
	}
	logInfo()
	fmt.Print("Processing ")
	c := color.New(color.Bold)
	c.Print(file)
	fmt.Print(" .... (", count, " of ", fileCount, ")\n")
	j := 0
	if exists(outputFilepath) && outputSize != 0 {
		logInfo()
		fmt.Println("Image has already been compressed. Skipping....")
	}
	for !exists(outputFilepath) || outputSize == 0 {
		doneCount++
		j++
		if j > 10 {
			doneCount--
			logError()
			fmt.Println("Too many re-tries! Skipping....")
			break
		}
		if j > 1 {
			doneCount--
			logInfo()
			fmt.Println("Re-try", j-1)
		}
		filepath := filesDir + "/" + file
		logInfo()
		fmt.Println("Compressing....")
		f, _ := os.Stat(filepath)
		origSize = int(f.Size())
		status, content := uploadImage(apiURL, filepath, origSize)
		if status != "201 Created" {
			logWarning()
			fmt.Print("Invalid server response: ", status, ". Retrying....\n")
			continue
		}
		data := map[string]interface{}{}
		dec := json.NewDecoder(strings.NewReader(string(content)))
		dec.Decode(&data)
		jq := jsonq.NewQuery(data)
		newSize, err := jq.Int("output", "size")
		if err != nil {
			logWarning()
			fmt.Println(err)
		}
		newURL, err := jq.String("output", "url")
		if err != nil {
			logWarning()
			fmt.Println(err)
		}
		status = downloadImage(newURL, outputFilepath, newSize)
		if status != "200 OK" {
			logWarning()
			fmt.Print("Invalid server response: ", status, ". Retrying....\n")
			continue
		}
		f, err = os.Stat(outputFilepath)
		if err == nil {
			outputSize = int(f.Size())
		}
	}
}
