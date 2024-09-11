/*
Package nc provides a MQTT client to connect to the Notification Centre server.
MQTT could be viewed as an application layer version of TCP since it uses similar
mechanisms to ensure message delivery.
*/
package nc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/network"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	mqttp "github.com/eclipse/paho.mqtt.golang/packets"
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
	Start() error
	Stop() error
	Revoke(bool) bool
}

type ClientBuilder interface {
	Build(opts *mqtt.ClientOptions) mqtt.Client
}

type MqttClientBuilder struct{}

func (MqttClientBuilder) Build(opts *mqtt.ClientOptions) mqtt.Client {
	return mqtt.NewClient(opts)
}

// Client is a client for Notification center
type Client struct {
	clientBuilder ClientBuilder
	// MQTT Docs say that reusing client after doing Disconnect can lead to panics.
	// Since we are doing connect manually with our exponential backoff, we are in risk of those panics.
	// That's why this mutex must be locked everytime client is used.
	subjectInfo       events.Publisher[string]
	subjectErr        events.Publisher[error]
	subjectPeerUpdate events.Publisher[[]string]
	credsFetcher      CredentialsGetter

	startMu          sync.Mutex
	started          bool
	cancelConnecting context.CancelFunc // Used to stop connecting attempts if we are already stopping
}

// NewClient is a constructor for a NC client
func NewClient(
	clientBuilder ClientBuilder,
	subjectInfo events.Publisher[string],
	subjectErr events.Publisher[error],
	subjectPeerUpdate events.Publisher[[]string],
	credsFetcher CredentialsGetter,
) *Client {
	return &Client{
		clientBuilder:     clientBuilder,
		subjectInfo:       subjectInfo,
		subjectErr:        subjectErr,
		subjectPeerUpdate: subjectPeerUpdate,
		credsFetcher:      credsFetcher,
	}
}

type authLost struct{}

type connectionLost struct{}

type mqttMessage struct {
	message mqtt.Message
}

func (c *Client) createClientOptions(
	credentials config.NCData,
	managementChan chan<- interface{},
	ctx context.Context) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.SetCleanSession(false)
	opts.SetOrderMatters(false)
	opts.SetAutoReconnect(false) // handled manually with the exponential backoff
	opts.AddBroker(credentials.Endpoint)
	opts.SetUsername(credentials.Username)
	opts.SetPassword(credentials.Password)
	opts.SetClientID(credentials.UserID.String())
	opts.SetConnectTimeout(timeout)

	opts.SetDefaultPublishHandler(func(_ mqtt.Client, m mqtt.Message) {
		log.Println(logPrefix, "MQTT message received.")
		select {
		case managementChan <- mqttMessage{message: m}:
			return
		case <-ctx.Done():
			log.Println(logPrefix, "message received but client was stopped before it could be handled.")
			return
		}
	})

	opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		log.Println(logPrefix, "connection lost: ", err)
		var message interface{}
		if errors.Is(err, mqttp.ErrorRefusedNotAuthorised) {
			message = authLost{}
		} else {
			message = connectionLost{}
		}
		select {
		case managementChan <- message:
			return
		case <-ctx.Done():
			return
		}
	})

	return opts
}

type connectionState int

// connection states
const (
	needsAuthorization connectionState = iota
	connecting
	connectedSuccessfully
)

// tryConnect performs connection actions appropriate to provided connection state and returns new client and state
// after those actions were performed. The desired state is always connectedSuccessfully, so providing this state will
// result in a noop.
func (c *Client) tryConnect(
	client mqtt.Client,
	logFunc func(v ...any),
	connectionState connectionState,
	managementChan chan<- interface{},
	ctx context.Context) (mqtt.Client, connectionState) {
	if logFunc == nil {
		logFunc = func(args ...any) {}
	}

	if connectionState == connectedSuccessfully {
		// this error is unusual(this function should never be called with a 'connected' sate), so it's better to risk
		// spaming the logs so that we do not miss it.
		log.Println(logPrefix, "connection attempt with connected client!")
		return client, connectionState
	}

	if connectionState == needsAuthorization {
		credentials, err := c.credsFetcher.GetCredentialsFromAPI()
		if err != nil {
			logFunc(logPrefix, "failed to fetch credentials: ", err.Error())
			return client, needsAuthorization
		}

		opts := c.createClientOptions(credentials, managementChan, ctx)
		client = c.clientBuilder.Build(opts)
		connectionState = connecting
	}

	if connectionState == connecting {
		token := client.Connect()

		if !token.WaitTimeout(timeout) {
			logFunc(logPrefix, "failed to connect: timeout")
			// client is still connecting at this point, so we have to disconnect to not get double connection
			client.Disconnect(0)
			return client, connecting
		}

		if err := token.Error(); err != nil {
			logFunc(logPrefix, "failed to connect: ", err.Error())
			if errors.Is(err, mqttp.ErrorRefusedNotAuthorised) {
				logFunc(logPrefix, "credentials invalidated, will retry with new creds")
				return client, needsAuthorization
			} else {
				return client, connecting
			}
		}

		c.subjectInfo.Publish(logPrefix + " Connected")
		return client, connectedSuccessfully
	}

	return client, connecting
}

func (c *Client) connectWithBackoff(client mqtt.Client,
	credentialsInvalidated bool,
	managementChan chan<- interface{},
	ctx context.Context) mqtt.Client {
	log.Println(logPrefix, "start connection loop")

	connectionState := connecting
	if credentialsInvalidated {
		connectionState = needsAuthorization
	}

	for tries := 0; ; tries++ {
		// we only want to log the errors every on 1st and every 10th try, so that we do not spam the logs
		logFunc := log.Println
		if tries%10 != 0 {
			logFunc = nil
		}
		client, connectionState = c.tryConnect(client, logFunc, connectionState, managementChan, ctx)
		if connectionState == connectedSuccessfully {
			break
		}

		select {
		case <-ctx.Done():
			log.Println(logPrefix, "stopping connection loop")
			if client != nil {
				client.Disconnect(0)
			}
			return client
		case <-time.After(network.ExponentialBackoff(tries)):
		}
	}

	token := client.SubscribeMultiple(subscriptions, nil)
	if token.WaitTimeout(timeout) && token.Error() != nil {
		c.subjectErr.Publish(
			fmt.Errorf(logPrefix+" subscribing to %v topics: %s", unsubscriptions, token.Error()),
		)
	}

	log.Println(logPrefix, "Connected")

	return client
}

func (c *Client) sendDeliveryConfirmation(client mqtt.Client, messageID string) error {
	payload, err := json.Marshal(ConfirmationPayload{
		MessageID: messageID,
	})
	if err != nil {
		return fmt.Errorf("marshaling confirmation payload: %w", err)
	}
	if token := client.Publish(topicDelivered, 1, false, payload); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (c *Client) sendAcknowledgement(client mqtt.Client, messageID, trackType, actionSlug string) error {
	payload, err := json.Marshal(AcknowledgementPayload{
		MessageID:  messageID,
		TrackType:  trackType,
		ActionSlug: actionSlug,
	})
	if err != nil {
		return fmt.Errorf("marshaling acknowledgement payload: %w", err)
	}
	if token := client.Publish(topicAcknowledged, 1, false, payload); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (c *Client) handleMessage(client mqtt.Client, msg mqtt.Message) {
	log.Println(logPrefix, "handle message")
	var payload RecPayload
	if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
		c.subjectErr.Publish(fmt.Errorf("%s parsing message payload: %s", logPrefix, err))
	}

	metadata := payload.Message.Data.Metadata
	if opts := client.OptionsReader(); metadata.TargetUID != opts.ClientID() {
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

	if err := c.sendDeliveryConfirmation(client, metadata.MessageID); err != nil {
		c.subjectErr.Publish(
			fmt.Errorf("%s Delivery confirmation: %v", logPrefix, err),
		)
	}

	if err := c.sendAcknowledgement(client, metadata.MessageID, trackTypeProcessed, ""); err != nil {
		c.subjectErr.Publish(
			fmt.Errorf("%s Acknowledgement: %v", logPrefix, err),
		)
	}

	if payload.Message.Data.Event.Type == typeMeshnet {
		c.subjectPeerUpdate.Publish(payload.Message.Data.Event.Attributes.AffectedMachines)
	}
}

// ncClientManagementLoop starts a background goroutine that handles events related to the notification client and
// attempts reconnection in case of disconnection.
func (c *Client) ncClientManagementLoop(ctx context.Context) error {
	managementChan := make(chan interface{})

	log.Println(logPrefix, "starting management loop")

	var client mqtt.Client

	credentialsInvalidated := false
	credentials, err := c.credsFetcher.GetCredentialsFromConfig()
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			// Client will be initialized when connecting if credentials are invalidated. We want to do this as a part
			// of connection loop, because we might not have internet connection at this point
			credentialsInvalidated = true
		} else if err != nil {
			return fmt.Errorf("fetching credentials: %w", err)
		}
	} else {
		opts := c.createClientOptions(credentials, managementChan, ctx)
		client = c.clientBuilder.Build(opts)
	}

	go func() {
		client = c.connectWithBackoff(client, credentialsInvalidated, managementChan, ctx)
		for {
			select {
			case <-ctx.Done():
				log.Println(logPrefix, "stopping management loop")
				if client != nil {
					client.Unsubscribe(unsubscriptions...)
					client.Disconnect(0)
					client = nil
				}
				log.Println(logPrefix, "stopped management loop")
				return
			case event := <-managementChan:
				switch ev := event.(type) {
				case authLost:
					client = c.connectWithBackoff(client, true, managementChan, ctx)
				case connectionLost:
					client = c.connectWithBackoff(client, false, managementChan, ctx)
				case mqttMessage:
					c.handleMessage(client, ev.message)
				}
			}
		}
	}()

	return nil
}

// Start initiates the connection with the NC server and subscribes to mandatory topics
func (c *Client) Start() error {
	c.startMu.Lock()
	defer c.startMu.Unlock()

	if c.started {
		log.Println(logPrefix, "attemtp to start client that was already started")
		return nil
	}

	log.Println(logPrefix, "start")

	ctx, cancelFunc := context.WithCancel(context.Background())
	c.cancelConnecting = cancelFunc

	if err := c.ncClientManagementLoop(ctx); err != nil {
		return fmt.Errorf("starting NC management loop: %w", err)
	}

	c.started = true

	return nil
}

// Stop unsubscribes the topics and drops a connection with the NC server
func (c *Client) Stop() error {
	c.startMu.Lock()
	defer c.startMu.Unlock()

	if !c.started {
		log.Println(logPrefix, "attempt to stop client that was already stopped")
		return nil
	}

	log.Println(logPrefix, "stop")
	c.cancelConnecting()
	c.started = false

	return nil
}

// Revoke revokes the NC communication token
func (c *Client) Revoke(purgeSession bool) bool {
	c.startMu.Lock()
	defer c.startMu.Unlock()

	if c.started {
		log.Println(logPrefix, "attempt to revoke token for running client")
		return false
	}

	log.Println(logPrefix, "revoking token, purgeSession:", purgeSession)
	ok, err := c.credsFetcher.RevokeCredentials(purgeSession)
	if ok {
		log.Println(logPrefix, "token revoked successfully")
		return true
	} else {
		log.Println(logPrefix, "token not revoked:", err)
		return false
	}
}
