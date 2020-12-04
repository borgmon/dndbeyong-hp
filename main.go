package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

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
	charList     = make(map[string]*Character)
	dataChan     = make(chan [][]string)
	charID       string
	interval     string
	campaignName string
)

const (
	characterAPIURL  = "https://character-service.dndbeyond.com/character/v4/character/"
	characterPageURL = "https://www.dndbeyond.com/characters/"
)

func main() {
	cliApp := getCLI()
	err := cliApp.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}

	writer := uilive.New()
	writer.Start()

	if len(os.Args) > 1 && strings.Contains(os.Args[1], "h") {
		return
	}

	if charID == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Fprint(writer, "What is your character ID: ")
		scanner.Scan()
		charID = scanner.Text()
		fmt.Fprintln(writer)
	}

	_, err = strconv.Atoi(charID)
	if err != nil {
		log.Fatalln("Invalid Character ID")
	}

	cronJob := cron.New()
	cronJob.Start()
	cronJob.AddFunc(interval, func() {
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

func getCLI() *cli.App {
	app := &cli.App{
		Name:                 "dndbeyond-hp",
		Usage:                "Get your characters hp from dndbeyond! Please set all characters to public.",
		UsageText:            "{Character ID} - Your dndbeyond character id",
		EnableBashCompletion: true,
		HideHelpCommand:      true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "interval",
				Aliases:     []string{"i"},
				Value:       "@every 1m",
				Usage:       "Set refresh interval. Not recommend to set lower as dndbeyond has DDOS protection. Example: https://godoc.org/github.com/robfig/cron",
				Destination: &interval,
			},
		},
		Action: func(c *cli.Context) error {
			_, err := cron.ParseStandard(interval)
			if err != nil {
				return err
			}
			if c.NArg() > 0 {
				charID = c.Args().Get(0)
			}
			return nil
		},
	}
	return app
}

func render(data [][]string, writer io.Writer) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"Name", "CurrentHP", "MaxHP"})
	table.SetFooter([]string{"Campaign Name", "", campaignName})
	table.AppendBulk(data)
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

	campaignName = charPayload.Data.Campaign.Name

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
