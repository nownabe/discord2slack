package guild

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/bwmarrin/discordgo"
	"github.com/nownabe/discord2slack/pkg/worker/channel"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
)

const (
	GuildIDKey = "guildID"
)

type Handler struct {
	discordToken string
}

func New(discordToken string) *Handler {
	return &Handler{discordToken}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	discord, err := discordgo.New("Bot " + h.discordToken)
	if err != nil {
		msg := fmt.Sprintf("failed to get discord client: %v", err)
		log.Errorf(ctx, msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	discord.Client = urlfetch.Client(ctx)

	guildID := r.FormValue(GuildIDKey)

	guild, err := discord.Guild(guildID)
	if err != nil {
		msg := fmt.Sprintf("failed to get discord guild %s: %v", guildID, err)
		log.Errorf(ctx, msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	channels, err := discord.GuildChannels(guild.ID)
	if err != nil {
		msg := fmt.Sprintf("failed to get discord channels in %s: %v", guild.ID, err)
		log.Errorf(ctx, msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	for _, c := range channels {
		if c.Type == discordgo.ChannelTypeGuildText {
			v := url.Values{}
			v.Add(channel.GuildIDKey, guild.ID)
			v.Add(channel.GuildNameKey, guild.Name)
			v.Add(channel.ChannelIDKey, c.ID)
			v.Add(channel.ChannelNameKey, c.Name)
			t := taskqueue.NewPOSTTask("/worker/channel", v)
			if _, err := taskqueue.Add(ctx, t, ""); err != nil {
				log.Errorf(ctx,
					"[%s - %s] failed to enqueue channel check task: %v",
					guild.Name, c.Name, err)
			}
		}
	}

	fmt.Fprint(w, "ok")
}
