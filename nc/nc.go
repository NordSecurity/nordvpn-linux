/*
Package nc provides a MQTT client to connect to the Notification Centre server.
MQTT could be viewed as an application layer version of TCP since it uses similar
mechanisms to ensure message delivery.
*/
package nc

import (
	"context"
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
	client mqtt.Client // nil if not started or stopped, always check for nil before using
	// MQTT Docs say that reusing client after doing Disconnect can lead to panics.
	// Since we are doing connect manually with our exponential backoff, we are in risk of those panics.
	// That's why this mutex must be locked everytime client is used.
	clientMutex       sync.Mutex
	subjectInfo       events.Publisher[string]
	subjectErr        events.Publisher[error]
	subjectPeerUpdate events.Publisher[[]string]
	cancelConnecting  context.CancelFunc // Used to stop connecting attempts if we are already stopping
}

// NewClient is a constructor for a NC client
func NewClient(
	subjectInfo events.Publisher[string],
	subjectErr events.Publisher[error],
	subjectPeerUpdate events.Publisher[[]string],
) *Client {
	return &Client{
		subjectInfo:       subjectInfo,
		subjectErr:        subjectErr,
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

	if err := c.newClient(opts); err != nil {
		return err
	}
	c.subjectInfo.Publish(fmt.Sprintf("%s Client has started", logPrefix))

	var ctx context.Context
	ctx, c.cancelConnecting = context.WithCancel(context.Background())
	c.connectWithBackoff(ctx, network.ExponentialBackoff)
	return nil
}

func (c *Client) newClient(opts *mqtt.ClientOptions) error {
	c.clientMutex.Lock()
	defer c.clientMutex.Unlock()
	if c.client != nil {
		return fmt.Errorf("already started")
	}

	c.client = mqtt.NewClient(opts)
	return nil
}

func (c *Client) connectWithBackoff(ctx context.Context, backoff func(int) time.Duration) {
	for tries := 0; ; tries++ {
		err := c.connect()
		if err == nil {
			break
		}
		if tries == 0 {
			// Don't spam the logs, only print the first time
			c.subjectErr.Publish(fmt.Errorf("%s failed to connect: %w", logPrefix, err))
		}

		select {
		case <-ctx.Done():
			break
		case <-time.After(backoff(tries)):
		}
	}
}

func (c *Client) connect() error {
	c.clientMutex.Lock()
	defer c.clientMutex.Unlock()
	if c.client == nil {
		// Can only happen if Client is Stopped while trying to connect, so exit silently
		return nil
	}

	token := c.client.Connect()
	if !token.WaitTimeout(timeout) {
		return fmt.Errorf("connect timeout")
	}
	if token.Error() == nil {
		c.subjectInfo.Publish(logPrefix + " Connected")
	}
	return token.Error()
}

// Stop unsubscribes the topics and drops a connection with the NC server
func (c *Client) Stop() error {
	c.clientMutex.Lock()
	defer c.clientMutex.Unlock()
	if c.client == nil {
		return fmt.Errorf("already stopped")
	}

	c.cancelConnecting()
	token := c.client.Unsubscribe(unsubscriptions...)
	if token.WaitTimeout(timeout) && token.Error() != nil {
		c.subjectErr.Publish(fmt.Errorf("%s unsubscribing to topics: %w", logPrefix, token.Error()))
	}
	c.client.Disconnect(0)
	c.client = nil
	c.subjectInfo.Publish(logPrefix + " Client has stopped")

	return nil
}

func (c *Client) onConnect(mqtt.Client) {
	c.clientMutex.Lock()
	defer c.clientMutex.Unlock()
	if c.client == nil {
		return
	}

	token := c.client.SubscribeMultiple(subscriptions, nil)
	if token.WaitTimeout(timeout) && token.Error() != nil {
		c.subjectErr.Publish(
			fmt.Errorf(logPrefix+" subscribing to %v topics: %s", unsubscriptions, token.Error()),
		)
	}
}

func (c *Client) onConnectionLost(_ mqtt.Client, err error) {
	c.subjectErr.Publish(fmt.Errorf("%s connection lost: %s", logPrefix, err.Error()))
	var ctx context.Context
	ctx, c.cancelConnecting = context.WithCancel(context.Background())
	c.connectWithBackoff(ctx, network.ExponentialBackoff)
}

func (c *Client) receiveMessage(_ mqtt.Client, msg mqtt.Message) {
	c.clientMutex.Lock()
	defer c.clientMutex.Unlock()
	if c.client == nil {
		return
	}

	var payload RecPayload
	if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
		c.subjectErr.Publish(fmt.Errorf("%s parsing message payload: %s", logPrefix, err))
		return
	}

	metadata := payload.Message.Data.Metadata
	if opts := c.client.OptionsReader(); metadata.TargetUID != opts.ClientID() {
		c.subjectErr.Publish(
			fmt.Errorf("%s attempted to publish message to incorrect recipient", logPrefix),
		)
		return
	}

	if metadata.Acked {
		c.subjectErr.Publish(
			fmt.Errorf("%s message was already published successfully", logPrefix),
		)
		return
	}

	if err := c.sendDeliveryConfirmation(metadata.MessageID); err != nil {
		c.subjectErr.Publish(
			fmt.Errorf("%s Delivery confirmation: %v", logPrefix, err),
		)
	}

	if err := c.sendAcknowledgement(metadata.MessageID, trackTypeProcessed, ""); err != nil {
		c.subjectErr.Publish(
			fmt.Errorf("%s Acknowledgement: %v", logPrefix, err),
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
