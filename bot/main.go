package main

import (
	"GoBot/tools"
	"GoBot/tools/discord"
	"GoBot/tools/sql"
	"GoBot/tools/youtube"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var INTERVAL = 1

func main() {
	initial()

	go func() {
		YoutubeStreamNotify()

		ticker := time.NewTicker(time.Duration(INTERVAL) * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			YoutubeStreamNotify()
		}
	}()

	select {}
}

func YoutubeStreamNotify() {
	for _, channel := range db.FindChannels() {
		defer func() {
			if r := recover(); r != nil {
				tools.DiscordNotify(s, "Youtube", channel.Title)
				tools.ErrorRecord(r)
			}
		}()

		defer fmt.Printf("%-20s Youtube notification end!\n", channel.Title)

		channelId, discordChannelId := channel.Id, channel.DiscordChannelId

		channel, err := youtube.GetChannel(channelId)
		if err != nil {
			panic(err)
		}

		baseEmbed := discord.BaseEmbed("Youtube", channel.Title, channel.Url, channel.Icon)

		videoIds := db.Distinct("video", channelId)
		videos, err := youtube.GetPlaylistItems(strings.Replace(channelId, "UC", "UU", 1), 3)
		if err != nil {
			panic(err)
		}

		for _, video := range videos {
			if IsContain(videoIds, video.Id) {
				continue
			}

			video, err := youtube.GetVideo(video.Id)
			if err != nil {
				panic(err)
			}

			if video.LiveStatus != 0 {
				baseEmbed.NewNotify(video).Send(s, discordChannelId)
			}

			db.Insert("Video", video.Map())
		}

		oldVideos := db.FindLivestreams(channelId)
		newVideos, err := youtube.GetVideos(db.Distinct("livestream", channelId))
		if err != nil {
			panic(err)
		}

		for i := range oldVideos {
			old, new := oldVideos[i], newVideos[i]

			if new.Private {
				new, err = youtube.GetVideo(new.Id)
				if err != nil {
					panic(err)
				}

				if new.Private {
					baseEmbed.New(old.Title, old.Url, "預定直播已被取消了！", old.Thumbnail).Send(s, discordChannelId)
					db.Update("Video", old.Id, "Private", new.Private)
				}
			} else if old.ScheduledTime != new.ScheduledTime {
				baseEmbed.New(new.Title, new.Url, "直播預定時間變更了！", new.Thumbnail).Change(old.ScheduledTime.String("full"), new.ScheduledTime.String("full")).Send(s, discordChannelId)
				db.Update("Video", new.Id, "ScheduledTime", new.ScheduledTime.String())
			} else if old.LiveStatus == 1 && new.LiveStatus == 2 {
				baseEmbed.New(new.Title, new.Url, "直播串流開始了！", new.Thumbnail).StartTime(new.StartTime).Send(s, discordChannelId)
				db.Update("Video", new.Id, "LiveStatus", new.LiveStatus, "StartTime", new.StartTime.String())
			} else if old.LiveStatus == 2 && new.LiveStatus == 0 {
				baseEmbed.New(new.Title, new.Url, "直播串流結束了！", new.Thumbnail).EndTime(new.EndTime, new.Length).Send(s, discordChannelId)
				db.Update("Video", new.Id, "LiveStatus", new.LiveStatus, "EndTime", new.EndTime.String(), "Length", new.Length.String())
			}
		}
	}
}

func initial() {
	logFile, err := os.OpenFile("/bot/error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}

	log.SetOutput(logFile)

	db, err = sql.ConnectToMySQL(os.Getenv("USERNAME"), os.Getenv("PASSWORD"), os.Getenv("HOST"), os.Getenv("PORT"), os.Getenv("DB_NAME"))
	if err != nil {
		log.Fatalf("Error connecting to MySQL: %v", err)
	}

	s, err = discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if handler, ok := commandsHandlers[i.ApplicationCommandData().Name]; ok {
				handler(s, i)
			}
		}
	})

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("Bot is running.")
	})

	for _, command := range commands {
		_, err := s.ApplicationCommandCreate(os.Getenv("APP_ID"), os.Getenv("GUILD_ID"), command)
		if err != nil {
			log.Fatalf("Cannot create slash command: %v", err)
		}
	}

	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()
}
