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
				log.Println("1")
				member, err := s.GuildMember(m.GuildID, m.Author.ID)
				if err != nil {
					log.Println("2")
					s.ChannelMessageSend(m.ChannelID, "Something went wrong")
					log.Println("Error while getting guild member:", err.Error())
					return
				}
				isAuthorized := false
				for _, roleID := range member.Roles {
					log.Println("3")
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
					log.Println("4")
					s.ChannelMessageSend(m.ChannelID, "You are not authorized to do this action bozo")
					return
				}
				if partsLen != 3 {
					log.Println("5")
					s.ChannelMessageSend(
						m.ChannelID,
						"Incorrect usage of command 'create_role'\n\nUsage: @jerome create_role <role_name>",
					)
					return
				}
				log.Println("6")
				roleName := parts[2]
				_, err = s.GuildRoleCreate(m.GuildID, &discordgo.RoleParams{
					Name:        roleName,
					Color:       &defaultRolesColor,
					Hoist:       &defaultRolesHoist,
					Mentionable: &defaultRoleMentionable,
				})
				log.Println("7")
				if err != nil {
					log.Println("8")
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to create role %s. Cause: %s :)", roleName, err.Error()))
					return
				}
				log.Println("9")
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Role %s created :)", roleName))
				return
			}
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
		}
	})
	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
		fmt.Printf("%+v", m)
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
