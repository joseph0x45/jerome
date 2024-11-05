package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var defaultRolesColor int = 0x00ff00
var defaultRolesHoist bool = false
var defaultRoleMentionable bool = true

const prefix = "!jerome"

const rolesMessageID = "1303494297357783092"
const rolesChannelID = "1302744612049387540"

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	token := os.Getenv("DISCORD_BOT_TOKEN")
	session, err := discordgo.New(fmt.Sprintf("Bot %s", token))
	if err != nil {
		panic(err)
	}
	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}
		parts := strings.Split(m.Content, " ")
		if parts[0] != prefix {
			return
		}
		partsLen := len(parts)
		if partsLen >= 2 {
			cmd := parts[1]
			if cmd == "create_role" {
				member, err := s.GuildMember(m.GuildID, m.Author.ID)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Something went wrong")
					log.Println("Error while getting guild member:", err.Error())
					return
				}
				isAuthorized := false
				for _, roleID := range member.Roles {
					role, err := s.State.Role(m.GuildID, roleID)
					if err != nil {
						s.ChannelMessageSend(m.ChannelID, "Something went wrong")
						log.Println("Error while getting role informations:", err.Error())
						return
					}
					if role.Name == "admin" {
						isAuthorized = true
						break
					}
				}
				if !isAuthorized {
					s.ChannelMessageSend(m.ChannelID, "You are not authorized to do this action bozo")
					return
				}
				if partsLen != 3 {
					s.ChannelMessageSend(
						m.ChannelID,
						"Incorrect usage of command 'create_role'\n\nUsage: @jerome create_role <role_name>",
					)
					return
				}
				roleName := parts[2]
				_, err = s.GuildRoleCreate(m.GuildID, &discordgo.RoleParams{
					Name:        roleName,
					Color:       &defaultRolesColor,
					Hoist:       &defaultRolesHoist,
					Mentionable: &defaultRoleMentionable,
				})
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to create role %s. Cause: %s :)", roleName, err.Error()))
					return
				}
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Role %s created :)", roleName))
				return
			}
			// List all roles
			if cmd == "list_roles" {
				roles, err := s.GuildRoles(m.GuildID)
				if err != nil {
					log.Println("Error while getting all roles:", err.Error())
					s.ChannelMessageSend(m.ChannelID, "Something went wrong")
					return
				}
				message := "**Roles in this server**\n"
				for _, role := range roles {
					if role.Name == "@everyone" {
						continue
					}
					message += fmt.Sprintf("- %s\n", role.Name)
				}
				s.ChannelMessageSend(m.ChannelID, message)
				return
			}
			//Setup the reactions
			if cmd == "setup_reactions" {
				message, err := s.ChannelMessage(rolesChannelID, rolesMessageID)
				if err != nil {
					log.Println("Error while getting message:", err.Error())
					s.ChannelMessageSend(m.ChannelID, "Something went wrong")
					return
				}
				reactionsMap := map[string]bool{}
				for _, reaction := range message.Reactions {
					reactionsMap[reaction.Emoji.ID] = true
				}
				emojis, err := s.GuildEmojis(m.GuildID)
				if err != nil {
					log.Println("Error while getting all custom emojis:", err.Error())
					s.ChannelMessageSend(m.ChannelID, "Something went wrong")
					return
				}
				for _, emoji := range emojis {
					if reactionsMap[emoji.ID] {
						log.Println("Already reacted with this emoji. Skipping")
						continue
					}
					err = s.MessageReactionAdd(rolesChannelID, rolesMessageID, fmt.Sprintf("%s:%s", emoji.Name, emoji.ID))
					if err != nil {
						log.Println("Error while reacting to message:", err.Error())
					}
				}
				return
			}
		}
	})

	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
		roles, err := s.GuildRoles(m.GuildID)
		if err != nil {
			log.Println("Error while getting all roles:", err.Error())
			s.ChannelMessageSend(m.ChannelID, "Something went wrong")
			return
		}
		roleID := ""
		for _, role := range roles {
			if role.Name == m.Emoji.Name {
				log.Println("Found role:", m.Emoji.Name)
				roleID = role.ID
				break
			}
		}
		if roleID == "" {
			log.Println("Role ", m.Emoji.Name, "not found")
			s.ChannelMessageSend(m.ChannelID, "Something went wrong")
			return
		}
		err = s.GuildMemberRoleAdd(m.GuildID, m.UserID, roleID)
		if err != nil {
			log.Println("Error while adding role to user:", err.Error())
			s.ChannelMessageSend(m.ChannelID, "Something went wrong")
			return
		}
	})

	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageReactionRemove) {
		roles, err := s.GuildRoles(m.GuildID)
		if err != nil {
			log.Println("Error while getting all roles:", err.Error())
			s.ChannelMessageSend(m.ChannelID, "Something went wrong")
			return
		}
		roleID := ""
		for _, role := range roles {
			if role.Name == m.Emoji.Name {
				log.Println("Found role:", m.Emoji.Name)
				roleID = role.ID
				break
			}
		}
		if roleID == "" {
			log.Println("Role ", m.Emoji.Name, "not found")
			s.ChannelMessageSend(m.ChannelID, "Something went wrong")
			return
		}
		err = s.GuildMemberRoleRemove(m.GuildID, m.UserID, roleID)
		if err != nil {
			log.Println("Error while removing role from user:", err.Error())
			s.ChannelMessageSend(m.ChannelID, "Something went wrong")
			return
		}
	})

	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged
	err = session.Open()
	if err != nil {
		panic(err)
	}
	defer session.Close()
	log.Println("Bot online")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop
}
