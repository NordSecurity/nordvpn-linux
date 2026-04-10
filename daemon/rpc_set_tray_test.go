package daemon

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/peer"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	testnorduser "github.com/NordSecurity/nordvpn-linux/test/mock/norduser/service"
)

const trayTestUID uint32 = 1000

func peerCtx(uid uint32) context.Context {
	return peer.NewContext(
		context.Background(),
		&peer.Peer{AuthInfo: internal.UcredAuth{Uid: uid}},
	)
}

func TestSetTray_NoPeerContext(t *testing.T) {
	category.Set(t, category.Unit)

	cm := mock.NewMockConfigManager()
	norduserMock := testnorduser.NewMockNorduserCombinedService()
	rpc := &RPC{cm: cm, norduser: &norduserMock}

	resp, err := rpc.SetTray(context.Background(), &pb.SetTrayRequest{Tray: false})

	assert.NoError(t, err)
	assert.Equal(t, internal.CodeInternalError, resp.Type)
	assert.Equal(t, 0, cm.SaveCallCount)
}

func TestSetTray_NoUcredAuth(t *testing.T) {
	category.Set(t, category.Unit)

	cm := mock.NewMockConfigManager()
	norduserMock := testnorduser.NewMockNorduserCombinedService()
	rpc := &RPC{cm: cm, norduser: &norduserMock}

	ctx := peer.NewContext(context.Background(), &peer.Peer{}) // no AuthInfo
	resp, err := rpc.SetTray(ctx, &pb.SetTrayRequest{Tray: false})

	assert.NoError(t, err)
	assert.Equal(t, internal.CodeInternalError, resp.Type)
	assert.Equal(t, 0, cm.SaveCallCount)
}

func TestSetTray_Disable(t *testing.T) {
	category.Set(t, category.Unit)

	cm := mock.NewMockConfigManager()
	norduserMock := testnorduser.NewMockNorduserCombinedService()
	rpc := &RPC{cm: cm, norduser: &norduserMock}

	resp, err := rpc.SetTray(peerCtx(trayTestUID), &pb.SetTrayRequest{Tray: false})

	assert.NoError(t, err)
	assert.Equal(t, internal.CodeSuccess, resp.Type)
	assert.True(t, cm.Cfg.UsersData.TrayOff[int64(trayTestUID)], "tray should be recorded as off")
	assert.Equal(t, []uint32{trayTestUID}, norduserMock.ActionToUIDs[testnorduser.Restart])
}

func TestSetTray_Enable(t *testing.T) {
	category.Set(t, category.Unit)

	cm := mock.NewMockConfigManager()
	cm.Cfg.UsersData.TrayOff[int64(trayTestUID)] = true // start with tray off
	norduserMock := testnorduser.NewMockNorduserCombinedService()
	rpc := &RPC{cm: cm, norduser: &norduserMock}

	resp, err := rpc.SetTray(peerCtx(trayTestUID), &pb.SetTrayRequest{Tray: true})

	assert.NoError(t, err)
	assert.Equal(t, internal.CodeSuccess, resp.Type)
	assert.False(t, cm.Cfg.UsersData.TrayOff[int64(trayTestUID)], "tray should be removed from TrayOff")
	assert.Equal(t, []uint32{trayTestUID}, norduserMock.ActionToUIDs[testnorduser.Restart])
}

func TestSetTray_AlreadyEnabled_NothingToDo(t *testing.T) {
	category.Set(t, category.Unit)

	cm := mock.NewMockConfigManager()
	norduserMock := testnorduser.NewMockNorduserCombinedService()
	rpc := &RPC{cm: cm, norduser: &norduserMock}

	resp, err := rpc.SetTray(peerCtx(trayTestUID), &pb.SetTrayRequest{Tray: true})

	assert.NoError(t, err)
	assert.Equal(t, internal.CodeNothingToDo, resp.Type)
	assert.Equal(t, 0, cm.SaveCallCount)
	assert.Empty(t, norduserMock.ActionToUIDs[testnorduser.Restart])
}

func TestSetTray_AlreadyDisabled_NothingToDo(t *testing.T) {
	category.Set(t, category.Unit)

	cm := mock.NewMockConfigManager()
	cm.Cfg.UsersData.TrayOff[int64(trayTestUID)] = true // tray already off
	norduserMock := testnorduser.NewMockNorduserCombinedService()
	rpc := &RPC{cm: cm, norduser: &norduserMock}

	resp, err := rpc.SetTray(peerCtx(trayTestUID), &pb.SetTrayRequest{Tray: false})

	assert.NoError(t, err)
	assert.Equal(t, internal.CodeNothingToDo, resp.Type)
	assert.Equal(t, 0, cm.SaveCallCount)
	assert.Empty(t, norduserMock.ActionToUIDs[testnorduser.Restart])
}

func TestSetTray_ConfigSaveError(t *testing.T) {
	category.Set(t, category.Unit)

	cm := mock.NewMockConfigManager()
	cm.SaveErr = errors.New("disk full")
	norduserMock := testnorduser.NewMockNorduserCombinedService()
	rpc := &RPC{cm: cm, norduser: &norduserMock}

	resp, err := rpc.SetTray(peerCtx(trayTestUID), &pb.SetTrayRequest{Tray: false})

	assert.NoError(t, err)
	assert.Equal(t, internal.CodeConfigError, resp.Type)
	assert.Empty(t, norduserMock.ActionToUIDs[testnorduser.Restart])
}

func TestSetTray_UidIsolation(t *testing.T) {
	category.Set(t, category.Unit)

	const uidA uint32 = 1001
	const uidB uint32 = 1002

	cm := mock.NewMockConfigManager()
	norduserMock := testnorduser.NewMockNorduserCombinedService()
	rpc := &RPC{cm: cm, norduser: &norduserMock}

	resp, err := rpc.SetTray(peerCtx(uidA), &pb.SetTrayRequest{Tray: false})

	assert.NoError(t, err)
	assert.Equal(t, internal.CodeSuccess, resp.Type)
	assert.True(t, cm.Cfg.UsersData.TrayOff[int64(uidA)], "uidA tray should be off")
	assert.False(t, cm.Cfg.UsersData.TrayOff[int64(uidB)], "uidB tray should be unaffected")
	assert.Equal(t, []uint32{uidA}, norduserMock.ActionToUIDs[testnorduser.Restart], "only uidA should be restarted")
}
