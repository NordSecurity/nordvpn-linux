package tray

import (
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"

	"github.com/NordSecurity/nordvpn-linux/log"
)

const (
	sniPath      = dbus.ObjectPath("/StatusNotifierItem")
	sniIface     = "org.kde.StatusNotifierItem"
	watcherDest  = "org.kde.StatusNotifierWatcher"
	watcherPath  = dbus.ObjectPath("/StatusNotifierWatcher")
	watcherIface = "org.kde.StatusNotifierWatcher"

	sniCategory = "ApplicationStatus"
	sniID       = "nordvpn"
	sniTitle    = "NordVPN"
	sniStatus   = "Active"

	dbusListNames    = "org.freedesktop.DBus.ListNames"
	dbusIntrospect   = "org.freedesktop.DBus.Introspectable"
	dbusRegisterItem = watcherIface + ".RegisterStatusNotifierItem"

	sniServiceNameFmt = "org.kde.StatusNotifierItem-%d-1"
)

type sniPixmap struct {
	W, H int32
	D    []byte
}
type sniTooltip struct {
	S    string
	P    []sniPixmap
	T, E string
}

var (
	ActivateCh = make(chan struct{}, 1)
	sniDone    = make(chan struct{})
)

func sniStop() { close(sniDone) }

func sniIsAvailable() bool {
	conn, err := dbus.SessionBus()
	if err != nil {
		return false
	}
	var names []string
	if err := conn.BusObject().Call(dbusListNames, 0).Store(&names); err != nil {
		return false
	}
	return slices.Contains(names, watcherDest)
}

func sniRun(baseIcon, connectedIcon string) {
	conn, serviceName, err := sniConnect()
	if err != nil {
		log.Errorf("%s connect failed: %v", logTag, err)
		return
	}
	defer conn.Close() // nolint:errcheck

	props, err := sniExport(conn, baseIcon)
	if err != nil {
		log.Errorf("%s export failed: %v", logTag, err)
		return
	}

	sniRegister(conn, serviceName)
	go sniStayRegistered(conn, serviceName)
	go sniWatchState(conn, props, baseIcon, connectedIcon)
	<-sniDone
	log.Infof("%s stopped", logTag)
}

func sniConnect() (*dbus.Conn, string, error) {
	conn, err := dbus.SessionBusPrivate()
	if err != nil {
		return nil, "", err
	}
	if err := conn.Auth(nil); err != nil {
		_ = conn.Close()
		return nil, "", err
	}
	if err := conn.Hello(); err != nil {
		_ = conn.Close()
		return nil, "", err
	}
	name := fmt.Sprintf(sniServiceNameFmt, os.Getpid())
	reply, err := conn.RequestName(name, dbus.NameFlagDoNotQueue)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		_ = conn.Close()
		return nil, "", fmt.Errorf("claim %s: %w", name, err)
	}
	return conn, name, nil
}

func sniExport(conn *dbus.Conn, iconName string) (*prop.Properties, error) {
	propMap := prop.Map{
		sniIface: {
			"Category":   {Value: sniCategory, Emit: prop.EmitConst},
			"Id":         {Value: sniID, Emit: prop.EmitConst},
			"Title":      {Value: sniTitle, Emit: prop.EmitConst},
			"Status":     {Value: sniStatus, Emit: prop.EmitFalse},
			"IconName":   {Value: iconName, Emit: prop.EmitFalse},
			"IconPixmap": {Value: []sniPixmap{}, Emit: prop.EmitFalse},
			"ItemIsMenu": {Value: false, Emit: prop.EmitConst},
			"Menu":       {Value: dbus.ObjectPath("/"), Emit: prop.EmitConst},
			"ToolTip":    {Value: sniTooltip{T: sniTitle}, Emit: prop.EmitFalse},
		},
	}
	props, err := prop.Export(conn, sniPath, propMap)
	if err != nil {
		return nil, fmt.Errorf("props: %w", err)
	}

	methods := map[string]any{
		"Activate": func(x, y int32) *dbus.Error { return sniActivate() },
	}
	if err := conn.ExportMethodTable(methods, sniPath, sniIface); err != nil {
		return nil, fmt.Errorf("methods: %w", err)
	}

	node := &introspect.Node{
		Name: string(sniPath),
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			sniInterfaceSpec,
		},
	}
	if err := conn.Export(introspect.NewIntrospectable(node), sniPath, dbusIntrospect); err != nil {
		return nil, fmt.Errorf("introspect: %w", err)
	}

	return props, nil
}

func sniActivate() *dbus.Error {
	log.Infof("%s activated", logTag)
	select {
	case ActivateCh <- struct{}{}:
	default:
	}
	return nil
}

func sniRegister(conn *dbus.Conn, serviceName string) {
	obj := conn.Object(watcherDest, watcherPath)
	if err := obj.Call(dbusRegisterItem, 0, serviceName).Err; err != nil {
		log.Errorf("%s register failed: %v", logTag, err)
	} else {
		log.Infof("%s registered as %s", logTag, serviceName)
	}
}

func sniStayRegistered(conn *dbus.Conn, serviceName string) {
	if err := conn.AddMatchSignal(
		dbus.WithMatchSender("org.freedesktop.DBus"),
		dbus.WithMatchMember("NameOwnerChanged"),
		dbus.WithMatchArg(0, watcherDest),
	); err != nil {
		log.Errorf("%s AddMatchSignal failed: %v", logTag, err)
		return
	}

	ch := make(chan *dbus.Signal, 4)
	conn.Signal(ch)
	defer conn.RemoveSignal(ch)

	for {
		select {
		case sig := <-ch:
			if sig == nil || len(sig.Body) < 3 {
				continue
			}
			if newOwner, _ := sig.Body[2].(string); newOwner != "" {
				time.Sleep(200 * time.Millisecond)
				sniRegister(conn, serviceName)
			}
		case <-sniDone:
			return
		}
	}
}

var sniInterfaceSpec = introspect.Interface{
	Name: sniIface,
	Methods: []introspect.Method{
		{Name: "Activate", Args: sniXYArgs},
	},
	Signals: []introspect.Signal{
		{Name: "NewIcon"},
		{Name: "NewStatus", Args: []introspect.Arg{{Name: "status", Type: "s"}}},
	},
}

var sniXYArgs = []introspect.Arg{
	{Name: "x", Type: "i", Direction: "in"},
	{Name: "y", Type: "i", Direction: "in"},
}
