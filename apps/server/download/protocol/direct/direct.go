package direct

import (
	"context"
	"database/sql"
	"downite/db"
	"downite/types"
	"errors"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kkdai/youtube/v2"
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

	mutexForDownloads    sync.Mutex
	downloads            map[int]*types.Download
	downloadsPrevSizeMap map[int]uint64
	httpClient           *http.Client
	Config               *DownloadClientConfig
	db                   *db.Database
	// part contexts for each download. we need to cancel them if download is cancelled
	mutexForPartContexts sync.Mutex
	partContextMap       map[int][]*contextWithCancel
}
type contextWithCancel struct {
	ctx    *context.Context
	cancel context.CancelFunc
}

func CreateDownloadClient(config *DownloadClientConfig, db *db.Database) (*Client, error) {
	return &Client{
		Config: config,
		httpClient: &http.Client{
			Transport: http.DefaultTransport,
		},
		downloadsPrevSizeMap: make(map[int]uint64),
		db:                   db,
	}, nil
}
func NewClientDefaultConfig() (*DownloadClientConfig, error) {
	executablePath, err := os.Executable()
	if err != nil {
		panic(fmt.Errorf("Cannot get executable path : %s", err))
	}
	defaultDownloadsDir := filepath.Join(filepath.Dir(executablePath), "/tmp/downloads")
	// Check if the directory exists
	_, err = os.Stat(defaultDownloadsDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Create the directory if it doesn't exist
			if err := os.MkdirAll(defaultDownloadsDir, os.ModePerm); err != nil {
				return nil, fmt.Errorf("Error creating directory: %s", err)
			}
		} else {
			return nil, fmt.Errorf("Error checking default downloads directory: %s", err)
		}
	}
	defaultClientConfig := DownloadClientConfig{
		DownloadPath: defaultDownloadsDir,
		PartCount:    8,
	}
	return &defaultClientConfig, nil
}
func (client *Client) InitDownloads() error {
	client.downloads = make(map[int]*types.Download, 0)
	client.partContextMap = make(map[int][]*contextWithCancel)

	downloads, err := client.db.GetDownloads()
	if err != nil {
		return err
	}
	for _, download := range downloads {
		// get the parts
		parts, err := client.db.GetDownloadParts(download.Id)
		if err != nil {
			return err
		}
		download.Parts = parts
		download.Progress = float64(download.DownloadedBytes) / float64(download.TotalSize) * 100

		go func() {
			client.AddDownload(&download)
			if download.Status == types.DownloadStatusDownloading.String() {
				err = client.StartDownload(download.Id)
				if err != nil {
					fmt.Printf("Error while starting download %s", err)
				}
			}
		}()
	}
	go client.updateDownloadSpeeds()

	return nil
}
func (client *Client) Stop() []error {
	errs := make([]error, 0)
	for _, download := range client.downloads {
		errs = append(errs, client.PauseDownload(download.Id))
	}
	return errs
}
func (client *Client) updateDownloadSpeeds() {
	for {
		client.mutexForDownloads.Lock()
		for _, download := range client.downloads {
			if download.Status == types.DownloadStatusDownloading.String() {
				prevSize := client.downloadsPrevSizeMap[download.Id]
				downloadedByteCount := download.DownloadedBytes - prevSize
				download.DownloadSpeed = downloadedByteCount / 1024

				//set new totalsize as prevsize
				client.downloadsPrevSizeMap[download.Id] = download.DownloadedBytes
			}
		}
		client.mutexForDownloads.Unlock()
		time.Sleep(time.Second)
	}
}
func (client *Client) PauseDownload(id int) error {
	if client.CheckDownloadStatus(id, types.DownloadStatusPaused) {
		return fmt.Errorf("download is already paused")
	}
	if client.CheckDownloadStatus(id, types.DownloadStatusCompleted) {
		return fmt.Errorf("download is already completed")
	}

	download, err := client.GetDownload(id)
	if err != nil {
		return err
	}
	fmt.Printf("Pausing download : %s \n", download.Name)

	//cancel all part downloads
	client.mutexForPartContexts.Lock()
	partContexts, ok := client.partContextMap[id]
	if !ok {
		return fmt.Errorf("download not found")
	}

	for _, ctxWithCancel := range partContexts {
		ctxWithCancel.cancel()
	}

	//delete part contexts from map
	delete(client.partContextMap, id)
	client.mutexForPartContexts.Unlock()

	client.mutexForDownloads.Lock()
	download.Status = types.DownloadStatusPaused.String()
	download.DownloadSpeed = 0
	client.mutexForDownloads.Unlock()
	err = client.db.UpdateDownload(download)
	if err != nil {
		return err
	}

	for _, downloadPart := range download.Parts {
		downloadPart.Status = types.DownloadStatusPaused.String()
		err = client.db.UpdateDownloadPart(downloadPart)
		if err != nil {
			return err
		}
	}

	return nil
}
func (client *Client) CheckDownloadStatus(id int, state types.DownloadStatus) bool {
	client.mutexForDownloads.Lock()
	defer client.mutexForDownloads.Unlock()
	download, ok := client.downloads[id]
	if !ok {
		return false
	}
	return download.Status == state.String()
}

func (client *Client) ResumeDownload(id int) error {
	if client.CheckDownloadStatus(id, types.DownloadStatusDownloading) {
		return fmt.Errorf("download is already running")
	}
	if client.CheckDownloadStatus(id, types.DownloadStatusCompleted) {
		return fmt.Errorf("download is already completed")
	}

	err := client.StartDownload(id)
	if err != nil {
		return fmt.Errorf("could not start download : %s\n", err)
	}

	download, err := client.GetDownload(id)
	if err != nil {
		return err
	}
	download.Status = types.DownloadStatusDownloading.String()
	err = client.db.UpdateDownload(download)
	if err != nil {
		return err
	}

	for _, downloadPart := range download.Parts {
		downloadPart.Status = types.DownloadStatusDownloading.String()
		err = client.db.UpdateDownloadPart(downloadPart)
		if err != nil {
			return err
		}
	}

	return nil
}
func (client *Client) updateDownloadState(id int, state types.DownloadStatus) error {
	client.mutexForDownloads.Lock()
	defer client.mutexForDownloads.Unlock()

	download, ok := client.downloads[id]
	if !ok {
		return fmt.Errorf("download not found")
	}
	download.Status = state.String()

	err := client.db.UpdateDownload(download)
	if err != nil {
		return err
	}

	return nil
}
func (client *Client) GetDownloadMeta(rawUrl string) (*types.DownloadMeta, error) {
	req, err := http.NewRequest("HEAD", rawUrl, nil)

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
	fileTypeHeader := res.Header.Get("Content-Type")

	if contentLengthHeader == "" {
		return nil, fmt.Errorf("cannot find content length in headers")
	}

	var fileName string
	var fileType string

	if contentDispositionHeader == "" {
		parsedUrl, err := url.Parse(rawUrl)
		if err != nil {
			return nil, fmt.Errorf("cannot parse url : %s", err)
		}
		filename := path.Base(parsedUrl.Path)
		if filename == "" {
			return nil, fmt.Errorf("cannot find file name")
		}
		fileName = filename
	} else {
		//if it has filename in its string
		fileName = getFileNameFromHeader(contentDispositionHeader)
	}

	if fileName == "" {
		return nil, fmt.Errorf("cannot find file name")
	}

	fileType = path.Ext(fileName)
	if fileType == "" && fileTypeHeader != "" {
		foundFileTypes, err := mime.ExtensionsByType(fileTypeHeader)
		if err != nil {
			return nil, err
		}
		if len(foundFileTypes) != 0 {
			fileType = foundFileTypes[len(foundFileTypes)-1]
			fileName = fileName + fileType
		}
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
		Url:            rawUrl,
		FileType:       fileType,
		IsRangeAllowed: rangesHeader == "bytes",
	}, nil

}

func (client *Client) DownloadFromUrl(rawUrl string, partCount int, savePath string, startDownload bool, addTopOfQueue bool) (*types.Download, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}
	// modify url for youtube links
	if parsedUrl.Host == "youtube" || parsedUrl.Host == "youtu.be" || parsedUrl.Host == "www.youtube.com" {
		youtubeClient := &youtube.Client{}
		video, err := youtubeClient.GetVideo(rawUrl)
		if err != nil {
			return nil, err
		}
		formats := video.Formats.Quality("1080p")
		if len(formats) == 0 {
			return nil, fmt.Errorf("no audio formats found")
		}
		var targetFormat *youtube.Format
		for _, format := range formats {
			if format.AudioChannels > 0 {
				targetFormat = &format
				break
			}
		}

		if targetFormat == nil {
			targetFormat = &formats[0]
		}

		videoUrl, err := youtubeClient.GetStreamURL(video, targetFormat)
		if err != nil {
			return nil, err
		}

		rawUrl = videoUrl
		parsedUrl, err = url.Parse(rawUrl)
		if err != nil {
			return nil, err
		}
		// destroy youtube client
		youtubeClient = nil
	}
	//GET METAINFO
	metaInfo, err := client.GetDownloadMeta(rawUrl)
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
		Url:             rawUrl,
		TotalSize:       uint64(metaInfo.TotalSize),
		DownloadedBytes: 0,
		Progress:        0,
		Status:          types.DownloadStatusPaused.String(),
	}

	//REGISTER DOWNLOAD to DB
	//from now on download has id from db
	err = client.RegisterDownload(download, addTopOfQueue)
	if err != nil {
		return nil, err
	}
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
			fmt.Printf("last queue number : %d \n", lastQueueNumber)
			download.QueueNumber = lastQueueNumber + 1
		}
	}

	id, err := client.db.InsertDownload(download, addTopOfQueue)
	if err != nil {
		return err
	}

	download.Id = id
	for i := 0; i < download.PartCount; i++ {

		startByteIndex := uint64(i) * download.PartLength
		endByteIndex := uint64((uint64(i)+1)*download.PartLength) - 1
		partLength := download.PartLength

		if i == download.PartCount-1 {
			//this is last part
			endByteIndex = download.TotalSize
			partLength = download.TotalSize - startByteIndex
		}

		download.Parts[i] = &types.DownloadPart{
			CreatedAt:       time.Now(),
			PartIndex:       i + 1,
			StartByteIndex:  startByteIndex,
			EndByteIndex:    endByteIndex,
			PartLength:      partLength,
			Status:          types.DownloadStatusPaused.String(),
			DownloadId:      download.Id,
			DownloadedBytes: 0,
			Progress:        0,
		}
	}

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
	client.downloadsPrevSizeMap[download.Id] = 0
}
func (client *Client) RemoveDownload(id int) error {
	err := client.PauseDownload(id)
	if err != nil {
		// TODO: improve error handling . we don't have error types
		if err.Error() != "download is already completed" && err.Error() != "download is already paused" {
			return err
		}
	}

	err = client.db.DeleteDownload(id)
	if err != nil {
		return err
	}

	err = client.deleteDownloadParts(id)
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
		if err != os.ErrNotExist {
			return err
		}
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

	partProcessChan := make(chan *types.DownloadPart, download.PartCount)
	errorChan := make(chan error)

	var completedPartCount int = 0

	downloadPartContexts := make([]*contextWithCancel, 0, download.PartCount)

	for _, part := range download.Parts {

		ctx, cancel := context.WithCancel(context.Background())

		downloadPartContexts = append(downloadPartContexts, &contextWithCancel{
			ctx:    &ctx,
			cancel: cancel,
		})

		// we are creating new goroutine for each part
		go func() {
			filePartBuffer, err := os.OpenFile(filepath.Join(download.SavePath, fmt.Sprintf("%s_part%d", download.Name, part.PartIndex)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				part.Status = types.DownloadStatusError.String()
				part.Error = err.Error()
				errorChan <- err
				return
			}
			defer filePartBuffer.Close()

			err = client.downloadFilePart(download, part, filePartBuffer, download.Url, ctx)
			if err != nil {
				// if download is canceled then return
				if errors.Is(err, context.Canceled) {
					return
				}
				part.Status = types.DownloadStatusError.String()
				part.Error = err.Error()
				errorChan <- err
				return
			}
			if part.DownloadedBytes == part.PartLength {
				partProcessChan <- part
				return
			}

			part.Status = types.DownloadStatusError.String()
			err = fmt.Errorf("downloaded bytes %d is not equal to end byte index %d", part.DownloadedBytes, part.EndByteIndex)
			part.Error = err.Error()
			errorChan <- err
		}()

	}

	client.mutexForPartContexts.Lock()
	client.partContextMap[download.Id] = downloadPartContexts
	client.mutexForPartContexts.Unlock()

	download.Status = types.DownloadStatusDownloading.String()
	download.StartedAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	err = client.db.UpdateDownload(download)
	if err != nil {
		return err
	}
	go func() {
		for completedPartCount != download.PartCount {
			select {
			case err := <-errorChan:
				fmt.Printf("Error while downloading file parts for %s : %s", download.Name, err)
			case partProcess := <-partProcessChan:
				completedPartCount += 1

				partProcess.Status = types.DownloadStatusCompleted.String()
				partProcess.FinishedAt = sql.NullTime{
					Time:  time.Now(),
					Valid: true,
				}
				err = client.db.UpdateDownloadPart(partProcess)
				if err != nil {
					fmt.Printf("Error while updating download part in db : %s", err)
					return
				}

				if completedPartCount != download.PartCount {
					continue
				}

				download.Status = types.DownloadStatusCompleted.String()
				download.FinishedAt = sql.NullTime{
					Time:  time.Now(),
					Valid: true,
				}
				err = client.db.UpdateDownload(download)
				if err != nil {
					fmt.Printf("Error while updating download in db : %s", err)
					return
				}

				fmt.Printf("download completed : %s \n", filepath.Join(download.SavePath, download.Name))

				_, err := os.Stat(filepath.Join(download.SavePath, download.Name))
				if err == nil {
					fmt.Printf("deleting existing file : %s \n", filepath.Join(download.SavePath, download.Name))
					err := os.Remove(filepath.Join(download.SavePath, download.Name))
					if err != nil {
						fmt.Printf("Error while deleting existing download file : %s \n", err)
						return
					}
				}
				downloadedFile, err := os.OpenFile(filepath.Join(download.SavePath, download.Name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Printf("Error while creating new download file : %s \n", err)
					return
				}
				defer downloadedFile.Close()
				for _, part := range download.Parts {
					if part.Status != types.DownloadStatusCompleted.String() {
						fmt.Printf("error download incomplete : %s \n", filepath.Join(download.SavePath, fmt.Sprintf("%s_part%d", download.Name, part.PartIndex)))
						return
					}
					partBuffer, err := os.ReadFile(filepath.Join(download.SavePath, fmt.Sprintf("%s_part%d", download.Name, part.PartIndex)))
					if err != nil {
						fmt.Printf("Error while reading part file : %s \n", err)
						return
					}
					_, err = downloadedFile.Write(partBuffer)
					if err != nil {
						fmt.Printf("Error while writing part file : %s \n", err)
						return
					}
				}

				downloadedFileStats, err := downloadedFile.Stat()
				if downloadedFileStats.Size() != int64(download.TotalSize) {
					fmt.Printf("Error downloaded bytes %d is not equal to total size %d", downloadedFileStats.Size(), download.TotalSize)
					return
				}
				err = client.deleteDownloadParts(download.Id)
				if err != nil {
					fmt.Printf("Error %s \n", err)
					return
				}
				break
			}
		}
	}()
	return nil
}

// delete download parts
func (client *Client) deleteDownloadParts(id int) error {
	client.mutexForPartContexts.Lock()
	delete(client.partContextMap, id)
	client.mutexForPartContexts.Unlock()

	client.mutexForDownloads.Lock()
	defer client.mutexForDownloads.Unlock()
	download, ok := client.downloads[id]
	if !ok {
		return fmt.Errorf("download not found")
	}

	parts, err := client.db.GetDownloadParts(id)
	if err != nil {
		return err
	}
	for _, part := range parts {
		fmt.Printf("removing part : %s_part%d \n", download.Name, part.PartIndex)
		err := os.Remove(filepath.Join(download.SavePath, fmt.Sprintf("%s_part%d", download.Name, part.PartIndex)))
		if err != nil {
			if err == os.ErrNotExist {
				continue
			}
			return fmt.Errorf("while deleting part file : %s \n", err)
		}
	}

	err = client.db.DeleteDownloadParts(id)
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

func (client *Client) downloadFilePart(download *types.Download, downloadPart *types.DownloadPart, filePart *os.File, url string, ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("while creating request: %s", err)
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", downloadPart.StartByteIndex+downloadPart.DownloadedBytes, downloadPart.EndByteIndex))

	res, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("while download : %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusPartialContent && res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code while downloading: %s", res.Status)
	}

	multiWriters := io.MultiWriter(downloadPart, download)
	downloadedFilePartReader := io.TeeReader(res.Body, multiWriters)
	_, err = io.Copy(filePart, downloadedFilePartReader)
	if err != nil {
		return err
	}

	return nil
}
