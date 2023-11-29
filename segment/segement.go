package segment

import (
	"github.com/segmentio/analytics-go/v3"
	"log"
	"os"
)

var client analytics.Client = nil

func InitClient() analytics.Client {
	segmentApiKey := os.Getenv("SEGMENT_API_KEY")
	if segmentApiKey == "" {
		log.Printf("Not initializing segment because SEGMENT_API_KEY is missing")
		return nil
	}
	if client == nil {
		client = analytics.New("YOUR_WRITE_KEY")
	}
	return client
}

func CloseClient() {
	if client == nil {
		return
	}
	client.Close()
}

func IdentifyClient(userId string, userFullName string, username string, email string, organisationName string, organisationId string, userPlan string) {
	if client == nil {
		return
	}
	client.Enqueue(analytics.Identify{
		UserId: userId,
		Traits: analytics.NewTraits().
			SetName(userFullName).
			SetUsername(username).
			SetEmail(email).
			Set("organisationName", organisationName).
			Set("organisationId", organisationId).
			Set("plan", userPlan),
	})
}

func Track(userId string, action string) {
	if client == nil {
		return
	}
	client.Enqueue(analytics.Track{
		Event:  action,
		UserId: userId,
	})
}