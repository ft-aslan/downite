package http

import (
	"context"
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
	downloads         map[int]*types.Download
	httpClient        *http.Client
	Config            *DownloadClientConfig
	db                *db.Database
}
type contextWithCancel struct {
	ctx    *context.Context
	cancel *context.CancelFunc
}

func CreateDownloadClient(config DownloadClientConfig) (*Client, error) {
	return &Client{
		Config: &config,
	}, nil
}
func (client *Client) InitDownloads() error {
	client.downloads = make(map[int]*types.Download, 0)
	return nil
}

func (client *Client) DownloadFromUrl(url string, partCount int, savePath string, startDownload bool) error {

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
		Parts:           make([]*types.DownloadPart, partCount),
		Name:            fileName,
		Path:            savePath,
		PartCount:       partCount,
		PartLength:      partLength,
		Url:             url,
		TotalSize:       uint64(contentLength),
		DownloadedBytes: 0,
		Status:          types.DownloadStatusPaused,
	}

	//REGISTER DOWNLOAD
	//from now on download has id from db
	client.RegisterDownload(download)

	//START SPLIT DOWNLOAD
	if startDownload {
		client.AddDownload(download)
	}
	return nil
}
func (client *Client) RegisterDownload(download *types.Download) error {
	for i := 0; i < download.PartCount; i++ {

		startByteIndex := uint64(i) * download.PartLength
		endByteIndex := uint64((uint64(i) + 1) * download.PartLength)

		if endByteIndex > download.TotalSize {
			endByteIndex = download.TotalSize
		}

		download.Parts[i] = &types.DownloadPart{
			PartIndex:      i + 1,
			StartByteIndex: startByteIndex,
			EndByteIndex:   endByteIndex,
			PartLength:     download.PartLength,
			Status:         types.DownloadStatusPaused,
			DownloadId:     download.Id,
		}
	}

	id, err := client.db.InsertDownload(download)
	if err != nil {
		return err
	}
	download.Id = id

	err = client.db.InsertDownloadParts(download.Parts)
	if err != nil {
		return err
	}

	return nil
}
func (client *Client) AddDownload(download *types.Download) {
	client.mutexForDownloads.Lock()
	defer client.mutexForDownloads.Unlock()

	client.downloads[download.Id] = download
}
func (client *Client) RemoveDownload(download *types.Download) error {
	client.mutexForDownloads.Lock()
	defer client.mutexForDownloads.Unlock()

	delete(client.downloads, download.Id)
	return nil
}
func (client *Client) GetDownload(id int) (*types.Download, error) {
	client.mutexForDownloads.Lock()
	defer client.mutexForDownloads.Unlock()

	download, ok := client.downloads[id]
	if !ok {
		return nil, fmt.Errorf("download not found")
	}
	return download, nil
}
func (client *Client) StartDownload(id int) error {
	download, err := client.GetDownload(id)
	if err != nil {
		return err
	}

	fileBuffer := make([]byte, int(download.TotalSize))
	var partStartIndex uint64 = 0

	partProcessChan := make(chan *types.DownloadPart, download.PartCount)
	errorChan := make(chan error)

	var completedPartCount int = 0

	downloadPartContexts := make([]*contextWithCancel, 0, download.PartCount)

	for _, part := range download.Parts {
		fmt.Printf("requesting : start %d | end %d \n", part.StartByteIndex, part.EndByteIndex)

		ctx, cancel := context.WithCancel(context.Background())

		downloadPartContexts[part.PartIndex-1] = &contextWithCancel{
			ctx:    &ctx,
			cancel: &cancel,
		}

		// we are creating new goroutine for each part
		go func() {
			filePartBufferReader, err := client.downloadFilePart(part, download.Url, ctx)
			if err != nil {
				errorChan <- err
				return
			}
			buffer := make([]byte, part.EndByteIndex+1-partStartIndex)

			_, err = io.ReadFull(filePartBufferReader, buffer)
			if err != nil {
				errorChan <- err
				return
			}
			// defer filePartBufferReader.Close()
			partProgress := &types.DownloadPart{
				PartIndex:      part.PartIndex,
				StartByteIndex: partStartIndex,
				EndByteIndex:   part.EndByteIndex,
				Buffer:         buffer,
			}
			partProcessChan <- partProgress
		}()

		partStartIndex += download.PartLength + 1

	}

	for completedPartCount != download.PartCount {
		select {
		case err := <-errorChan:
			return fmt.Errorf("while downloading file parts : %s", err)
		case partProcess := <-partProcessChan:
			start := partProcess.StartByteIndex
			end := partProcess.EndByteIndex + 1
			if end > download.TotalSize {
				end = download.TotalSize
			}
			copy(fileBuffer[start:end], partProcess.Buffer)

			fmt.Printf("copied to start index id : %d | end index id : %d | part id : %d \n", partProcess.StartByteIndex, partProcess.EndByteIndex, partProcess.PartIndex)
			completedPartCount += 1
		}
	}

	outFile, err := os.Create(download.Path + download.Name)
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

func (client *Client) downloadFilePart(downloadPart *types.DownloadPart, url string, ctx context.Context) (io.Reader, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("while creating request: %s", err)
	}
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", downloadPart.StartByteIndex, downloadPart.EndByteIndex))

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("while download : %s", err)
	}
	defer res.Body.Close()

	filePartBufferReader := io.TeeReader(res.Body, downloadPart)
	if err != nil {
		return nil, fmt.Errorf("while reading response : %s", err)
	}

	if res.StatusCode == http.StatusPartialContent || res.StatusCode == http.StatusOK {
		return filePartBufferReader, nil
	}

	return nil, fmt.Errorf("unexpected status code: %s", res.Status)

}
