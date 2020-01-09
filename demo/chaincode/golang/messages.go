/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package golang

import (
	"encoding/json"
)

// Messages
type AuctionStatus struct {
	State       string `json:"state"` // "clock" | "assign" | "done" | "failed_fsr"
	ClockRound  int    `json:"clockRound"`
	RoundActive bool   `json:"roundActive"`
}

type Status struct {
	Rc  int    `json:"rc"`
	Msg string `json:"message"`
}

type ReturnMsg struct {
	Status   *Status     `json:"status"`
	Response interface{} `json:"response"`
}

func buildReturnJson(status *Status, response interface{}) []byte {
	msg := ReturnMsg{
		Status:   status,
		Response: response,
	}

	payload, _ := json.Marshal(msg)
	return payload
}

type RoundRequest struct {
	AuctionId int `json:"auctionId"`
	Round     int `json:"round"`
}

type RoundInfo struct {
	Prices []Price `json:"prices"`
	Active bool    `json:"active"`
}

type Price struct {
	TerritoryId int `json:"terId"`
	MinPrice    int `json:"minPrice"`
	ClockPrice  int `json:"clockPrice"`
}

type BidderRoundResult struct {
	Result                 []Result `json:"result"`
	FutureEligibility      int      `json:"futureEligibility"`
	RequiredFutureActivity int      `json:"requiredFutureActivity"`
}

type Result struct {
	TerritoryId       int `json:"terId"`
	PostedPrice       int `json:"postedPrice"`
	ExcessDemand      int `json:"excessDemand"`
	ProcessedLicenses int `json:"processedLicenses"`
}

type AuctionRequest struct {
	AuctionId int `json:"auctionId"`
}

// State

type Auction struct {
	Owner                         *Principal    `json:"owner"`
	Name                          string        `json:"name"`
	Territories                   []Territory   `json:"territories"`
	Bidders                       []Bidder      `json:"bidders"`
	InitialEligibilities          []Eligibility `json:"initialEligibilities"`
	ActivityRequirementPercentage int           `json:"activityRequirementPercentage"`
	ClockPriceIncrementPercentage int           `json:"clockPriceIncrementPercentage"`
}

type Principal struct {
	Mspid string `json:"mspId"`
	Dn    string `json:"dn"`
}

type Territory struct {
	Id           int       `json:"id"`
	Name         string    `json:"name"`
	IsHighDemand bool      `json:"isHighDemand"`
	MinPrice     int       `json:"minPrice"`
	Channels     []Channel `json:"channels"`
}

type Eligibility struct {
	BidderId int `json:"bidderId"`
	Number   int `json:"number"`
}

type Bidder struct {
	Id          int        `json:"id"`
	DisplayName string     `json:"displayName"`
	Principal   *Principal `json:"principal"`
}

type Channel struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Impairment int    `json:"impairment"`
}
