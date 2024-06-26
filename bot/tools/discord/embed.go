package discord

import (
	"GoBot/tools"
	"GoBot/tools/youtube"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Embed discordgo.MessageEmbed

func BaseEmbed(class, name, url, icon string) *Embed {
	embed := &Embed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    name,
			URL:     url,
			IconURL: icon,
		},
	}

	var color int
	var footerIcon string

	switch class {
	case "Youtube":
		color = 0xff0000
		footerIcon = "https://imgur.com/8Ne5sku.png"
	case "Twitch":
		color = 0x9b00ff
		footerIcon = "https://imgur.com/FcA3VwK.png"
	case "Twitcasting":
		color = 0x28a0ff
		footerIcon = "https://imgur.com/KOPaI0A.png"
	case "Tiktok":
		color = 0x000000
		footerIcon = "https://imgur.com/etYPtfz.png"
	case "Fanbox":
		color = 0xFFFFFF
		footerIcon = "https://imgur.com/LMrQB1e.png"
	case "Twitter":
		color = 0x46A3FF
		footerIcon = "https://imgur.com/7F0vjj4.png"
	}

	embed.Color = color
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text:    class,
		IconURL: footerIcon,
	}

	return embed
}

func (embed *Embed) New(title, url, description, image string) *Embed {
	return &Embed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    embed.Author.Name,
			URL:     embed.Author.URL,
			IconURL: embed.Author.IconURL,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    embed.Footer.Text,
			IconURL: embed.Footer.IconURL,
		},
		Image: &discordgo.MessageEmbedImage{
			URL: image,
		},
		Title:       title,
		URL:         url,
		Description: description,
		Color:       embed.Color,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}
}

func (embed *Embed) Send(s *discordgo.Session, discordChannelId string) {
	s.ChannelMessageSendEmbed(discordChannelId, (*discordgo.MessageEmbed)(embed))
}

func (embed *Embed) Change(before, after string) *Embed {
	embed.addField("【變更前】", before)
	embed.addField("【變更後】", after)

	return embed
}

func (embed *Embed) UpcomingTime(time tools.Time) *Embed {
	embed.addField("直播預定時間", fmt.Sprintf("%s (%s)", time.String("full"), time.String("relative")))

	return embed
}

func (embed *Embed) StartTime(time tools.Time) *Embed {
	embed.addField("直播開始時間", fmt.Sprintf("%s (%s)", time.String("full"), time.String("relative")))

	return embed
}

func (embed *Embed) EndTime(time tools.Time, duration tools.Duration) *Embed {
	embed.addField("直播結束時間", time.String("full"))
	embed.addField("直播總時長", duration.String("full"))

	return embed
}

func (embed *Embed) NewNotify(video youtube.Video) *Embed {
	if !video.Live && video.LiveStatus == 0 {
		embed = embed.New(video.Title, video.Url, "上傳了新影片！", video.Thumbnail)
	} else if video.LiveStatus == 1 {
		embed = embed.New(video.Title, video.Url, "建立了新的待機台！", video.Thumbnail).UpcomingTime(video.ScheduledTime)
	} else if video.LiveStatus == 2 {
		embed = embed.New(video.Title, video.Url, "直播串流開始了！", video.Thumbnail).StartTime(video.StartTime)
	}

	return embed
}

func (embed *Embed) addField(name, value string) {
	field := discordgo.MessageEmbedField{
		Name:   name,
		Value:  trimString(value),
		Inline: true,
	}
	embed.Fields = append(embed.Fields, &field)
}

func trimString(s string) string {
	if len(s) > 750 {
		return s[:750] + "..."
	}

	return s
}
