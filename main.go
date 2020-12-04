package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

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
	charID   string
)

const (
	characterAPIURL  = "https://character-service.dndbeyond.com/character/v4/character/"
	characterPageURL = "https://www.dndbeyond.com/characters/"
)

func main() {
	charID = os.Args[1]

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
		sortByName(data)
		writer.Flush()
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
	char := Character{ID: charID}
	charList[charID] = &char
	ctx, cancel := chromedp.NewContext(context.Background())
	browserChan := make(chan func() (*Character, error))

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

	for k := range charList {
		v := charList[k]
		go getInfoFromBrowser(ctx, v, browserChan)
	}

	for i := 0; i < len(charList); i++ {
		char, err := (<-browserChan)()
		if err != nil {
			log.Fatalln(err)
		}
		charList[char.ID] = char
		data = append(data, []string{char.Name, char.CurHP, char.MaxHP})
	}

	defer cancel()
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

func getInfoFromBrowser(ctx context.Context, char *Character, ch chan func() (*Character, error)) {
	ctx, cancel := chromedp.NewContext(ctx)
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	err := chromedp.Run(ctx,
		chromedp.Emulate(device.IPhone7landscape),
		chromedp.Navigate(characterPageURL+char.ID),
		chromedp.Text(`.ct-status-summary-mobile__hp-current`, &char.CurHP, chromedp.NodeVisible, chromedp.ByQuery),
		chromedp.Text(`.ct-status-summary-mobile__hp-max`, &char.MaxHP, chromedp.NodeVisible, chromedp.ByQuery),
	)
	defer cancel()
	ch <- (func() (*Character, error) { return char, err })
}

func sortByName(ls [][]string) {
	sort.Slice(ls, func(i, j int) bool {
		return ls[i][0] < ls[j][0]
	})
}
