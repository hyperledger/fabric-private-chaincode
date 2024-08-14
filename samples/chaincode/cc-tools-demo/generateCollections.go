package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type ArrayFlags []string

func (i *ArrayFlags) String() string {
	return "my string representation"
}

func (i *ArrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type CollectionElem struct {
	Name              string `json:"name"`
	RequiredPeerCount int    `json:"requiredPeerCount"`
	MaxPeerCount      int    `json:"maxPeerCount"`
	BlockToLive       int    `json:"blockToLive"`
	MemberOnlyRead    bool   `json:"memberOnlyRead"`
	Policy            string `json:"policy"`
}

func generateCollection(orgs ArrayFlags) {
	collection := []CollectionElem{}

	for _, a := range assetTypeList {
		if len(a.Readers) > 0 {
			elem := CollectionElem{
				Name:              a.Tag,
				RequiredPeerCount: 0,
				MaxPeerCount:      3,
				BlockToLive:       1000000,
				MemberOnlyRead:    true,
				Policy:            generatePolicy(a.Readers, orgs),
			}
			collection = append(collection, elem)
		}
	}

	b, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = os.WriteFile("collections.json", b, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func generatePolicy(readers []string, orgs ArrayFlags) string {
	firstElem := true
	policy := "OR("
	for _, r := range readers {
		if len(orgs) > 0 {
			found := false
			for _, o := range orgs {
				if r == o {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if !firstElem {
			policy += ", "
		}
		policy += fmt.Sprintf("'%s.member'", r)
		firstElem = false
	}
	policy += ")"
	return policy
}
