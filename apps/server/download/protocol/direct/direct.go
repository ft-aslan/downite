package direct

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
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
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
	// part contexts for each download. we need to cancel them if download is cancelled
	partContextMap          map[int][]*contextWithCancel
	downloadStateChangeChan chan map[int]types.DownloadStatus
}
type contextWithCancel struct {
	ctx    *context.Context
	cancel context.CancelFunc
}

func CreateDownloadClient(config DownloadClientConfig, db *db.Database) (*Client, error) {
	return &Client{
		Config: &config,
		httpClient: &http.Client{
			Transport: http.DefaultTransport,
		},
		db: db,
	}, nil
}
func (client *Client) InitDownloads() error {
	client.downloads = make(map[int]*types.Download, 0)
	client.partContextMap = make(map[int][]*contextWithCancel)
	client.downloadStateChangeChan = make(chan map[int]types.DownloadStatus)

	go func() {
		for stateChanges := range client.downloadStateChangeChan {
			for id, stateChange := range stateChanges {
				partContexts, ok := client.partContextMap[id]
				if !ok {
					continue
				}
				for _, ctxWithCancel := range partContexts {
					ctxWithCancel.cancel()
				}

				delete(client.partContextMap, id)
				client.mutexForDownloads.Lock()
				download := client.downloads[id]
				download.Status = stateChange
				client.mutexForDownloads.Unlock()
			}
		}
	}()
	return nil
}
func (client *Client) PauseDownload(id int) error {
	client.downloadStateChangeChan <- map[int]types.DownloadStatus{
		id: types.DownloadStatusPaused,
	}

	client.mutexForDownloads.Lock()
	download := client.downloads[id]
	download.Status = types.DownloadStatusPaused
	client.mutexForDownloads.Unlock()

	return nil
}

func (client *Client) ResumeDownload(id int) error {
	client.downloadStateChangeChan <- map[int]types.DownloadStatus{
		id: types.DownloadStatusDownloading,
	}

	client.mutexForDownloads.Lock()
	download := client.downloads[id]
	download.Status = types.DownloadStatusPaused
	client.mutexForDownloads.Unlock()

	return nil
}
func (client *Client) GetDownloadMeta(url string) (*types.DownloadMeta, error) {
	req, err := http.NewRequest("HEAD", url, nil)

	if err != nil {
		return nil, fmt.Errorf("while creating request: %s", err)
	}

	res, err := client.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("while head request: %s", err)
	}
	//check if server accepts split downloads
	rangesHeader := res.Header.Get("Accept-Ranges")
	//total file size
	contentLengthHeader := res.Header.Get("Content-Length")
	//EXAMPLE HEADER = "attachment; filename=\"test.txt\""
	contentDispositionHeader := res.Header.Get("Content-Disposition")

	if contentLengthHeader == "" {
		return nil, fmt.Errorf("cannot find content length in headers")
	}

	var fileName string

	if contentDispositionHeader == "" {
		fileName = path.Base(url)
	} else {
		//if it has filename in its string
		fileName = getFileNameFromHeader(contentDispositionHeader)
	}

	if fileName == "" {
		return nil, fmt.Errorf("cannot find file name")
	}
	if contentLengthHeader == "" {
		return nil, fmt.Errorf("cannot find content length in headers")
	}
	contentLength, err := strconv.ParseUint(contentLengthHeader, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("cannot convert content length header to int : %s", err)
	}

	return &types.DownloadMeta{
		FileName:       fileName,
		TotalSize:      contentLength,
		Url:            url,
		FileType:       path.Ext(fileName),
		IsRangeAllowed: rangesHeader == "bytes",
	}, nil

}

func (client *Client) DownloadFromUrl(url string, partCount int, savePath string, startDownload bool, addTopOfQueue bool) (*types.Download, error) {
	//GET METAINFO
	metaInfo, err := client.GetDownloadMeta(url)
	if err != nil {
		return nil, err
	}

	var partLength uint64 = 0

	if metaInfo.IsRangeAllowed {
		partLength = uint64(math.Floor(float64(metaInfo.TotalSize) / float64(partCount)))
	} else {
		partLength = metaInfo.TotalSize
		partCount = 1
	}

	// if save path empty use default path
	if savePath == "" {
		savePath = client.Config.DownloadPath
	}

	download := &types.Download{
		CreatedAt:       time.Now(),
		Parts:           make([]*types.DownloadPart, partCount),
		Name:            metaInfo.FileName,
		SavePath:        savePath,
		PartCount:       partCount,
		PartLength:      partLength,
		Url:             url,
		TotalSize:       uint64(metaInfo.TotalSize),
		DownloadedBytes: 0,
		Progress:        0,
		Status:          types.DownloadStatusPaused,
	}

	//REGISTER DOWNLOAD to DB
	//from now on download has id from db
	client.RegisterDownload(download, addTopOfQueue)
	//ADD DOWNLOAD TO client
	client.AddDownload(download)
	//START SPLIT DOWNLOAD
	if startDownload {
		err := client.StartDownload(download.Id)
		if err != nil {
			return nil, err
		}
	}

	return download, nil
}
func (client *Client) RegisterDownload(download *types.Download, addTopOfQueue bool) error {
	for i := 0; i < download.PartCount; i++ {

		startByteIndex := uint64(i) * download.PartLength
		endByteIndex := uint64((uint64(i) + 1) * download.PartLength)

		if endByteIndex > download.TotalSize {
			endByteIndex = download.TotalSize
		}

		download.Parts[i] = &types.DownloadPart{
			CreatedAt:       time.Now(),
			PartIndex:       i + 1,
			StartByteIndex:  startByteIndex,
			EndByteIndex:    endByteIndex,
			PartLength:      download.PartLength,
			Status:          types.DownloadStatusPaused,
			DownloadId:      download.Id,
			DownloadedBytes: 0,
			Progress:        0,
		}
	}

	if addTopOfQueue {
		download.QueueNumber = 1
	} else {
		if len(client.downloads) == 0 {
			download.QueueNumber = 1
		} else {
			lastQueueNumber, err := client.db.GetLastQueueNumberOfDownloads()
			if err != nil {
				return err
			}
			download.QueueNumber = lastQueueNumber + 1
		}
	}

	id, err := client.db.InsertDownload(download, addTopOfQueue)
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
func (client *Client) RemoveDownload(id int) error {

	err := client.PauseDownload(id)
	if err != nil {
		return err
	}

	err = client.db.DeleteDownload(id)
	if err != nil {
		return err
	}
	err = client.db.DeleteDownloadParts(id)
	if err != nil {
		return err
	}

	client.mutexForDownloads.Lock()
	delete(client.downloads, id)
	client.mutexForDownloads.Unlock()

	err = client.updateDownloadQueueNumbers()
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) updateDownloadQueueNumbers() error {
	dbDownloads, err := client.db.GetDownloads()
	client.mutexForDownloads.Lock()
	defer client.mutexForDownloads.Unlock()
	if err != nil {
		return err
	}
	for _, dbDownload := range dbDownloads {
		client.downloads[dbDownload.Id].QueueNumber = dbDownload.QueueNumber
	}
	return nil
}

func (client *Client) DeleteDownload(id int) error {

	client.mutexForDownloads.Lock()
	savePath := client.downloads[id].SavePath
	fileName := client.downloads[id].Name
	client.mutexForDownloads.Unlock()

	err := client.RemoveDownload(id)
	if err != nil {
		return err
	}
	err = os.RemoveAll(filepath.Join(savePath, fileName))
	if err != nil {
		return err
	}
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
func (client *Client) GetDownloads() ([]*types.Download, error) {
	client.mutexForDownloads.Lock()
	defer client.mutexForDownloads.Unlock()

	if client.downloads == nil {
		return nil, fmt.Errorf("downloads are not initilized")
	}
	downloads := make([]*types.Download, 0, len(client.downloads))
	for _, download := range client.downloads {
		downloads = append(downloads, download)
	}
	return downloads, nil
}
func (client *Client) StartDownload(id int) error {
	download, err := client.GetDownload(id)
	if err != nil {
		return err
	}
	fmt.Printf("starting download : %s \n", filepath.Join(download.SavePath, download.Name))
	fileBuffer, err := os.OpenFile(filepath.Join(download.SavePath, download.Name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fileBuffer.Close()

	partProcessChan := make(chan *types.DownloadPart, download.PartCount)
	errorChan := make(chan error)

	var completedPartCount int = 0

	downloadPartContexts := make([]*contextWithCancel, 0, download.PartCount)

	for _, part := range download.Parts {
		fmt.Printf("requesting : start %d | end %d \n", part.StartByteIndex, part.EndByteIndex)

		ctx, cancel := context.WithCancel(context.Background())

		downloadPartContexts = append(downloadPartContexts, &contextWithCancel{
			ctx:    &ctx,
			cancel: cancel,
		})

		// we are creating new goroutine for each part
		go func() {
			filePartBufferReader, err := client.downloadFilePart(part, download.Url, ctx)
			if err != nil {
				errorChan <- err
				cancel()
				return
			}
			buffer := make([]byte, part.EndByteIndex+1-part.StartByteIndex)

			_, err = io.ReadFull(filePartBufferReader, buffer)
			if err != nil {
				errorChan <- err
				cancel()
				return
			}
			// defer filePartBufferReader.Close()
			partProgress := &types.DownloadPart{
				PartIndex:      part.PartIndex,
				StartByteIndex: part.StartByteIndex,
				EndByteIndex:   part.EndByteIndex,
				PartLength:     part.PartLength,
				Status:         types.DownloadStatusCompleted,
				DownloadId:     part.DownloadId,
				Progress:       100.0,
				Buffer:         buffer,
			}
			partProcessChan <- partProgress
		}()

	}
	download.Status = types.DownloadStatusDownloading
	download.StartedAt = time.Now()
	err = client.db.UpdateDownload(download)
	if err != nil {
		return err
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
			_, err := fileBuffer.Seek(int64(start), io.SeekStart)
			if err != nil {
				return err
			}
			_, err = fileBuffer.Write(partProcess.Buffer)
			if err != nil {
				return err
			}

			fmt.Printf("copied to start index id : %d | end index id : %d | part id : %d \n", partProcess.StartByteIndex, partProcess.EndByteIndex, partProcess.PartIndex)
			completedPartCount += 1
		}
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

	if res.StatusCode == http.StatusPartialContent || res.StatusCode == http.StatusOK {
		return filePartBufferReader, nil
	}

	return nil, fmt.Errorf("unexpected status code: %s", res.Status)

}
