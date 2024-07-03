package http

import (
	"downite/db"
	"downite/types"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

type DownloadClientConfig struct {
	DownloadPath string
	PartCount    int
}

// HTTP DOWNLOAD CLIENT
type Client struct {
	// config *ClientConfig
	// logger log.Logger

	// defaultStorage *storage.Client
	onClose []func()

	mutexForDownloads sync.Mutex
	downloads         []*types.Download
	httpClient        *http.Client
	Config            *DownloadClientConfig
	db                *db.Database
}

func CreateDownloadClient(config DownloadClientConfig) (*Client, error) {
	return &Client{
		Config: &config,
	}, nil
}
func (client *Client) InitDownloads() error {
	client.downloads = make([]*types.Download, 0)
	return nil
}

func (client *Client) DownloadFromUrl(url string, partCount int, fileDownloadPath string) error {

	//check if server accepts split downloads
	req, err := http.NewRequest("HEAD", url, nil)

	if err != nil {
		return fmt.Errorf("while creating request: %s", err)
	}

	res, err := client.httpClient.Do(req)

	if err != nil {
		return fmt.Errorf("while head request: %s", err)
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
	if contentLengthHeader == "" {
		return fmt.Errorf("cannot find content length in headers")
	}
	contentLength, err := strconv.ParseUint(contentLengthHeader, 10, 64)
	if err != nil {
		return fmt.Errorf("cannot convert content length header to int : %s", err)
	}

	var partLength uint64 = 0
	if rangesHeader != "" {
		partLength = uint64(math.Floor(float64(contentLength) / float64(partCount)))
	} else {
		partLength = contentLength
		partCount = 1
	}
	download := &types.Download{
		PartProgress:    make([]*types.DownloadPart, partCount),
		Name:            fileName,
		Path:            fileDownloadPath,
		PartCount:       partCount,
		PartLength:      partLength,
		Url:             url,
		TotalSize:       uint64(contentLength),
		DownloadedBytes: 0,
		Status:          types.DownloadStatusPaused,
	}

	fileBuffer := make([]byte, int(contentLength))

	if rangesHeader != "" {
		download.Status = types.DownloadStatusDownloading
		for i := 0; i < partCount; i++ {
			download.PartProgress[i] = &types.DownloadPart{
				PartIndex:      i + 1,
				StartByteIndex: uint64(uint64(i) * partLength),
				EndByteIndex:   uint64((uint64(i) + 1) * partLength),
				Status:         types.DownloadStatusPaused,
			}
		}

		id, err := client.db.InsertDownload(download)
		if err != nil {
			return err
		}
		download.Id = id

		client.mutexForDownloads.Lock()
		client.downloads = append(client.downloads, download)
		client.mutexForDownloads.Unlock()

		var currentStartByteIndex uint64 = 0

		partProcessChan := make(chan types.DownloadPart)
		errorChan := make(chan error)

		var completedPartCount int = 0

		for partIndex := 1; partIndex <= partCount; partIndex++ {
			endByteIndex := currentStartByteIndex + partLength

			if endByteIndex > contentLength {
				endByteIndex = contentLength
			}
			fmt.Printf("requesting : start %f | end %f \n", currentStartByteIndex, endByteIndex)

			go func(partStartIndex uint64, partEndIndex uint64) {
				filePartBuffer, err := client.downloadFilePart(partStartIndex, partEndIndex, url)
				if err != nil {
					errorChan <- err
					return
				}
				partProcess := types.DownloadPart{PartIndex: partIndex, Buffer: filePartBuffer, StartByteIndex: partStartIndex, EndByteIndex: partEndIndex}
				partProcessChan <- partProcess
			}(currentStartByteIndex, endByteIndex)

			currentStartByteIndex += partLength + 1

		}

		for completedPartCount != partCount {
			select {
			case err := <-errorChan:
				return fmt.Errorf("while downloading file parts : %s", err)
			case partProcess := <-partProcessChan:
				start := partProcess.StartByteIndex
				end := partProcess.EndByteIndex + 1
				if end > contentLength {
					end = contentLength
				}
				copy(fileBuffer[start:end], partProcess.Buffer)

				fmt.Printf("copied to start index id : %d | end index id : %d | part id : %d \n", partProcess.StartByteIndex, partProcess.EndByteIndex, partProcess.PartIndex)
				completedPartCount += 1
			}
		}
	} else {
		download.Status = types.DownloadStatusDownloading
		download.PartProgress[0] = &types.DownloadPart{
			PartIndex:      1,
			StartByteIndex: 0,
			EndByteIndex:   contentLength,
			Status:         types.DownloadStatusPaused,
		}
		// download.Id = id

		client.mutexForDownloads.Lock()
		client.downloads = append(client.downloads, download)
		client.mutexForDownloads.Unlock()

		go func() {
			fileBuffer = make([]byte, int(contentLength))
			res, err := client.httpClient.Get(url)
			if err != nil {
				fmt.Printf("error while downloading : %s", err)
				client.mutexForDownloads.Lock()

				client.mutexForDownloads.Unlock()

				return
			}

			defer res.Body.Close()
			_, err = io.ReadFull(res.Body, fileBuffer)
			if err != nil {
				fmt.Printf("error while reading download : %s", err)
				return
			}
		}()
	}

	outFile, err := os.Create(fileDownloadPath + fileName)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = outFile.Write(fileBuffer)
	if err != nil {
		return err
	}

	return nil
}
func (client *Client) RegisterDownload(download *types.Download) error {
	_, err := client.db.InsertDownload(download)
	if err != nil {
		return err
	}
	return nil
}
func (client *Client) AddDownload(download *types.Download) error {
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

func (client *Client) downloadFilePart(startByteIndex uint64, endByteIndex uint64, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("while creating request: %s", err)
	}
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", startByteIndex, endByteIndex))

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("while download : %s", err)
	}
	defer res.Body.Close()

	filePartBuffer, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("while reading response : %s", err)
	}

	if res.StatusCode == http.StatusPartialContent || res.StatusCode == http.StatusOK {
		return filePartBuffer, nil
	}

	return nil, fmt.Errorf("unexpected status code: %s", res.Status)

}
