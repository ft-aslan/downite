package directdownload

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

type partProcess struct {
	PartId         uint32
	startByteIndex uint32
	endByteIndex   uint32
	Buffer         []byte
}

func DownloadFromUrl(url string, partCount uint32, fileDownloadPath string) error {
	// Create a new HTTP client
	client := &http.Client{}

	req, err := http.NewRequest("HEAD", url, nil)

	if err != nil {
		return fmt.Errorf("while creating request:", err)
	}

	res, err := client.Do(req)

	if err != nil {
		return fmt.Errorf("while head request:", err)
	}
	//check if server accepts split downloads
	rangesHeader := res.Header.Get("Accept-Ranges")
	//total file size
	contentLengthHeader := res.Header.Get("Content-Length")
	//EXAMPLE HEADER = "attachment; filename=\"test.txt\""
	contentDispositionHeader := res.Header.Get("Content-Disposition")

	if contentLengthHeader == "" {
		return fmt.Errorf("cannot find content length in headers")
	}

	var fileName string

	if contentDispositionHeader == "" {
		fileName = path.Base(url)
	} else {
		//if it has filename in its string
		fileName = getFileNameFromHeader(contentDispositionHeader)
	}

	if fileName == "" {
		return fmt.Errorf("cannot find file name")
	}

	if rangesHeader != "" && contentLengthHeader != "" {
		contentLength, err := strconv.ParseFloat(contentLengthHeader, 10)
		if err != nil {
			return fmt.Errorf("cannot convert content length header to float : ", err)
		}
		var partLength float64 = contentLength / float64(partCount)
		var currentStartByteIndex float64 = 0

		fileBuffer := make([]byte, int(contentLength))

		partProcessChan := make(chan partProcess)
		errorChan := make(chan error)

		var completedPartCount uint32 = 0

		for partIndex := uint32(1); partIndex <= partCount; partIndex++ {
			endByteIndex := math.Floor(currentStartByteIndex + partLength)

			if endByteIndex > contentLength {
				endByteIndex = contentLength
			}
			fmt.Printf("requesting : start %f | end %f \n", currentStartByteIndex, endByteIndex)

			go func(partStartIndex uint32, partEndIndex uint32) {
				filePartBuffer, err := downloadFilePart(partStartIndex, partEndIndex, url)
				if err != nil {
					errorChan <- err
					return
				}
				partProcess := partProcess{PartId: partIndex, Buffer: filePartBuffer, startByteIndex: partStartIndex, endByteIndex: partEndIndex}
				partProcessChan <- partProcess
			}(uint32(currentStartByteIndex), uint32(endByteIndex))

			currentStartByteIndex += math.Floor(partLength + 1)

		}

		for completedPartCount != partCount {
			select {
			case err := <-errorChan:
				return fmt.Errorf("while downloading file parts : ", err)
			case partProcess := <-partProcessChan:
				start := partProcess.startByteIndex
				end := partProcess.endByteIndex + 1
				if end > uint32(contentLength) {
					end = uint32(contentLength)
				}
				copy(fileBuffer[start:end], partProcess.Buffer)

				fmt.Printf("copied to start index id : %d | end index id : %d | part id : %d \n", partProcess.startByteIndex, partProcess.endByteIndex, partProcess.PartId)
				completedPartCount += 1
			}
		}
		outFile, err := os.Create(fileDownloadPath + fileName)
		if err != nil {
			return err
		}
		defer outFile.Close()

		_, err = outFile.Write(fileBuffer)

	}

	return nil
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

func downloadFilePart(startByteIndex uint32, endByteIndex uint32, url string) ([]byte, error) {
	// Create a new HTTP client
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("while creating request:", err)
	}
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", startByteIndex, endByteIndex))

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("while download : ", err)
	}
	defer res.Body.Close()

	filePartBuffer, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("while reading response : ", err)
	}

	if res.StatusCode == http.StatusPartialContent || res.StatusCode == http.StatusOK {
		return filePartBuffer, nil
	}

	return nil, fmt.Errorf("unexpected status code: %s", res.Status)

}
