package firebase_notifications

import (
	"sync"
	"log"
	"firebase.google.com/go"
	"google.golang.org/api/option"
	"context"
	"firebase.google.com/go/messaging"
	"os"
)

type NotificationData struct {
	Title       string
	Body        string
	Icon        string
	Topic       string
	Id          string
}


var (
	notificationQueue chan *messaging.Message
	once sync.Once
)


func Start() error {
	var err error

	opt := option.WithCredentialsJSON([]byte(os.Getenv("firebase_credentials")))

	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatal(err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatal(err)
	}

	queue := getQueue()

	for message := range queue {
		log.Println("sending..")
		_, err = client.Send(ctx, message)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("message has been sent")
	}

	return nil
}


func getQueue() chan *messaging.Message {
	once.Do(func() {
		notificationQueue = make(chan *messaging.Message, 0)
	})
	return notificationQueue
}


func (notification *NotificationData) Send() error {
	message := &messaging.Message{
		/*Webpush: &messaging.WebpushConfig{
			Notification: &messaging.WebpushNotification{
				Title: notification.Title,
				Body:  notification.Body,
				CustomData: notification.Data,
				Actions: []*messaging.WebpushNotificationAction {
					{
						Title: "test",
						Action: "test",
					},
				},
			},
		},*/
		Data: map[string]string {
			"title": notification.Title,
			"body": notification.Body,
			"id": notification.Id,
		},

		Topic: notification.Topic,
	}

	getQueue() <- message
	return nil
}