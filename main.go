package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Define a struct that holds a bool and a Discord channel ID
type GameChannel struct {
	Flag       bool
	Channel    string
	JudgeID    string
	TargetID   string
	ImposterID string
	TargetDM   string
	ImposterDM string
	TargetName string
	TargetIs1  bool
}

func main() {
	fileName := "bottoken"

	// Open the file for reading
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// Read the entire file content into a byte slice
	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Convert byte slice to string
	bottoken := string(content)

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + bottoken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	ingame := GameChannel{
		Flag:       false,
		Channel:    "",
		JudgeID:    "",
		TargetID:   "",
		ImposterID: "",
		TargetDM:   "",
		ImposterDM: "",
		TargetName: "",
	}

	wrapper := func(s *discordgo.Session, m *discordgo.MessageCreate) {
		messageCreate(s, m, &ingame)
		//fmt.Printf("meow")
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(wrapper)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	//sc := make(chan os.Signal, 1)
	//signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	//<-sc

	// Cleanly close down the Discord session.
	//dg.Close()

	select {}

}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate, ingame *GameChannel) {

	fmt.Printf("%s: %s\n", m.Author.Username, m.Content)

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	keywordReplies(s, m)

	handleGameSetup(s, m, ingame)

	if ingame.Flag {
		playGame(s, m, ingame)
	}

}

func keywordReplies(s *discordgo.Session, m *discordgo.MessageCreate) {
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "pong!")
	}

	if strings.Contains(m.Content, fmt.Sprintf("<@!%s>", s.State.User.ID)) {
		if strings.Contains(m.Content, "ping") {
			s.ChannelMessageSend(m.ChannelID, "pong!")
		} else {
			s.ChannelMessageSend(m.ChannelID, "You rang?")
		}

	}

	if m.Content == "+help" {
		s.ChannelMessageSend(m.ChannelID, "Hello! So far, this bot only runs when I manually put it on. "+
			"I will update this message when this stops being the case!"+
			"Current commands: +help, ping, hey hal, tell me * responder, and the Turing Test Triangle game commands. "+
			"Type +gamehelp to find out more.")
	}

	if m.Content == "+gamehelp" {
		s.ChannelMessageSend(m.ChannelID, "Game: Turing Test Triangle (TTT). "+
			"The judge tries to guess which of 'target 1' and 'target 2' is the real target and which is the imposter. "+
			"To begin the game, find three players, who then register for each of the three roles "+
			"by typing +judge, +target, or +imposter."+
			"The target and the imposter must type to the DMs of the bot to anonymise who is who. "+
			"The judge takes a guess by typing +real 1 or +real 2.")
	}

	if strings.Contains(m.Content, "tell me") && strings.Contains(m.Content, "about") &&
		(strings.Contains(m.Content, " AR") || strings.Contains(m.Content, "responder")) {
		autoresponder := "It seems you have asked about DS's chat client auto-responder. " +
			"This is an application designed to simulate DS's otherwise inimitably rad typing style, " +
			"tone, cadence, personality, and substance of retort while he is away from the computer. " +
			"The algorithms are guaranteed to be 9" + fmt.Sprintf("%d", rand.Intn(10)) + "% " +
			"indistinguishable from DS's native neurological responses, " +
			"based on some statistical analysis I basically just pulled out of my ass right now."

		s.ChannelMessageSend(m.ChannelID, autoresponder)
	}

	if strings.Contains(strings.ToLower(m.Content), "hey hal") {
		s.ChannelMessageSend(m.ChannelID, "What's up, bro?")
	}
}

func handleGameSetup(s *discordgo.Session, m *discordgo.MessageCreate, ingame *GameChannel) {

	if ingame.Flag == true {
		if m.ChannelID != ingame.Channel {
			s.ChannelMessageSend(m.ChannelID, "Only set up for one game at a time currently! "+
				"Simultaneous multiple games might happen at some point. We will see.")
			return
		} else {
			s.ChannelMessageSend(m.ChannelID, "Use +exit to reset the current game!")
		}
	}

	// Define a slice with all the substrings to look for
	roleKeywords := []string{"judge", "target", "imposter"}

	role := containsAnyCommandKeywords(m.Content, roleKeywords)

	if role != "" {
		ingame.Channel = m.ChannelID
		switch role {
		case roleKeywords[0]:
			ingame.JudgeID = m.Author.ID
			s.ChannelMessageSend(ingame.Channel, "Player "+m.Author.Username+" joined game as the judge.")
		case roleKeywords[1]:
			ingame.TargetID = m.Author.ID
			ingame.TargetName = m.Author.Username
			s.ChannelMessageSend(ingame.Channel, "Player "+m.Author.Username+" joined game as the target.")
		case roleKeywords[2]:
			ingame.ImposterID = m.Author.ID
			s.ChannelMessageSend(ingame.Channel, "Player "+m.Author.Username+" joined game as the imposter.")
		}

		//fmt.Printf("j:%s t:%s i:%s\n", ingame.Judge, ingame.Target, ingame.Imposter)

	}

	if !ingame.Flag && ingame.JudgeID != "" && ingame.TargetID != "" && ingame.ImposterID != "" {
		s.ChannelMessageSend(ingame.Channel, "All participants assigned. Starting game now.")
		ingame.Flag = true
		dmUserGame(s, ingame, roleKeywords[1], "You are the target. Write your messages here in this DM.")
		dmUserGame(s, ingame, roleKeywords[2], "You are the imposter. Write your messages here in this DM.")
		ingame.TargetIs1 = rand.Intn(2) == 0
	}
}

func dmUserGame(s *discordgo.Session, ingame *GameChannel, role string, message string) {

	roleKeywords := []string{"judge", "target", "imposter"}

	switch role {
	case roleKeywords[1]:
		dmUser(s, ingame.TargetID, message)
		if ingame.TargetDM == "" {
			dmchan, err := s.UserChannelCreate(ingame.TargetID)
			if err != nil {
				return
			}
			ingame.TargetDM = dmchan.ID
		}
	case roleKeywords[2]:
		dmUser(s, ingame.ImposterID, message)
		if ingame.ImposterDM == "" {
			dmchan, err := s.UserChannelCreate(ingame.ImposterID)
			if err != nil {
				return
			}
			ingame.ImposterDM = dmchan.ID
		}
	}
}

func dmUser(s *discordgo.Session, userid string, message string) {
	dmchan, err := s.UserChannelCreate(userid)
	if err != nil {
		fmt.Println("error creating DM channel,", err)
		return
	}
	s.ChannelMessageSend(dmchan.ID, message)
}

func playGame(s *discordgo.Session, m *discordgo.MessageCreate, ingame *GameChannel) {

	roleKeywords := []string{"judge", "target", "imposter"}

	if containsAnyCommandKeywords(m.Content, roleKeywords) != "" {
		return
	}

	if m.ChannelID == ingame.Channel {
		dmUserGame(s, ingame, "target", m.Author.Username+": "+m.Content)
		dmUserGame(s, ingame, "imposter", m.Author.Username+": "+m.Content)
	}

	if m.ChannelID == ingame.TargetDM {
		nickname := ingame.TargetName
		if ingame.TargetIs1 {
			nickname = nickname + " 1"
		} else {
			nickname = nickname + " 2"
		}
		s.ChannelMessageSend(ingame.Channel, nickname+": "+m.Content)
		s.ChannelMessageSend(ingame.ImposterDM, nickname+": "+m.Content)
	}

	if m.ChannelID == ingame.ImposterDM {
		nickname := ingame.TargetName
		if ingame.TargetIs1 {
			nickname = nickname + " 2"
		} else {
			nickname = nickname + " 1"
		}
		s.ChannelMessageSend(ingame.Channel, nickname+": "+m.Content)
		s.ChannelMessageSend(ingame.TargetDM, nickname+": "+m.Content)
	}

	if m.Author.ID == ingame.JudgeID && strings.HasPrefix(m.Content, "+real") {
		if ingame.TargetIs1 {
			if strings.Contains(m.Content, "1") {
				s.ChannelMessageSend(ingame.Channel, "Correct! "+ingame.TargetName+" 1 was the real "+ingame.TargetName)
			}
			if strings.Contains(m.Content, "2") {
				s.ChannelMessageSend(ingame.Channel, "Wrong! "+ingame.TargetName+" 1 was the real "+ingame.TargetName)
			}
		} else {
			if strings.Contains(m.Content, "1") {
				s.ChannelMessageSend(ingame.Channel, "Wrong! "+ingame.TargetName+" 2 was the real "+ingame.TargetName)
			}
			if strings.Contains(m.Content, "2") {
				s.ChannelMessageSend(ingame.Channel, "Correct! "+ingame.TargetName+" 2 was the real "+ingame.TargetName)
			}
		}

		exitGame(ingame)

	}

	if m.Content == "+exit" {
		exitGame(ingame)
	}

}

func exitGame(ingame *GameChannel) {
	ingame.Flag = false
	ingame.Channel = ""
	ingame.JudgeID = ""
	ingame.TargetID = ""
	ingame.TargetName = ""
	ingame.TargetDM = ""
	ingame.ImposterID = ""
	ingame.ImposterDM = ""
}

func containsAnyCommandKeywords(content string, substrings []string) string {
	for _, s := range substrings {
		if strings.HasPrefix(content, "+"+s) {
			return s
		}
	}
	return ""
}
