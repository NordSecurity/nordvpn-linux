package state

import (
	"log"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/state/pb"
)

type Server struct {
	statePublisher *StatePublisher
	pb.UnimplementedStateServer
}

func NewServer(statePublisher *StatePublisher) Server {
	return Server{
		statePublisher: statePublisher,
	}
}

// statusStream starts streaming status events received by stateChan to the subscriber. When the stream is stopped(i.e
// when subscribers stops listening), stopChan will be closed.
func statusStream(stateChan <-chan interface{}, stopChan chan<- struct{}, srv pb.State_SubscribeServer) {
	for {
		select {
		case <-srv.Context().Done():
			close(stopChan)
			return
		case ev := <-stateChan:
			switch e := ev.(type) {
			case events.DataConnect:
				state := pb.ConnectionState_CONNECTING
				if e.Type == events.ConnectSuccess {
					state = pb.ConnectionState_CONNECTED
				}

				status := pb.ConnectionStatus{
					State:          state,
					ServerIp:       e.TargetServerIP,
					ServerCountry:  e.TargetServerCountry,
					ServerCity:     e.TargetServerCity,
					ServerName:     e.TargetServerName,
					ServerHostname: e.TargetServerDomain,
					IsMeshPeer:     e.IsMeshnetPeer,
				}
				if err := srv.Send(
					&pb.AppState{State: &pb.AppState_ConnectionStatus{ConnectionStatus: &status}}); err != nil {
					log.Println("vpn enabled failed to send state update: ", err)
				}
			case events.DataDisconnect:
				if err := srv.Send(
					&pb.AppState{State: &pb.AppState_ConnectionStatus{
						ConnectionStatus: &pb.ConnectionStatus{State: pb.ConnectionState_DISCONNECTED}}}); err != nil {
					log.Println("vpn disabled failed to send state update: ", err)
				}
			default:
			}
		}
	}
}

func (s *Server) Subscribe(_ *pb.Empty, srv pb.State_SubscribeServer) error {
	log.Println(internal.InfoPrefix + " Received new subscription request")

	stateChan, stopChan := s.statePublisher.AddSubscriber()
	statusStream(stateChan, stopChan, srv)

	return nil
}
