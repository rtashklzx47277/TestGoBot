package main

import (
	"GoBot/tools/sql"
	"GoBot/tools/youtube"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	db      *sql.MySQL
	s       *discordgo.Session
	guildId = os.Getenv("GUILD_ID")
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
	{
		Name:        "help",
		Description: "List Commands",
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

		channelId, check := checkChannelId(i.ApplicationCommandData().Options[0].StringValue())
		if !check {
			sendResponse(s, i, "無法取得頻道資料，請確認頻道ID是否正確或重新嘗試。")
			return
		}

		if db.Find("Channel", "WHERE Id = ?", channelId) {
			channel := db.FindChannel(channelId)
			sendResponse(s, i, fmt.Sprintf("**%s**已在關注頻道清單中！", channel.Title))
			return
		}

		channel, err := youtube.GetChannel(channelId)
		if err != nil {
			sendResponse(s, i, "無法取得頻道資料，請確認頻道ID是否正確或重新嘗試。")
			return
		}

		channel.DiscordChannelId = i.ChannelID

		db.Insert("Channel", channel.Map())
		sendResponse(s, i, fmt.Sprintf("**%s**已新增至關注頻道清單中！", channel.Title))
	},
	"delete": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if r := recover(); r != nil {
				sendResponse(s, i, "似乎發生了什麼錯誤...")
				log.Println("handler panic:", r)
			}
		}()

		channelId, check := checkChannelId(i.ApplicationCommandData().Options[0].StringValue())
		if !check {
			sendResponse(s, i, "無法取得頻道資料，請確認頻道ID是否正確或重新嘗試。")
			return
		}

		if !db.Find("Channel", "WHERE Id = ?", channelId) {
			sendResponse(s, i, "此頻道不在關注頻道清單中！")
			return
		}

		channel := db.FindChannel(channelId)
		db.Delete("Channel", "WHERE Id = ?", channelId)
		sendResponse(s, i, fmt.Sprintf("已將**%s**自關注頻道清單中刪除！", channel.Title))
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
			channelsList = append(channelsList, fmt.Sprintf("%-20s %s", channel.Title, fmt.Sprintf("https://discord.com/channels/%s/%s", guildId, channel.DiscordChannelId)))
		}

		if len(channelsList) == 0 {
			sendResponse(s, i, "沒有關注中的頻道！")
		} else {
			sendResponse(s, i, strings.Join(channelsList, "\n"))
		}
	},
	"help": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if r := recover(); r != nil {
				sendResponse(s, i, "似乎發生了什麼錯誤...")
				log.Println("handler panic:", r)
			}
		}()

		sendResponse(s, i, fmt.Sprintf("```/%-10s 新增關注頻道\n/%-10s 刪除關注頻道\n/%-10s 列出已關注頻道```", "add", "delete", "list"))
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

func IsContain(list []string, target string) bool {
	for _, element := range list {
		if element == target {
			return true
		}
	}

	return false
}

func checkChannelId(str string) (string, bool) {
	re := regexp.MustCompile(`UC[0-9A-Za-z-_]{22}`)

	match := re.FindString(str)

	if match != "" {
		return match, true
	}

	if strings.HasPrefix(str, "http") {
		req, err := http.NewRequest("GET", str, nil)
		if err != nil {
			return "", false
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", false
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", false
		}

		match := re.FindString(string(bodyBytes))

		if match != "" {
			return match, true
		}
	}

	return "", false
}
