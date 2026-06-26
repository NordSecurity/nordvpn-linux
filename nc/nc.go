/*
Package nc provides a MQTT client to connect to the Notification Centre server.
MQTT could be viewed as an application layer version of TCP since it uses similar
mechanisms to ensure message delivery.
*/
package nc

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net"
	"net/url"
	"slices"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/NordSecurity/nordvpn-linux/network"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	mqttp "github.com/eclipse/paho.mqtt.golang/packets"
)

const (
	typeMeshnet               = "meshnet_network_update"
	typeUserServiceUpdate     = "user_service_update"
	typeDedicatedServerUpdate = "dedicated_server_update"

	topicDelivered    = "delivered"
	topicAcknowledged = "ack"

	trackTypeProcessed = "processed"
)

const (
	timeout = 5 * time.Second
)

var subscriptions = map[string]byte{
	"content":                 byte(1),
	"linux":                   byte(1),
	"meshnet":                 byte(1),
	"user_service_update":     byte(1),
	"dedicated_server_update": byte(1),
}

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
	Revoke() bool
}

type ClientBuilder interface {
	Build(opts *mqtt.ClientOptions) mqtt.Client
}

type MqttClientBuilder struct{}

func (MqttClientBuilder) Build(opts *mqtt.ClientOptions) mqtt.Client {
	return mqtt.NewClient(opts)
}

type CalculateRetryDelayForAttempt func(attempt int) time.Duration

// Client is a client for Notification center
type Client struct {
	clientBuilder ClientBuilder
	// MQTT Docs say that reusing client after doing Disconnect can lead to panics.
	// Since we are doing connect manually with our exponential backoff, we are in risk of those panics.
	// That's why this mutex must be locked every time client is used.
	subjectInfo                   events.Publisher[string]
	subjectErr                    events.Publisher[error]
	subjectPeerUpdate             events.Publisher[[]string]
	subjectServicesUpdate         events.Publisher[any]
	subjectDedicatedServersUpdate events.Publisher[any]
	credsFetcher                  CredentialsGetter
	retryDelayFunc                CalculateRetryDelayForAttempt
	fwmark                        uint32
	resolver                      network.DNSResolver

	startMu          sync.Mutex
	started          bool
	cancelConnecting context.CancelFunc // Used to stop connecting attempts if we are already stopping
	statusChan       <-chan any
}

// NewClient is a constructor for a NC client
func NewClient(
	clientBuilder ClientBuilder,
	subjectInfo events.Publisher[string],
	subjectErr events.Publisher[error],
	subjectPeerUpdate events.Publisher[[]string],
	subjectServicesUpdate events.Publisher[any],
	subjectDedicatedServersUpdate events.Publisher[any],
	credsFetcher CredentialsGetter,
	fwmark uint32,
	resolver network.DNSResolver,
) *Client {
	return &Client{
		clientBuilder:                 clientBuilder,
		subjectInfo:                   subjectInfo,
		subjectErr:                    subjectErr,
		subjectPeerUpdate:             subjectPeerUpdate,
		subjectServicesUpdate:         subjectServicesUpdate,
		subjectDedicatedServersUpdate: subjectDedicatedServersUpdate,
		credsFetcher:                  credsFetcher,
		retryDelayFunc:                network.ExponentialBackoff,
		fwmark:                        fwmark,
		resolver:                      resolver,
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
	ctx context.Context) (*mqtt.ClientOptions, error) {
	opts := mqtt.NewClientOptions()
	opts.SetCleanSession(false)
	opts.SetOrderMatters(false)
	opts.SetAutoReconnect(false) // handled manually with the exponential backoff

	// Parse endpoint URL to extract hostname for DNS resolution and TLS verification.
	// Example: "ssl://mqtt.example.com:8883" -> hostname="mqtt.example.com", port="8883"
	log.NC.Info("try DNS resolution for original endpoint:", credentials.Endpoint)
	u, err := url.Parse(credentials.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("NC original endpoint URL parse error: %w", err)
	}

	hostname := u.Hostname()
	port := u.Port()

	// Resolve hostname to IPs using resolver with fwmark (bypasses killswitch).
	// Add each IP as a separate broker - MQTT library will try them sequentially.
	ips, resolveErr := c.resolver.Resolve(hostname)
	if resolveErr == nil && len(ips) > 0 {
		for _, ip := range ips {
			if ip.Is6() {
				log.NC.Debug("got IPv6 address:", ip, " ignore.")
				continue
			}
			brokerURL := fmt.Sprintf("%s://%s", u.Scheme, ip.String())
			if port != "" {
				brokerURL = fmt.Sprintf("%s:%s", brokerURL, port)
			}
			opts.AddBroker(brokerURL)
		}
	}

	if len(opts.Servers) == 0 {
		if resolveErr != nil {
			log.NC.Error("DNS resolution failed, using original endpoint, err:", resolveErr)
		} else {
			log.NC.Warn("no usable IPv4 addresses resolved, using original endpoint")
		}
		opts.AddBroker(credentials.Endpoint)
	}

	opts.SetUsername(credentials.Username)
	opts.SetPassword(credentials.Password)
	opts.SetClientID(credentials.UserID.String())
	opts.SetConnectTimeout(timeout)

	// Set TLS config with original hostname for certificate verification.
	// Required when connecting to IP address instead of hostname.
	opts.SetTLSConfig(&tls.Config{
		ServerName: hostname,
		MinVersion: tls.VersionTLS12,
	})

	if c.fwmark != 0 {
		// Create dialer with fwmark on TCP socket to bypass killswitch.
		dialer := &net.Dialer{
			Timeout: timeout,
			Control: network.NewFwmarkControlFn(c.fwmark),
		}
		opts.SetDialer(dialer)
	}

	opts.SetDefaultPublishHandler(func(_ mqtt.Client, m mqtt.Message) {
		log.NC.Info("MQTT message received.")
		select {
		case managementChan <- mqttMessage{message: m}:
			return
		case <-ctx.Done():
			log.NC.Info("message received but client was stopped before it could be handled.")
			return
		}
	})

	opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		log.NC.Info("connection lost: ", err)
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

	return opts, nil
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
		log.NC.Info("connection attempt with connected client!")
		return client, connectionState
	}

	if connectionState == needsAuthorization {
		credentials, err := c.credsFetcher.GetCredentialsFromAPI()
		if err != nil {
			logFunc("failed to fetch credentials: ", err.Error())
			return client, needsAuthorization
		}

		// send new creation date to the management loop
		select {
		case managementChan <- credentials.ExpirationDate:
		case <-ctx.Done():
			return client, connectionState
		}

		opts, err := c.createClientOptions(credentials, managementChan, ctx)
		if err != nil {
			logFunc("failed to create client options, err:", err.Error())
			return client, connectionState
		}
		client = c.clientBuilder.Build(opts)
		connectionState = connecting
	}

	if connectionState == connecting {
		token := client.Connect()

		if !token.WaitTimeout(timeout) {
			logFunc("failed to connect: timeout")
			// client is still connecting at this point, so we have to disconnect to not get double connection
			client.Disconnect(0)
			return client, connecting
		}

		if err := token.Error(); err != nil {
			logFunc("failed to connect: ", err.Error())
			if errors.Is(err, mqttp.ErrorRefusedNotAuthorised) {
				logFunc("credentials invalidated, will retry with new creds")
				return client, needsAuthorization
			} else {
				return client, connecting
			}
		}

		c.subjectInfo.Publish("[NC] Connected")
		return client, connectedSuccessfully
	}

	return client, connecting
}

func (c *Client) connectWithBackoff(client mqtt.Client,
	credentialsInvalidated bool,
	managementChan chan<- interface{},
	ctx context.Context) mqtt.Client {
	log.NC.Info("start connection loop")

	connectionState := connecting
	if credentialsInvalidated {
		connectionState = needsAuthorization
	}

	for tries := 0; ; tries++ {
		// we only want to log the errors every on 1st and every 10th try, so that we do not spam the logs
		logFunc := log.NC.Info
		if tries%10 != 0 {
			logFunc = nil
		}
		client, connectionState = c.tryConnect(client, logFunc, connectionState, managementChan, ctx)
		if connectionState == connectedSuccessfully {
			break
		}

		select {
		case <-ctx.Done():
			log.NC.Info("stopping connection loop")
			if client != nil {
				client.Disconnect(0)
			}
			return client
		case <-time.After(c.retryDelayFunc(tries)):
		}
	}

	token := client.SubscribeMultiple(subscriptions, nil)
	if token.WaitTimeout(timeout) && token.Error() != nil {
		c.subjectErr.Publish(
			fmt.Errorf("[NC] subscribing to topics: %s", token.Error()),
		)
	}

	log.NC.Info("Connected")

	return client
}

func (c *Client) connect(client mqtt.Client,
	credentialsInvalidated bool,
	connectionContext context.Context,
	managementChan chan<- interface{},
	connectedChan chan<- mqtt.Client) {
	client = c.connectWithBackoff(client, credentialsInvalidated, managementChan, connectionContext)
	select {
	case connectedChan <- client:
	case <-connectionContext.Done():
	}
}

func publishOnTopic(client mqtt.Client, topic string, payload []byte, ctx context.Context) error {
	token := client.Publish(topic, 1, false, payload)

	select {
	case <-token.Done():
		if err := token.Error(); err != nil {
			return fmt.Errorf("publishing on topic: %w", err)
		}
	case <-ctx.Done():
	}

	return nil
}

func (c *Client) sendDeliveryConfirmation(client mqtt.Client, messageID string, ctx context.Context) error {
	payload, err := json.Marshal(ConfirmationPayload{
		MessageID: messageID,
	})
	if err != nil {
		return fmt.Errorf("marshaling confirmation payload: %w", err)
	}

	err = publishOnTopic(client, topicDelivered, payload, ctx)
	if err != nil {
		return fmt.Errorf("publishing delivery confirmation topic: %w", err)
	}

	return nil
}

func (c *Client) sendAcknowledgement(client mqtt.Client, messageID, trackType, actionSlug string, ctx context.Context) error {
	payload, err := json.Marshal(AcknowledgementPayload{
		MessageID:  messageID,
		TrackType:  trackType,
		ActionSlug: actionSlug,
	})
	if err != nil {
		return fmt.Errorf("marshaling acknowledgement payload: %w", err)
	}

	err = publishOnTopic(client, topicAcknowledged, payload, ctx)
	if err != nil {
		return fmt.Errorf("publishing acknowledgmenet: %w", err)
	}

	return nil
}

func (c *Client) handleMessage(client mqtt.Client, msg mqtt.Message, ctx context.Context) {
	log.NC.Info("handle message")
	var payload RecPayload
	if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
		c.subjectErr.Publish(fmt.Errorf("[NC] parsing message payload: %s", err))
	}

	metadata := payload.Message.Data.Metadata
	if opts := client.OptionsReader(); metadata.TargetUID != opts.ClientID() {
		c.subjectErr.Publish(
			fmt.Errorf("[NC] attempted to publish message to incorrect recipient"),
		)
		return
	}

	if metadata.Acked {
		c.subjectErr.Publish(
			fmt.Errorf("[NC] message was already published successfully"),
		)
		return
	}

	if err := c.sendDeliveryConfirmation(client, metadata.MessageID, ctx); err != nil {
		c.subjectErr.Publish(
			fmt.Errorf("[NC] Delivery confirmation: %v", err),
		)
	}

	if err := c.sendAcknowledgement(client, metadata.MessageID, trackTypeProcessed, "", ctx); err != nil {
		c.subjectErr.Publish(
			fmt.Errorf("[NC] Acknowledgement: %v", err),
		)
	}

	log.NC.Info("received", payload.Message.Data.Event.Type)

	switch payload.Message.Data.Event.Type {
	case typeUserServiceUpdate:
		c.subjectServicesUpdate.Publish(struct{}{})
	case typeMeshnet:
		c.subjectPeerUpdate.Publish(payload.Message.Data.Event.Attributes.AffectedMachines)
	case typeDedicatedServerUpdate:
		c.subjectDedicatedServersUpdate.Publish(struct{}{})
	default:
		log.NC.Warn("received unknown MQTT message type:", payload.Message.Data.Event.Type)
	}
}

// ncClientManagementLoop starts a background goroutine that handles events related to the notification client and
// attempts reconnection in case of disconnection. It returns a status channel that will be closed once the control
// loop stops its operations.
func (c *Client) ncClientManagementLoop(ctx context.Context) (<-chan any, error) {
	managementChan := make(chan interface{})

	log.NC.Info("starting management loop")

	var client mqtt.Client

	connectionContext, cancelConnectionFunc := context.WithCancel(ctx)

	credentialsInvalidated := false
	credentials, err := c.credsFetcher.GetCredentialsFromConfig()
	credsExpirationChan := time.After(time.Until(credentials.ExpirationDate))
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			// Client will be initialized when connecting if credentials are invalidated. We want to do this as a part
			// of connection loop, because we might not have internet connection at this point
			credentialsInvalidated = true
			credsExpirationChan = nil
		} else {
			cancelConnectionFunc()
			return nil, fmt.Errorf("fetching credentials: %w", err)
		}
	}

	statusChan := make(chan any)
	go func() {
		defer func() {
			log.NC.Info("stopping management loop")
			cancelConnectionFunc()
			if client != nil {
				unsubscriptions := slices.Collect(maps.Keys(subscriptions))
				client.Unsubscribe(unsubscriptions...)
				client.Disconnect(0)
				client = nil
			}
			log.NC.Info("stopped management loop")
			close(statusChan)
		}()

		connectedChan := make(chan mqtt.Client)
		opts, err := c.createClientOptions(credentials, managementChan, connectionContext)
		if err != nil {
			log.NC.Error("failed to create client options, err:", err.Error())
			return
		}
		client = c.clientBuilder.Build(opts)
		go c.connect(client, credentialsInvalidated, connectionContext, managementChan, connectedChan)

		log.NC.Info("starting initial connection loop")
	CONNECTION_LOOP:
		for {
			select {
			case <-ctx.Done():
				return
			case client = <-connectedChan:
				break CONNECTION_LOOP
			case event := <-managementChan:
				if newCredentialsExpirationDate, ok := event.(time.Time); ok {
					log.NC.Info("new token expiration time:", newCredentialsExpirationDate)
					credsExpirationChan = time.After(time.Until(newCredentialsExpirationDate))
				}
			case <-credsExpirationChan:
				log.NC.Info("token expired in the initial connection loop")
				credsExpirationChan = nil
				cancelConnectionFunc()
				client.Disconnect(0)
				_, _ = c.credsFetcher.RevokeCredentials(false)
				connectionContext, cancelConnectionFunc = context.WithCancel(ctx)
				go c.connect(client, true, connectionContext, managementChan, connectedChan)
			}
		}
		log.NC.Info("initial connection established")

		for {
			select {
			case <-ctx.Done():
				return
			case client = <-connectedChan:
			case event := <-managementChan:
				switch ev := event.(type) {
				case authLost:
					go c.connect(client, true, connectionContext, managementChan, connectedChan)
				case connectionLost:
					go c.connect(client, false, connectionContext, managementChan, connectedChan)
				case mqttMessage:
					c.handleMessage(client, ev.message, connectionContext)
				case time.Time:
					log.NC.Info("new token expiration time:", ev)
					credsExpirationChan = time.After(time.Until(ev))
				}
			case <-credsExpirationChan:
				log.NC.Info("token expired in the management connection loop")
				credsExpirationChan = nil
				cancelConnectionFunc()
				client.Disconnect(0)
				_, _ = c.credsFetcher.RevokeCredentials(false)
				connectionContext, cancelConnectionFunc = context.WithCancel(ctx)
				go c.connect(client, true, connectionContext, managementChan, connectedChan)
			}
		}
	}()

	return statusChan, nil
}

// Start initiates the connection with the NC server and subscribes to mandatory topics
func (c *Client) Start() error {
	c.startMu.Lock()
	defer c.startMu.Unlock()

	if c.started {
		log.NC.Info("attemtp to start client that was already started")
		return nil
	}

	log.NC.Info("start")

	ctx, cancelFunc := context.WithCancel(context.Background())
	c.cancelConnecting = cancelFunc

	statusChan, err := c.ncClientManagementLoop(ctx)
	if err != nil {
		return fmt.Errorf("starting NC management loop: %w", err)
	}

	c.started = true
	c.statusChan = statusChan

	return nil
}

// Stop unsubscribes the topics and drops a connection with the NC server
func (c *Client) Stop() error {
	c.startMu.Lock()
	defer c.startMu.Unlock()

	if !c.started {
		log.NC.Info("attempt to stop client that was already stopped")
		return nil
	}

	log.NC.Info("stoping NC management loop")
	c.cancelConnecting()
	<-c.statusChan
	c.statusChan = nil
	log.NC.Info("stopped NC management loop")
	c.started = false

	return nil
}

// Revoke revokes the NC communication token
func (c *Client) Revoke() bool {
	c.startMu.Lock()
	defer c.startMu.Unlock()

	if c.started {
		log.NC.Info("attempt to revoke token for running client")
		return false
	}

	ok, err := c.credsFetcher.RevokeCredentials(true)
	if ok {
		log.NC.Info("token revoked successfully")
		return true
	} else {
		log.NC.Info("token not revoked:", err)
		return false
	}
}
