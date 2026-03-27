package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/lib/pq"
	wailsEvents "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	chequesChangedChannel = "cheques_changed"
	dbUpdatedEventName    = "db-updated"
	listenerDebounce      = 750 * time.Millisecond
	fallbackInterval      = 15 * time.Second
)

type dbUpdatedPayload struct {
	Source    string `json:"source"`
	Table     string `json:"table"`
	Operation string `json:"operation,omitempty"`
	ID        int    `json:"id,omitempty"`
	At        string `json:"at,omitempty"`
	Raw       string `json:"raw,omitempty"`
}

func (a *App) startDBSyncListener(dsn string) {
	if a.ctx == nil || dsn == "" || a.listenerStop != nil {
		return
	}

	listener := pq.NewListener(
		dsn,
		5*time.Second,
		time.Minute,
		func(event pq.ListenerEventType, err error) {
			if err != nil {
				log.Printf("pq listener event %v: %v", event, err)
			}
		},
	)

	if err := listener.Listen(chequesChangedChannel); err != nil {
		log.Printf("Warning: could not listen on channel %s: %v", chequesChangedChannel, err)
		_ = listener.Close()
		return
	}

	stop := make(chan struct{})
	done := make(chan struct{})

	a.pgListener = listener
	a.listenerStop = stop
	a.listenerDone = done

	go a.runDBSyncLoop(listener, stop, done)
}

func (a *App) stopDBSyncListener() {
	stop := a.listenerStop
	done := a.listenerDone
	listener := a.pgListener

	a.listenerStop = nil
	a.listenerDone = nil
	a.pgListener = nil

	if stop != nil {
		close(stop)
	}

	if listener != nil {
		if err := listener.UnlistenAll(); err != nil {
			log.Printf("Warning: could not unlisten channels: %v", err)
		}
		if err := listener.Close(); err != nil {
			log.Printf("Warning: could not close pq listener: %v", err)
		}
	}

	if done != nil {
		<-done
	}
}

func (a *App) runDBSyncLoop(listener *pq.Listener, stop <-chan struct{}, done chan<- struct{}) {
	defer close(done)

	fallbackTicker := time.NewTicker(fallbackInterval)
	defer fallbackTicker.Stop()

	debounceTimer := time.NewTimer(listenerDebounce)
	if !debounceTimer.Stop() {
		select {
		case <-debounceTimer.C:
		default:
		}
	}
	defer func() {
		if !debounceTimer.Stop() {
			select {
			case <-debounceTimer.C:
			default:
			}
		}
	}()

	debounceActive := false
	pendingPayload := dbUpdatedPayload{Source: "fallback-poll", Table: "cheques"}

	scheduleEmit := func(payload dbUpdatedPayload) {
		pendingPayload = payload

		if debounceActive {
			if !debounceTimer.Stop() {
				select {
				case <-debounceTimer.C:
				default:
				}
			}
		}
		debounceTimer.Reset(listenerDebounce)
		debounceActive = true
	}

	emit := func() {
		if a.ctx == nil {
			return
		}
		wailsEvents.EventsEmit(a.ctx, dbUpdatedEventName, pendingPayload)
	}

	for {
		select {
		case <-stop:
			return
		case <-fallbackTicker.C:
			scheduleEmit(dbUpdatedPayload{Source: "fallback-poll", Table: "cheques"})
		case notification, ok := <-listener.Notify:
			if !ok {
				return
			}

			if notification == nil || notification.Channel == chequesChangedChannel {
				scheduleEmit(parseNotificationPayload(notification))
			}
		case <-debounceTimer.C:
			debounceActive = false
			emit()
		}
	}
}

func parseNotificationPayload(notification *pq.Notification) dbUpdatedPayload {
	payload := dbUpdatedPayload{
		Source: "listen-notify",
		Table:  "cheques",
	}

	if notification == nil || notification.Extra == "" {
		return payload
	}

	var dbPayload struct {
		Table     string `json:"table"`
		Operation string `json:"operation"`
		ID        int    `json:"id"`
		At        string `json:"at"`
	}

	if err := json.Unmarshal([]byte(notification.Extra), &dbPayload); err != nil {
		payload.Raw = notification.Extra
		return payload
	}

	if dbPayload.Table != "" {
		payload.Table = dbPayload.Table
	}
	payload.Operation = dbPayload.Operation
	payload.ID = dbPayload.ID
	payload.At = dbPayload.At

	return payload
}
