package youtube

import (
	"GoBot/tools"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var apiKey = os.Getenv("YOUTUBE_API_KEY")

func getData(path string) (*tools.Json, error) {
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/%s&key=%s", path, apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &tools.Json{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &tools.Json{}, err
	}
	defer resp.Body.Close()

	data, err := tools.ToJson(resp.Body)
	if err != nil {
		return &tools.Json{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return &tools.Json{}, fmt.Errorf("HTTP request failed with status code: %d\n%s", resp.StatusCode, string(body))
	}

	return data, nil
}

func GetChannel(channelId string) (Channel, error) {
	path := fmt.Sprintf("channels?part=brandingSettings,snippet,statistics&id=%s", channelId)
	data, err := getData(path)
	if err != nil {
		return Channel{}, err
	}

	if !data.Exist("items") {
		return Channel{}, errors.New("fail to get data")
	}

	item := data.Get("items").Index(0)
	channel := Channel{
		Id:       item.Get("id").String(),
		CustomId: item.Get("snippet").Get("customUrl").Slice(1, -1),
		Url:      fmt.Sprintf("https://www.youtube.com/channel/%s", item.Get("id").String()),
		Title:    item.Get("snippet").Get("title").String(),
		Icon:     item.Get("snippet").Get("thumbnails").Image(),
	}

	return channel, nil
}

func GetPlaylistItems(playlistId string, num int) ([]Video, error) {
	var videos []Video
	var pageToken string

	for {
		path := fmt.Sprintf("playlistItems?part=snippet&playlistId=%s&maxResults=%d&pageToken=%s", playlistId, num, pageToken)
		data, err := getData(path)
		if err != nil {
			return []Video{}, err
		}

		if !data.Exist("items") {
			return []Video{}, errors.New("fail to get data")
		}

		pageToken = data.Get("nextPageToken").String()

		for _, item := range data.Get("items").JsonArray() {
			video := Video{
				Id:        item.Get("snippet").Get("resourceId").Get("videoId").String(),
				Url:       fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.Get("snippet").Get("resourceId").Get("videoId").String()),
				Title:     item.Get("snippet").Get("title").String(),
				Thumbnail: item.Get("snippet").Get("thumbnails").Image(),
			}

			videos = append(videos, video)
		}

		if pageToken == "" || num != 50 {
			break
		}
	}

	return videos, nil
}

func GetVideo(videoId string) (Video, error) {
	path := fmt.Sprintf("videos?part=contentDetails,liveStreamingDetails,snippet,statistics&id=%s", videoId)
	data, err := getData(path)
	if err != nil {
		return Video{}, err
	}

	if !data.Exist("items") {
		return Video{}, errors.New("fail to get data")
	}

	var item *tools.Json

	if len(data.Get("items").JsonArray()) != 0 {
		item = data.Get("items").Index(0)
	} else {
		item = nil
	}

	return getVideoStruct(item, videoId), nil
}

func GetVideos(videoIds []string) ([]Video, error) {
	var videos []Video
	length := len(videoIds)

	for n := 0; n < length; n += 50 {
		end := n + 50

		if end > length {
			end = length
		}

		path := fmt.Sprintf("videos?part=contentDetails,liveStreamingDetails,snippet,statistics&id=%s", strings.Join(videoIds[n:end], ","))
		data, err := getData(path)
		if err != nil {
			return []Video{}, err
		}

		if !data.Exist("items") {
			return []Video{}, errors.New("fail to get data")
		}

		dataList := data.Get("items").JsonArray()

		for _, videoId := range videoIds[n:end] {
			var item *tools.Json

			if len(dataList) != 0 && videoId == dataList[0].Get("id").String() {
				item = dataList[0]
				dataList = dataList[1:]
			} else {
				item = nil
			}

			videos = append(videos, getVideoStruct(item, videoId))
		}
	}

	return videos, nil
}
