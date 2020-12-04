package main

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
