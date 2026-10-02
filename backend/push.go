package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

const (
	pushTTL     = 12 * 60 * 60 // seconds
	pushTimeout = 10 * time.Second
)

// sendFunc matches webpush.SendNotification, so tests can replace the push service.
type sendFunc func(message []byte, subscription *webpush.Subscription, options *webpush.Options) (*http.Response, error)

type pusher struct {
	app    core.App
	cfg    config
	send   sendFunc
	client *http.Client
}

// pushPayload is what the service worker receives; it stays well under the 4 KB push limit.
type pushPayload struct {
	Type      string `json:"type"`
	ListID    string `json:"listId"`
	ListTitle string `json:"listTitle"`
	ListDate  string `json:"listDate"`
	Item      string `json:"item"`
	Qty       string `json:"qty"`
	By        string `json:"by"`
}

// registerPush binds the item hook and the /api/push routes to the app.
func registerPush(app core.App, cfg config, send sendFunc) {
	p := &pusher{app: app, cfg: cfg, send: send, client: &http.Client{Timeout: pushTimeout}}

	app.OnRecordAfterCreateSuccess("items").BindFunc(func(e *core.RecordEvent) error {
		item := e.Record.Clone()
		// The response to the client does not wait for the push services.
		go func() {
			if err := p.notifyItemAdded(item); err != nil {
				app.Logger().Error("Push for a new item failed", "item", item.Id, "error", err)
			}
		}()
		return e.Next()
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		registerPushRoutes(se, p)
		return se.Next()
	})
}

// notifyItemAdded pushes the new item to the devices of everyone except its author.
func (p *pusher) notifyItemAdded(item *core.Record) error {
	list, err := p.app.FindRecordById("lists", item.GetString("list"))
	if err != nil {
		return err
	}
	author, err := p.app.FindRecordById("users", item.GetString("added_by"))
	if err != nil {
		return err
	}

	payload, err := json.Marshal(pushPayload{
		Type:      "item_added",
		ListID:    list.Id,
		ListTitle: list.GetString("title"),
		ListDate:  list.GetString("date"),
		Item:      item.GetString("name"),
		Qty:       item.GetString("qty"),
		By:        author.GetString("name"),
	})
	if err != nil {
		return err
	}

	subscriptions, err := p.app.FindAllRecords("push_subscriptions", dbx.Not(dbx.HashExp{"user": author.Id}))
	if err != nil {
		return err
	}
	for _, subscription := range subscriptions {
		p.deliver(subscription, payload, list.Id)
	}
	return nil
}

// deliver sends one push. A subscription the push service no longer knows is deleted; other
// errors are only logged, without retries.
func (p *pusher) deliver(subscription *core.Record, payload []byte, topic string) {
	response, err := p.send(payload, &webpush.Subscription{
		Endpoint: subscription.GetString("endpoint"),
		Keys:     webpush.Keys{P256dh: subscription.GetString("p256dh"), Auth: subscription.GetString("auth")},
	}, &webpush.Options{
		HTTPClient:      p.client,
		Subscriber:      p.cfg.vapidSubject,
		VAPIDPublicKey:  p.cfg.vapidPublicKey,
		VAPIDPrivateKey: p.cfg.vapidPrivateKey,
		TTL:             pushTTL,
		// High urgency: phones saving battery would otherwise hold the message back.
		Urgency: webpush.UrgencyHigh,
		// The push service replaces undelivered messages of the same list.
		Topic: topic,
	})
	if err != nil {
		p.app.Logger().Error("Push failed", "subscription", subscription.Id, "error", err)
		return
	}
	defer response.Body.Close()

	switch {
	case response.StatusCode == http.StatusNotFound || response.StatusCode == http.StatusGone:
		if err := p.app.Delete(subscription); err != nil {
			p.app.Logger().Error("Deleting an expired push subscription failed", "subscription", subscription.Id, "error", err)
		}
	case response.StatusCode >= 400:
		p.app.Logger().Error("Push rejected", "subscription", subscription.Id, "status", response.StatusCode)
	}
}
