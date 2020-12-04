package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
	"github.com/gosuri/uilive"
	"github.com/olekukonko/tablewriter"
	"github.com/robfig/cron"
)

type Character struct {
	ID    string
	Name  string
	MaxHP string
	CurHP string
}

type CharacterAPIPayload struct {
	Data CharacterAPIPayloadData
}
type CharacterAPIPayloadData struct {
	Name     string
	Campaign CharacterAPIPayloadCampaign
}
type CharacterAPIPayloadCampaign struct {
	ID         int
	Name       string
	Characters []CharacterAPIPayloadCampaignCharacter
}
type CharacterAPIPayloadCampaignCharacter struct {
	CharacterID   int
	CharacterName string
}

var (
	charList = make(map[string]*Character)
	dataChan = make(chan [][]string)
)

const (
	characterAPIURL  = "https://character-service.dndbeyond.com/character/v4/character/"
	characterPageURL = "https://www.dndbeyond.com/characters/"
)

func main() {
	writer := uilive.New()
	writer.Start()

	cronJob := cron.New()
	cronJob.Start()
	cronJob.AddFunc("@every 30s", func() {
		go start()
	})
	go start()
	for {
		data := <-dataChan
		render(data, writer)
	}
}

func render(data [][]string, writer io.Writer) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"Name", "CurrentHP", "MaxHP"})

	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}

func start() {
	charID := os.Args[1]
	char := Character{ID: charID}
	charList[charID] = &char

	charPayload, err := getCharAPI(&char)
	if err != nil {
		log.Fatalln(err)
	}

	for i := range charPayload.Data.Campaign.Characters {
		v := charPayload.Data.Campaign.Characters[i]
		id := strconv.Itoa(v.CharacterID)
		newChar := Character{ID: id, Name: v.CharacterName}
		charList[id] = &newChar
	}

	var data [][]string

	for i := range charList {
		v := charList[i]
		err := getInfoFromBrowser(v)
		if err != nil {
			log.Fatalln(err)
		}
		data = append(data, []string{v.Name, v.CurHP, v.MaxHP})
	}
	dataChan <- data
}

func getCharAPI(char *Character) (charPayload *CharacterAPIPayload, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", characterAPIURL+char.ID, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&charPayload)
	if err != nil {
		return nil, err
	}
	return
}

func getInfoFromBrowser(char *Character) (err error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	err = chromedp.Run(ctx,
		chromedp.Emulate(device.IPhone7landscape),
		chromedp.Navigate(characterPageURL+char.ID),
		chromedp.Text(`.ct-status-summary-mobile__hp-current`, &char.CurHP, chromedp.NodeVisible, chromedp.ByQuery),
		chromedp.Text(`.ct-status-summary-mobile__hp-max`, &char.MaxHP, chromedp.NodeVisible, chromedp.ByQuery),
	)
	return
}
