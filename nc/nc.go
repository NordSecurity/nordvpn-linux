/*
Package nc provides a MQTT client to connect to the Notification Centre server.
MQTT could be viewed as an application layer version of TCP since it uses similar
mechanisms to ensure message delivery.
*/
package nc

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/network"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	typeMeshnet = "meshnet_network_update"

	topicDelivered    = "delivered"
	topicAcknowledged = "ack"

	trackTypeProcessed = "processed"
)

const (
	logPrefix = "[NC]"
	timeout   = 5 * time.Second
)

var subscriptions = map[string]byte{
	"content": byte(1),
	"linux":   byte(1),
	"meshnet": byte(1),
}

var unsubscriptions = []string{"content", "linux", "meshnet"}

// RecPayload defines a payload sent by a NC
type RecPayload struct {
	Message MessagePayload `json:"message"`
}

type MessagePayload struct {
	Data         DataPayload         `json:"data"`
	Notification NotificationPayload `json:"notification"`
}

type DataPayload struct {
	Event    EventPayload    `json:"event"`
	Metadata MetadataPayload `json:"metadata"`
}

type NotificationPayload struct {
	Title   string          `json:"title"`
	Body    string          `json:"body"`
	Actions []ActionPayload `json:"actions"`
}

type EventPayload struct {
	Type       string           `json:"type"`
	Attributes AttributePayload `json:"attributes"`
}

type ActionPayload struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Slug string `json:"slug"`
}

type AttributePayload struct {
	InviteeID        string   `json:"invitee_id"`
	ServerID         int      `json:"server_id"`
	ServerIP         string   `json:"server_ip"`
	AffectedMachines []string `json:"affected_machines"`
}

type MetadataPayload struct {
	Acked     bool   `json:"acked"`
	CreatedAt int    `json:"created_at"`
	MessageID string `json:"message_id"`
	TargetUID string `json:"target_uid"`
}

type ConfirmationPayload struct {
	MessageID string `json:"message_id"`
}

type AcknowledgementPayload struct {
	MessageID  string `json:"message_id"`
	TrackType  string `json:"track_type"`
	ActionSlug string `json:"action_slug"`
}

type NotificationClient interface {
	Start(endpoint string, clientID string, username string, password string) error
	Stop() error
}

// Client is a client for Notification center
type Client struct {
	client            mqtt.Client
	subjectDebug      events.Publisher[string]
	subjectPeerUpdate events.Publisher[[]string]
	sync.Mutex
}

// NewClient is a constructor for a NC client
func NewClient(subjectDebug events.Publisher[string], subjectPeerUpdate events.Publisher[[]string]) *Client {
	return &Client{
		subjectDebug:      subjectDebug,
		subjectPeerUpdate: subjectPeerUpdate,
	}
}

// Start initiates the connection with the NC server and subscribes to mandatory topics
func (c *Client) Start(endpoint string, clientID string, username string, password string) error {
	opts := mqtt.NewClientOptions()
	opts.SetCleanSession(false)
	opts.SetOrderMatters(false)
	opts.SetAutoReconnect(false) // handled manually with the exponential backoff
	opts.AddBroker(endpoint)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetDefaultPublishHandler(c.receiveMessage)
	opts.SetClientID(clientID)
	opts.SetOnConnectHandler(c.onConnect)
	opts.SetConnectionLostHandler(c.onConnectionLost)
	client := mqtt.NewClient(opts)

	if err := c.startWithExponentialBackoff(client); err != nil {
		return err
	}
	c.subjectDebug.Publish(fmt.Sprintf("%s Client has started", logPrefix))
	return nil
}

func (c *Client) start(client mqtt.Client) error {
	if token := client.Connect(); token.WaitTimeout(timeout) && token.Error() != nil {
		return fmt.Errorf("connecting to notification centre: %w", token.Error())
	}
	c.Lock()
	c.client = client
	c.Unlock()
	return nil
}

func (c *Client) startWithExponentialBackoff(client mqtt.Client) error {
	var err error
	for tries := 0; !client.IsConnected(); tries++ {
		if err = c.start(client); err == nil {
			break
		}
		<-time.After(network.ExponentialBackoff(tries))
	}
	return err
}

// Stop unsubscribes the topics and drops a connection with the NC server
func (c *Client) Stop() error {
	c.Lock()
	defer c.Unlock()

	if c.client != nil {
		token := c.client.Unsubscribe(unsubscriptions...)
		if token.WaitTimeout(timeout) && token.Error() != nil {
			return fmt.Errorf(`unsubscribing to %v topics: %w`, unsubscriptions, token.Error())
		}
		c.client.Disconnect(0)
		c.subjectDebug.Publish("[NC] Client has stopped")
	}
	return nil
}

func (c *Client) onConnect(client mqtt.Client) {
	token := client.SubscribeMultiple(subscriptions, nil)
	if token.WaitTimeout(timeout) && token.Error() != nil {
		c.subjectDebug.Publish(
			fmt.Sprintf("subscribing to %v topics: %s", unsubscriptions, token.Error()),
		)
	}
}

func (c *Client) onConnectionLost(client mqtt.Client, err error) {
	c.subjectDebug.Publish(fmt.Sprintf("%s connection lost: %s", logPrefix, err.Error()))
	if err := c.startWithExponentialBackoff(client); err != nil {
		c.subjectDebug.Publish(fmt.Sprintf("%s reconnecting: %s", logPrefix, err.Error()))
	} else {
		c.subjectDebug.Publish(fmt.Sprintf("%s reconnected", logPrefix))
	}
}

func (c *Client) receiveMessage(client mqtt.Client, msg mqtt.Message) {
	var payload RecPayload
	if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
		c.subjectDebug.Publish(
			fmt.Sprintf("%s parsing message payload: %s", logPrefix, err),
		)
		return
	}

	metadata := payload.Message.Data.Metadata
	if opts := c.client.OptionsReader(); metadata.TargetUID != opts.ClientID() {
		c.subjectDebug.Publish(
			fmt.Sprintf("%s attempted to publish message to incorrect recipient", logPrefix),
		)
		return
	}

	if metadata.Acked {
		c.subjectDebug.Publish(
			fmt.Sprintf("%s message was already published successfully", logPrefix),
		)
		return
	}

	if err := c.sendDeliveryConfirmation(metadata.MessageID); err != nil {
		c.subjectDebug.Publish(
			fmt.Sprintf("%s Delivery confirmation: %v", logPrefix, err),
		)
	}

	if err := c.sendAcknowledgement(metadata.MessageID, trackTypeProcessed, ""); err != nil {
		c.subjectDebug.Publish(
			fmt.Sprintf("%s Acknowledgement: %v", logPrefix, err),
		)
	}

	if payload.Message.Data.Event.Type == typeMeshnet {
		c.subjectPeerUpdate.Publish(payload.Message.Data.Event.Attributes.AffectedMachines)
	}
}

func (c *Client) sendDeliveryConfirmation(messageID string) error {
	payload, err := json.Marshal(ConfirmationPayload{
		MessageID: messageID,
	})
	if err != nil {
		return fmt.Errorf("marshaling confirmation payload: %w", err)
	}
	if token := c.client.Publish(topicDelivered, 1, false, payload); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (c *Client) sendAcknowledgement(messageID, trackType, actionSlug string) error {
	payload, err := json.Marshal(AcknowledgementPayload{
		MessageID:  messageID,
		TrackType:  trackType,
		ActionSlug: actionSlug,
	})
	if err != nil {
		return fmt.Errorf("marshaling acknowledgement payload: %w", err)
	}
	if token := c.client.Publish(topicAcknowledged, 1, false, payload); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
