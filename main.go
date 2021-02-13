package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var ( 
	Token string 
	FaceitAPIKey string
)

type FaceitResponse struct {
	Nickname string `json:"nickname,omitempty"`
	Games Games `json:"games,omitempty"`
}

type Games struct{
	Csgo Csgo `json:csgo,omitempty`
}

type Csgo struct{
	Skill_level int `json:skill_level,omitempty`
	Faceit_elo int `json:faceit_elo,omitempty`
}

func init(){

	if err := godotenv.Load(); err != nil {
        fmt.Println("No .env file found")
    }

	dbt, exists := os.LookupEnv("DISCORD_BOT_TOKEN")
	if exists {
		Token = dbt
	}

	fak, exists := os.LookupEnv("FACEIT_API_KEY")
	if exists {
		FaceitAPIKey = fak
	}
}

func main(){

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord bot",err)
		return
	}

	dg.AddHandler(messageCreate)

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate){

	if m.Author.ID == s.State.User.ID {
		return
	}
fmt.Println(m.Content)
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

	match, _ := regexp.MatchString("(!elo )[^ ]*", m.Content)
	fmt.Println(match)
	if match {
		nickname := m.Content[5:len(m.Content)]
		fmt.Println(nickname)
		elo := faceitElo(nickname)
		s.ChannelMessageSend(m.ChannelID, strconv.Itoa(elo))
	}
}

func faceitElo(nickname string) int{
	url := "https://open.faceit.com/data/v4/players?nickname="+nickname+"&game=csgo"
	bearer := "Bearer " + FaceitAPIKey

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", bearer)

	client := &http.Client{}
	
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("error", err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("error", err)
	}
	

	r := bytes.NewReader([]byte(body))
	decoder := json.NewDecoder(r)

	val := &FaceitResponse{}
	err = decoder.Decode(val)
	if err != nil {
        log.Fatal(err)
    }

	return val.Games.Csgo.Faceit_elo
}