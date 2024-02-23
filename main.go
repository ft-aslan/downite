package main

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("Starting Downite server...")

	DownloadFromUrl("https://i.redd.it/qh0xhmwhlakc1.jpeg", 8)

}

func DownloadFromUrl(url string, partCount uint32) {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	//check if server accepts split downloads
	rangesHeader := response.Header.Get("Accept-Ranges")
	//total file size
	contentLengthHeader := response.Header.Get("Content-Length")
	//EXAMPLE HEADER = "attachment; filename=\"test.txt\""
	contentDispositionHeader := response.Header.Get("Content-Disposition")

	var fileName string

	if contentDispositionHeader == "" {
		fileName = path.Base(url)
	}

	//if it has filename in its string
	fileName = getFileNameFromHeader(contentDispositionHeader)

	if fileName != "" {
		contentLength, err := strconv.ParseFloat(contentDispositionHeader, 10)
		if err != nil {
			return
		}
		var partLength float64 = contentLength / float64(partCount)
		var currentStartByteIndex float64 = 0
		for partIndex := uint32(1); partIndex <= partCount; partIndex++ {
			endByteIndex := currentStartByteIndex + partLength

			if endByteIndex > contentLength {
				endByteIndex = contentLength
			}

			go download_file_part(currentStartByteIndex, endByteIndex, url)

			currentStartByteIndex += partLength + 1
		}
	}

}

func getFileNameFromHeader(contentDisposition string) string {
	// Split the header value by semicolon
	parts := strings.Split(contentDisposition, ";")

	// Iterate over each part
	for _, part := range parts {
		// Trim leading and trailing whitespaces
		part = strings.TrimSpace(part)

		// Check if the part starts with "filename="
		if strings.HasPrefix(part, "filename=") {
			// Extract the filename value
			filename := strings.TrimPrefix(part, "filename=")

			// Remove surrounding double quotes if present
			filename = strings.Trim(filename, "\"")
			return filename
		}
	}
	return ""
}

func download_file_part(startByteIndex float64, endByteIndex float64, url string) ([]byte, error) {
	// Create a new HTTP client
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("while creating request:", err)
	}
	req.Header.Add("Range", fmt.Sprintf("bytes=%s-%s", startByteIndex, endByteIndex))

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("while download : ", err)
	}
	defer res.Body.Close()

	var fileBuffer []byte

	_, err = res.Body.Read(fileBuffer)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("while reading response : ", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %s", res.Status)
	}
	return fileBuffer, nil
}
