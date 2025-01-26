package fansly

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// UserData represents the data needed from the user to interact with the Fansly API
type UserData struct {
	FanslyToken   string
	UserAgent     string
	BasePath      string
	FanslyCreator string
	Metadata      bool
}

// AccountData represents the data returned by the Fansly API // "https://apiv3.fansly.com/api/v1/account?usernames=<username>"
type AccountData struct {
	Response []struct {
		Id          string `json:"id"`
		Username    string `json:"username"`
		DisplayName string `json:"displayName"`
	} `json:"response"`
}

// StreamData represents the data returned by the Fansly API // "https://apiv3.fansly.com/api/v1/streaming/channel/<account_id>"
type StreamData struct {
	Success  bool `json:"success"`
	Response struct {
		Id          string `json:"id"`
		AccountId   string `json:"accountId"`
		PlaybackUrl string `json:"playbackUrl"`
		ChatRoomId  string `json:"chatRoomId"`
		Status      int    `json:"status"`
		Version     int    `json:"version"`
		CreatedAt   int64  `json:"createdAt"`
		UpdatedAt   *int64 `json:"updatedAt"`
		Stream      struct {
			Id            string `json:"id"`
			HistoryId     string `json:"historyId"`
			ChannelId     string `json:"channelId"`
			AccountId     string `json:"accountId"`
			Title         string `json:"title"`
			Status        int    `json:"status"`
			ViewerCount   int    `json:"viewerCount"`
			Version       int    `json:"version"`
			CreatedAt     int64  `json:"createdAt"`
			UpdatedAt     *int64 `json:"updatedAt"`
			LastFetchedAt int64  `json:"lastFetchedAt"`
			StartedAt     int64  `json:"startedAt"` // Unix timestamp in milliseconds
			Permissions   struct {
				PermissionFlags        []interface{} `json:"permissionFlags"`
				AccountPermissionFlags struct {
					Flags    int    `json:"flags"`
					Metadata string `json:"metadata"`
				} `json:"accountPermissionFlags"`
			} `json:"permissions"`
			Whitelisted            bool   `json:"whitelisted"`
			AccountPermissionFlags int    `json:"accountPermissionFlags"`
			Access                 bool   `json:"access"`
			PlaybackUrl            string `json:"playbackUrl"`
		} `json:"stream"`
		Arn            *string `json:"arn"`
		IngestEndpoint *string `json:"ingestEndpoint"`
	} `json:"response"`
}

// createDirIfNotExist creates a directory if it does not exist

func (u UserData) createDirIfNotExist() (string, error) {
	dirPath := filepath.Join(u.BasePath, "FanslyDownloader", "Fansly", u.FanslyCreator)

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return absPath, nil
}

// request makes a GET request to the Fansly API
func request(url string, u *UserData) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", u.UserAgent)
	req.Header.Set("Referer", "https://fansly.com")
	req.Header.Set("Authorization", u.FanslyToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error occurred getting data from the Fansly API. Make sure your token/user agent is valid\n%v", err)
	}
	return resp.Body, nil
}

// getAccountData fetches the account data for the creator
func getAccountData(u *UserData) (AccountData, error) {
	var accountData AccountData
	url := "https://apiv3.fansly.com/api/v1/account?usernames=" + u.FanslyCreator
	resp, err := request(url, u)
	if err != nil {
		fmt.Println(err)
		return AccountData{}, err
	}
	defer resp.Close()
	if err := json.NewDecoder(resp).Decode(&accountData); err != nil {
		return AccountData{}, err
	}
	return accountData, nil
}

// getStreamData fetches the stream data for the livestream
func getStreamData(a *AccountData, u *UserData) (StreamData, error) {
	var streamData StreamData
	url := "https://apiv3.fansly.com/api/v1/streaming/channel/" + a.Response[0].Id
	resp, err := request(url, u)
	if err != nil {
		return StreamData{}, err
	}
	defer resp.Close()
	if err := json.NewDecoder(resp).Decode(&streamData); err != nil {
		return StreamData{}, err
	}
	return streamData, nil

}

// getPlaybackURL returns the playback URL of the stream
func (s *StreamData) getPlaybackURL() string {
	return s.Response.Stream.PlaybackUrl
}

// getStartedAt returns the start time of the stream in Unix timestamp in milliseconds
func (s *StreamData) getStartedAt() int64 {
	return s.Response.Stream.StartedAt
}

// getStreamStart returns the start time of the stream in the format "2006-01-02_15:04:05"
func (s StreamData) getStreamStart() string {
	return time.Unix(s.getStartedAt()/1000, 0).Format("2006-01-02_15:04:05")
}

// DownloadStream downloads the stream of the creator
func DownloadStream(u *UserData, download bool) error {
	logger := log.New(os.Stdout, "FanslyDownloader: ", log.LstdFlags)
	logger.Println("Fetching account data for creator:", u.FanslyCreator)
	accD, err := getAccountData(u)
	if err != nil {
		logger.Fatalf("Error getting account data: %v", err)
	}

	strD, err := getStreamData(&accD, u)
	if err != nil {
		logger.Fatalf("Error getting stream data: %v", err)
	}

	if !strD.Response.Stream.Access {
		logger.Fatalf("Stream is not accessible or no stream data available")
	}
	playbackURL := strD.getPlaybackURL()
	if playbackURL == "" {
		logger.Fatalf("%v is not streaming right now", u.FanslyCreator)
	}
	streamDate := strD.getStreamStart()
	tempFilePath, _ := u.createDirIfNotExist()
	outputDir, _ := filepath.Abs(fmt.Sprintf("%s%s%s.mp4", tempFilePath, string(filepath.Separator), streamDate))
	if download {
		cmd := exec.Command(
			"ffmpeg", "-hide_banner", "-v", "warning", "-stats", "-http_persistent", "0",
			"-user_agent", u.UserAgent,
			"-headers", fmt.Sprintf("Authorization: %s\r\n", u.FanslyToken),
			"-headers", "Referer: https://fansly.com\r\n",
			"-i", playbackURL, "-c", "copy",
			outputDir)

		_, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error downloading stream: %v", err)
		}
		fmt.Println("Stream downloaded successfully")
	}
	if u.Metadata {
		writeMetadata(u, &strD)
	}
	return nil
}

func writeMetadata(u *UserData, streamData *StreamData) {
	// Create metadata file
	metadataFilepath := filepath.Join(u.BasePath, streamData.getStreamStart()+".json")
	metadataFile, err := os.Create(metadataFilepath)
	if err != nil {
		log.Fatalf("Could not create metadata file: %v", err)
	}
	defer metadataFile.Close()

	// Write metadata to file
	metadata, err := json.MarshalIndent(streamData, "", "  ")

	if err != nil {
		log.Fatalf("Could not marshal metadata: %v", err)
	}
	metadataFile.Write(metadata)
}
