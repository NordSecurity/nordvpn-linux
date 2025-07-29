package remote

import (
	"log"

	"github.com/NordSecurity/nordvpn-linux/daemon/events"
)

type Analytics interface {
	NotifyDownload(client, featureName string, err error)
	NotifyLocalUse(client, featureName string, err error)
	NotifyJsonParse(client, featureName string, err error)
	NotifyPartialRollout(client, featureName string, frg int)
}

type MooseAnalytics struct {
	de  events.DebuggerEvents
	ctx Event
}

func NewMooseAnalytics(de events.DebuggerEvents, ver string, rg int) *MooseAnalytics {
	ctx := Event{
		UserInfo: UserInfo{
			AppVersion:   ver,
			RolloutGroup: rg,
		},
	}
	return &MooseAnalytics{de: de, ctx: ctx}
}

func (ma *MooseAnalytics) NotifyDownload(client, featureName string, err error) {
	log.Println("~~~NotifyDownload start")
	defer log.Println("~~~NotifyDownload end")

	var evt Event
	if err != nil { //TODO/FIXME: downloadErrorKind
		evt = NewDownloadFailureEvent(ma.ctx.UserInfo, client, featureName, DownloadErrorOther, err.Error())
	} else {
		evt = NewDownloadSuccessEvent(ma.ctx.UserInfo, client, featureName)
	}
	ma.de.DebuggerEvents.Publish(
		*evt.ToDebuggerEvent())
}

func (ma *MooseAnalytics) NotifyLocalUse(client, featureName string, err error) {
	log.Println("~~~NotifyLocalUse start")
	defer log.Println("~~~NotifyLocalUse end")

	ma.de.DebuggerEvents.Publish(
		*NewLocalUseEvent(ma.ctx.UserInfo, client, featureName).ToDebuggerEvent())
}

func (ma *MooseAnalytics) NotifyJsonParse(client, featureName string, err error) {
	log.Println("~~~NotifyJsonParse start")
	defer log.Println("~~~NotifyJsonParse end")

	ma.de.DebuggerEvents.Publish(
		*NewJSONParseEvent(ma.ctx.UserInfo, client, featureName, "", err.Error()).
			ToDebuggerEvent())
}

func (ma *MooseAnalytics) NotifyPartialRollout(client, featureName string, frg int) {
	log.Println("~~~NotifyPartialRollout start")
	defer log.Println("~~~NotifyPartialRollout end")

	ma.de.DebuggerEvents.Publish(
		*NewRolloutEvent(ma.ctx.UserInfo, client, featureName, frg, true).
			ToDebuggerEvent())
}
