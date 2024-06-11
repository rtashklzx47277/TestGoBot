package main

import (
	"GoBot/tools/sql"
	"GoBot/tools/youtube"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	db *sql.MySQL
	s  *discordgo.Session
)

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "add",
		Description: "Add New Youtube Channel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "channel-id",
				Description: "Youtube Channel Id",
				Required:    true,
			},
		},
	},
	{
		Name:        "delete",
		Description: "Delete New Youtube Channel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "channel-id",
				Description: "Youtube Channel Id",
				Required:    true,
			},
		},
	},
	{
		Name:        "list",
		Description: "List Youtube Channels",
	},
}

var commandsHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"add": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if r := recover(); r != nil {
				sendResponse(s, i, "似乎發生了什麼錯誤...")
				log.Println("handler panic:", r)
			}
		}()

		channelId := i.ApplicationCommandData().Options[0].StringValue()

		if db.Find("Channel", "WHERE Id = ?", channelId) {
			channel := db.FindChannel(channelId)
			sendResponse(s, i, fmt.Sprintf("**%s**已在通知頻道清單中！", channel.Title))
		} else {
			channel, err := youtube.GetChannel(channelId)
			if err != nil {
				sendResponse(s, i, "無法取得頻道資料，請確認頻道ID是否正確。")
				return
			}

			channel.DiscordChannelId = i.Message.ChannelID

			db.Insert("Channel", channel.Map())
			sendResponse(s, i, fmt.Sprintf("**%s**已新增至通知頻道清單中！", channel.Title))
		}
	},
	"delete": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if r := recover(); r != nil {
				sendResponse(s, i, "似乎發生了什麼錯誤...")
				log.Println("handler panic:", r)
			}
		}()

		channelId := i.ApplicationCommandData().Options[0].StringValue()

		if db.Find("Channel", "WHERE Id = ?", channelId) {
			channel := db.FindChannel(channelId)
			db.Delete("Channel", "WHERE Id = ?", channelId)
			sendResponse(s, i, fmt.Sprintf("已將**%s**自通知頻道清單中刪除！", channel.Title))
		} else {
			sendResponse(s, i, "此頻道不在通知頻道清單中！")
		}
	},
	"list": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if r := recover(); r != nil {
				sendResponse(s, i, "似乎發生了什麼錯誤...")
				log.Println("handler panic:", r)
			}
		}()

		var channelsList []string

		for _, channel := range db.FindChannels() {
			channelsList = append(channelsList, fmt.Sprintf("%s %s", channel.Title, channel.DiscordChannelId))
		}

		sendResponse(s, i, strings.Join(channelsList, "\n"))
	},
}

func sendResponse(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}
