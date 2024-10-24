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
type DirectDownloadEngine struct {
	downloads            map[int]*types.Download
	httpClient           *http.Client
	DownloadClientConfig *DownloadClientConfig
	db                   *db.Database
	partContextMap       map[int][]*contextWithCancel
	onClose              []func()
	mutexForDownloads    sync.Mutex
	mutexForPartContexts sync.Mutex
}
type contextWithCancel struct {
	ctx    *context.Context
	cancel context.CancelFunc
}

func CreateDownloadClient(config *DownloadClientConfig, db *db.Database) (*DirectDownloadEngine, error) {
	return &DirectDownloadEngine{
		DownloadClientConfig: config,
		httpClient: &http.Client{
			Transport: http.DefaultTransport,
		},
		db: db,
	}, nil
}

func NewClientDefaultConfig() (*DownloadClientConfig, error) {
	executablePath, err := os.Executable()
	if err != nil {
		panic(fmt.Errorf("cannot get executable path : %s", err))
	}
	defaultDownloadsDir := filepath.Join(filepath.Dir(executablePath), "/tmp/downloads")
	// Check if the directory exists
	_, err = os.Stat(defaultDownloadsDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Create the directory if it doesn't exist
			if err := os.MkdirAll(defaultDownloadsDir, os.ModePerm); err != nil {
				return nil, fmt.Errorf("creating directory: %s", err)
			}
		} else {
			return nil, fmt.Errorf("checking default downloads directory: %s", err)
		}
	}
	defaultClientConfig := DownloadClientConfig{
		DownloadPath: defaultDownloadsDir,
		PartCount:    8,
	}
	return &defaultClientConfig, nil
}

func (client *DirectDownloadEngine) InitDownloads() error {
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

func (client *DirectDownloadEngine) Stop() []error {
	errs := make([]error, 0)
	for _, download := range client.downloads {
		errs = append(errs, client.PauseDownload(download.Id))
	}
	return errs
}

func (client *DirectDownloadEngine) updateDownloadSpeeds() {
	for {
		// we calculate time to take mutex. because we need to calculate exact download speed per second
		start := time.Now()
		client.mutexForDownloads.Lock()
		timeToTakeMutex := time.Since(start)

		for _, download := range client.downloads {
			download.DownloadSpeed = download.BytesWritten / 1024

			// reset bytes written
			download.BytesWritten = 0
		}
		client.mutexForDownloads.Unlock()
		time.Sleep(time.Second - timeToTakeMutex)
	}
}

func (client *DirectDownloadEngine) CheckDownloadStatus(id int, state types.DownloadStatus) bool {
	download, err := client.GetDownload(id)
	if err != nil {
		return false
	}
	client.mutexForDownloads.Lock()
	defer client.mutexForDownloads.Unlock()

	return download.Status == state.String()
}

func (client *DirectDownloadEngine) PauseDownload(id int) error {
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

	client.mutexForDownloads.Lock()
	fmt.Printf("Pausing download : %s \n", download.Name)
	client.mutexForDownloads.Unlock()

	// cancel all part downloads
	client.mutexForPartContexts.Lock()
	partContexts, ok := client.partContextMap[id]
	if !ok {
		return fmt.Errorf("download not found")
	}

	for _, ctxWithCancel := range partContexts {
		ctxWithCancel.cancel()
	}

	// delete part contexts from map
	delete(client.partContextMap, id)
	client.mutexForPartContexts.Unlock()

	err = client.updateDownloadStatus(id, types.DownloadStatusPaused)
	if err != nil {
		return err
	}
	return nil
}

func (client *DirectDownloadEngine) ResumeDownload(id int) error {
	if client.CheckDownloadStatus(id, types.DownloadStatusDownloading) {
		return fmt.Errorf("download is already running")
	}
	if client.CheckDownloadStatus(id, types.DownloadStatusCompleted) {
		return fmt.Errorf("download is already completed")
	}

	download, err := client.GetDownload(id)
	if err != nil {
		return err
	}

	client.mutexForDownloads.Lock()
	fmt.Printf("Resuming download : %s \n", download.Name)
	isMultiPart := download.IsMultiPart
	client.mutexForDownloads.Unlock()

	if !isMultiPart {
		fmt.Printf("Reinitializing download progress because it is not multi part : %s \n", download.Name)
		err = client.ReinitilizeDownload(id)
		if err != nil {
			return err
		}
	}
	err = client.StartDownload(id)
	if err != nil {
		return fmt.Errorf("could not start download : %s", err)
	}

	return nil
}

func (client *DirectDownloadEngine) resetDownloadSpeed(id int) error {
	download, err := client.GetDownload(id)
	if err != nil {
		return err
	}
	client.mutexForDownloads.Lock()
	defer client.mutexForDownloads.Unlock()

	download.DownloadSpeed = 0
	return nil
}

func (client *DirectDownloadEngine) updateDownloadStatus(id int, status types.DownloadStatus) error {
	download, err := client.GetDownload(id)
	if err != nil {
		return err
	}

	client.mutexForDownloads.Lock()
	defer client.mutexForDownloads.Unlock()

	download.Status = status.String()

	for _, downloadPart := range download.Parts {
		downloadPart.Status = status.String()
		err := client.db.UpdateDownloadPart(downloadPart)
		if err != nil {
			return err
		}
	}

	err = client.db.UpdateDownload(download)
	if err != nil {
		return err
	}

	return nil
}

func (client *DirectDownloadEngine) CheckDownload(rawUrl string, fileName string, fileSize uint64) (bool, int) {
	client.mutexForDownloads.Lock()
	defer client.mutexForDownloads.Unlock()
	for _, download := range client.downloads {
		if download.Url == rawUrl && download.Name == fileName && download.TotalSize == fileSize {
			return true, download.Id
		}
	}
	return false, 0
}

func (client *DirectDownloadEngine) GetDownloadMeta(rawUrl string) (*types.DownloadMeta, error) {
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

		format := GetBestHighFormat(FilterFormats(video.Formats, "video"))
		fmt.Printf("format : %+v \n", format)
		videoUrl, err := youtubeClient.GetStreamURL(video, &format)
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

	req, err := http.NewRequest("HEAD", rawUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("while creating request: %s", err)
	}

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("while head request: %s", err)
	}
	// check if server accepts split downloads
	rangesHeader := res.Header.Get("Accept-Ranges")
	// total file size
	contentLengthHeader := res.Header.Get("Content-Length")
	// EXAMPLE HEADER = "attachment; filename=\"test.txt\""
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
		// if it has filename in its string
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

	ok, id := client.CheckDownload(rawUrl, fileName, contentLength)
	if ok {
		// return nil, fmt.Errorf("download with url %s already exists", rawUrl)
	}

	return &types.DownloadMeta{
		FileName:           fileName,
		TotalSize:          contentLength,
		Url:                rawUrl,
		FileType:           fileType,
		IsRangeAllowed:     rangesHeader == "bytes",
		IsExist:            ok,
		ExistingDownloadId: id,
	}, nil
}

// FilterFormats filters a list of YouTube formats based on a specific kind (e.g. "video/mp4")
func FilterFormats(formats youtube.FormatList, kind string) []youtube.Format {
	var filteredFormats []youtube.Format
	for _, format := range formats {
		if strings.Contains(format.MimeType, kind) {
			filteredFormats = append(filteredFormats, format)
		}
	}
	return filteredFormats
}

// GetBestHighFormat returns the format with the highest bitrate from a list of formats
func GetBestHighFormat(formats []youtube.Format) youtube.Format {
	var bestFormat youtube.Format
	for _, format := range formats {
		if format.Bitrate > bestFormat.Bitrate {
			bestFormat = format
		}
	}
	return bestFormat
}

func (client *DirectDownloadEngine) DownloadFromUrl(name string, rawUrl string, partCount int, savePath string, startDownload bool, addTopOfQueue bool, overwrite bool) (*types.Download, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}
	// modify url for youtube links
	if parsedUrl.Host == "youtube" || parsedUrl.Host == "youtu.be" || parsedUrl.Host == "www.youtube.com" {
		youtubeClient := &youtube.Client{}
		video, err := youtubeClient.GetVideo("https://www.youtube.com/embed/u9lj-c29dxI")
		if err != nil {
			return nil, err
		}

		format := GetBestHighFormat(FilterFormats(video.Formats, "video"))
		fmt.Printf("format : %+v \n", format)
		videoUrl, err := youtubeClient.GetStreamURL(video, &format)
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
	// GET METAINFO
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
		savePath = client.DownloadClientConfig.DownloadPath
	}

	if name == "" {
		name = metaInfo.FileName
	}

	if !overwrite {
		// CHECK IF FILE EXISTS
		if _, err := os.Stat(filepath.Join(savePath, name)); err == nil {
			return nil, fmt.Errorf("file already exists")
		}
	}

	download := &types.Download{
		CreatedAt:       time.Now(),
		Parts:           make([]*types.DownloadPart, partCount),
		Name:            name,
		SavePath:        savePath,
		PartCount:       partCount,
		PartLength:      partLength,
		Url:             rawUrl,
		TotalSize:       uint64(metaInfo.TotalSize),
		DownloadedBytes: 0,
		Progress:        0,
		IsMultiPart:     metaInfo.IsRangeAllowed,
		Status:          types.DownloadStatusPaused.String(),
	}

	// REGISTER DOWNLOAD to DB
	// from now on download has id from db
	err = client.RegisterDownload(download, addTopOfQueue)
	if err != nil {
		return nil, err
	}
	// ADD DOWNLOAD TO client
	client.AddDownload(download)
	// START SPLIT DOWNLOAD
	if startDownload {
		err := client.StartDownload(download.Id)
		if err != nil {
			return nil, err
		}
	}

	return download, nil
}

func (client *DirectDownloadEngine) RegisterDownload(download *types.Download, addTopOfQueue bool) error {
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

	err = client.RegisterDownloadParts(download)
	if err != nil {
		return err
	}

	return nil
}

func (client *DirectDownloadEngine) RegisterDownloadParts(download *types.Download) error {
	for i := 0; i < download.PartCount; i++ {

		startByteIndex := uint64(i) * download.PartLength
		endByteIndex := uint64((uint64(i)+1)*download.PartLength) - 1
		partLength := download.PartLength

		if i == download.PartCount-1 {
			// this is last part
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

	err := client.db.InsertDownloadParts(download.Parts)
	if err != nil {
		return err
	}

	return nil
}

func (client *DirectDownloadEngine) CreateNewFileNameForPath(path string, fileName string) (string, error) {
	fileExt := filepath.Ext(fileName)
	fileNameWithoutExt := strings.TrimSuffix(fileName, fileExt)
	newNumber := 0
	_, err := os.Stat(filepath.Join(path, fileName))
	if err != nil {
		if os.IsNotExist(err) {
			newNumber = 2
			return fmt.Sprintf("%s_%d%s", fileNameWithoutExt, newNumber, fileExt), nil
		} else {
			return fileName, err
		}
	}
	for i := 3; ; i++ {
		_, err := os.Stat(filepath.Join(path, fmt.Sprintf("%s_%d%s", fileNameWithoutExt, i, fileExt)))
		if err != nil {
			if os.IsNotExist(err) {
				newNumber = i
				break
			} else {
				return fileName, err
			}
		}
	}

	return fmt.Sprintf("%s_%d%s", fileNameWithoutExt, newNumber, fileExt), nil
}

func (client *DirectDownloadEngine) AddDownload(download *types.Download) {
	client.mutexForDownloads.Lock()
	defer client.mutexForDownloads.Unlock()

	client.downloads[download.Id] = download
}

func (client *DirectDownloadEngine) RemoveDownload(id int) error {
	err := client.PauseDownload(id)
	if err != nil {
		// TODO: improve error handling . we don't have error types
		if err.Error() != "download is already completed" && err.Error() != "download is already paused" {
			return err
		}
	}

	client.mutexForDownloads.Lock()
	fileName := client.downloads[id].Name
	client.mutexForDownloads.Unlock()
	fmt.Printf("Removing download : %s \n", fileName)

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

func (client *DirectDownloadEngine) ReinitilizeDownload(id int) error {
	download, err := client.GetDownload(id)
	if err != nil {
		return err
	}

	err = client.deleteDownloadParts(id)
	if err != nil {
		return err
	}

	client.mutexForDownloads.Lock()
	defer client.mutexForDownloads.Unlock()

	download.DownloadedBytes = 0
	download.Progress = 0
	download.StartedAt = sql.NullTime{
		Time:  time.Time{},
		Valid: false,
	}
	download.CurrentWrittenBytes = 0

	err = client.db.UpdateDownload(download)
	if err != nil {
		return err
	}

	err = client.RegisterDownloadParts(download)
	if err != nil {
		return err
	}

	return nil
}

func (client *DirectDownloadEngine) updateDownloadQueueNumbers() error {
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

func (client *DirectDownloadEngine) DeleteDownload(id int) error {
	client.mutexForDownloads.Lock()
	savePath := client.downloads[id].SavePath
	fileName := client.downloads[id].Name
	client.mutexForDownloads.Unlock()

	err := client.RemoveDownload(id)
	if err != nil {
		return err
	}

	fmt.Printf("Deleting download : %s \n", fileName)
	err = os.RemoveAll(filepath.Join(savePath, fileName))
	if err != nil {
		if err != os.ErrNotExist {
			return err
		}
	}
	return nil
}

func (client *DirectDownloadEngine) GetDownload(id int) (*types.Download, error) {
	client.mutexForDownloads.Lock()
	defer client.mutexForDownloads.Unlock()

	download, ok := client.downloads[id]
	if !ok {
		return nil, fmt.Errorf("download not found")
	}
	return download, nil
}

func (client *DirectDownloadEngine) GetDownloads() ([]*types.Download, error) {
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

func (client *DirectDownloadEngine) StartDownload(id int) error {
	download, err := client.GetDownload(id)
	if err != nil {
		return err
	}
	fmt.Printf("starting download : %s \n", filepath.Join(download.SavePath, download.Name))

	partProcessChan := make(chan *types.DownloadPart, download.PartCount)
errorChhttps: // www.reddit.com/r/Damnthatsinteresting/comments/1fz41kc/using_the_crispr_technique_to_genetically_modify/an := make(chan error)

	completedPartCount := 0

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

			isRangeAllowed := len(download.Parts) != 1

			err = client.downloadFilePart(download, part, filePartBuffer, download.Url, ctx, isRangeAllowed)
			if err != nil {
				// if download is canceled then return
				if errors.Is(err, context.Canceled) {
					part.Error = err.Error()
					errorChan <- err
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

	err = client.updateDownloadStatus(id, types.DownloadStatusDownloading)
	if err != nil {
		return err
	}
	download.StartedAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	err = client.db.UpdateDownload(download)
	if err != nil {
		return err
	}
	go func() {
		// TODO:(ft-aslan) we need to use mutex for this. but we need more compact way
		for completedPartCount != download.PartCount {
			select {
			case err := <-errorChan:
				fmt.Printf("Error while downloading file parts for %s : %s", download.Name, err)
				return
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

				if completedPartCount == download.PartCount {
					break
				} else {
					continue
				}
			}
		}

		// the download is completed now
		err = client.updateDownloadStatus(id, types.DownloadStatusCompleted)
		if err != nil {
			fmt.Printf("Error while updating download status in db : %s", err)
			return
		}
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
	}()
	return nil
}

// delete download parts with files
func (client *DirectDownloadEngine) deleteDownloadParts(id int) error {
	client.mutexForPartContexts.Lock()
	delete(client.partContextMap, id)
	client.mutexForPartContexts.Unlock()

	download, err := client.GetDownload(id)
	if err != nil {
		return err
	}

	client.mutexForDownloads.Lock()
	defer client.mutexForDownloads.Unlock()

	parts, err := client.db.GetDownloadParts(id)
	if err != nil {
		return err
	}
	for _, part := range parts {
		partPath := filepath.Join(download.SavePath, fmt.Sprintf("%s_part%d", download.Name, part.PartIndex))

		_, err := os.Stat(partPath)
		if err != nil {
			if err == os.ErrNotExist {
				continue
			}
			return fmt.Errorf("while checking part file : %s \n", err)
		}
		fmt.Printf("removing part : %s_part%d \n", download.Name, part.PartIndex)
		err = os.Remove(partPath)
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

func (client *DirectDownloadEngine) downloadFilePart(download *types.Download, downloadPart *types.DownloadPart, filePart *os.File, url string, ctx context.Context, isRangeAllowed bool) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("while creating request: %s", err)
	}

	if isRangeAllowed {
		req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", downloadPart.StartByteIndex+downloadPart.DownloadedBytes, downloadPart.EndByteIndex))
	}

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

func (client *DirectDownloadEngine) GetTotalDownloadSpeed() uint64 {
	client.mutexForDownloads.Lock()
	defer client.mutexForDownloads.Unlock()
	var totalDownloadSpeed uint64
	for _, download := range client.downloads {
		totalDownloadSpeed += download.DownloadSpeed
	}
	return totalDownloadSpeed
}
