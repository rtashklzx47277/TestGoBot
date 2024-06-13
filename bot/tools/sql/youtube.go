package sql

import (
	"GoBot/tools/youtube"
	"database/sql"
	"fmt"
)

func (mySQL *MySQL) FindChannel(channelId string) youtube.Channel {
	var channel youtube.Channel
	var customId, title, discordChannelId sql.NullString

	query := "SELECT Id, CustomId, Title, DiscordChannelId FROM Channel WHERE Id = ?"
	err := mySQL.db.QueryRow(query, channelId).Scan(&channel.Id, &customId, &title, &discordChannelId)
	if err != nil {
		fmt.Println(err)
	}

	channel.CustomId = handleNullString(customId)
	channel.Title = handleNullString(title)
	channel.DiscordChannelId = handleNullString(discordChannelId)
	channel.Url = fmt.Sprintf("https://www.youtube.com/channel/%s", channel.Id)

	return channel
}

func (mySQL *MySQL) FindChannels() []youtube.Channel {
	query := "SELECT Id, CustomId, Title, DiscordChannelId FROM Channel"
	rows, err := mySQL.db.Query(query)
	if err != nil {
		fmt.Println(err)
	}

	var channels []youtube.Channel

	for rows.Next() {
		var channel youtube.Channel

		var customId, title sql.NullString

		err := rows.Scan(&channel.Id, &customId, &title, &channel.DiscordChannelId)
		if err != nil {
			fmt.Println(err)
		}

		channel.CustomId = handleNullString(customId)
		channel.Title = handleNullString(title)
		channel.Url = fmt.Sprintf("https://www.youtube.com/channel/%s", channel.Id)

		channels = append(channels, channel)
	}

	return channels
}

func (mySQL *MySQL) FindLivestreams(channelId string) []youtube.Video {
	query := "SELECT DISTINCT Id, Title, LiveStatus, ScheduledTime FROM Video WHERE ChannelId = ? AND LiveStatus <> ? AND Private = ?"
	rows, err := mySQL.db.Query(query, channelId, 0, 0)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var livestreams []youtube.Video

	for rows.Next() {
		var livestream youtube.Video
		var title, scheduledTime sql.NullString
		var liveStatus sql.NullInt64

		err := rows.Scan(&livestream.Id, &title, &liveStatus, &scheduledTime)
		if err != nil {
			fmt.Println(err)
		}

		livestream.Title = handleNullString(title)
		livestream.LiveStatus = handleNullInt(liveStatus)
		livestream.ScheduledTime = stringToTime(handleNullString(scheduledTime))

		livestream.Thumbnail = fmt.Sprintf("https://i.ytimg.com/vi/%s/maxresdefault.jpg", livestream.Id)
		livestream.Url = fmt.Sprintf("https://www.youtube.com/watch?v=%s", livestream.Id)

		livestreams = append(livestreams, livestream)
	}

	return livestreams
}

func (mySQL *MySQL) Distinct(target, channelId string) []string {
	var query string
	var values []any

	switch target {
	case "video":
		query = "SELECT DISTINCT Id FROM Video WHERE ChannelId = ?"
		values = append(values, channelId)
	case "livestream":
		query = "SELECT DISTINCT Id FROM Video WHERE ChannelId = ? AND LiveStatus <> ? AND Video.Private = ?"
		values = append(values, channelId, 0, 0)
	}

	rows, err := mySQL.db.Query(query, values...)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var result []string

	for rows.Next() {
		var value string

		err := rows.Scan(&value)
		if err != nil {
			fmt.Println(err)
		}

		result = append(result, value)
	}

	return result
}
