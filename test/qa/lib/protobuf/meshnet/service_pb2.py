# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: service.proto
# Protobuf Python Version: 5.28.1
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import runtime_version as _runtime_version
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
_runtime_version.ValidateProtobufRuntimeVersion(
    _runtime_version.Domain.PUBLIC,
    5,
    28,
    1,
    '',
    'service.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


import empty_pb2 as empty__pb2
import fsnotify_pb2 as fsnotify__pb2
import invite_pb2 as invite__pb2
import peer_pb2 as peer__pb2
import service_response_pb2 as service__response__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\rservice.proto\x12\x06meshpb\x1a\x0b\x65mpty.proto\x1a\x0e\x66snotify.proto\x1a\x0cinvite.proto\x1a\npeer.proto\x1a\x16service_response.proto2\xac\x0f\n\x07Meshnet\x12\x37\n\rEnableMeshnet\x12\r.meshpb.Empty\x1a\x17.meshpb.MeshnetResponse\x12\x35\n\tIsEnabled\x12\r.meshpb.Empty\x1a\x19.meshpb.IsEnabledResponse\x12\x38\n\x0e\x44isableMeshnet\x12\r.meshpb.Empty\x1a\x17.meshpb.MeshnetResponse\x12\x38\n\x0eRefreshMeshnet\x12\r.meshpb.Empty\x1a\x17.meshpb.MeshnetResponse\x12\x37\n\nGetInvites\x12\r.meshpb.Empty\x1a\x1a.meshpb.GetInvitesResponse\x12\x37\n\x06Invite\x12\x15.meshpb.InviteRequest\x1a\x16.meshpb.InviteResponse\x12J\n\x0cRevokeInvite\x12\x19.meshpb.DenyInviteRequest\x1a\x1f.meshpb.RespondToInviteResponse\x12\x46\n\x0c\x41\x63\x63\x65ptInvite\x12\x15.meshpb.InviteRequest\x1a\x1f.meshpb.RespondToInviteResponse\x12H\n\nDenyInvite\x12\x19.meshpb.DenyInviteRequest\x1a\x1f.meshpb.RespondToInviteResponse\x12\x33\n\x08GetPeers\x12\r.meshpb.Empty\x1a\x18.meshpb.GetPeersResponse\x12\x43\n\nRemovePeer\x12\x19.meshpb.UpdatePeerRequest\x1a\x1a.meshpb.RemovePeerResponse\x12W\n\x12\x43hangePeerNickname\x12!.meshpb.ChangePeerNicknameRequest\x1a\x1e.meshpb.ChangeNicknameResponse\x12]\n\x15\x43hangeMachineNickname\x12$.meshpb.ChangeMachineNicknameRequest\x1a\x1e.meshpb.ChangeNicknameResponse\x12G\n\x0c\x41llowRouting\x12\x19.meshpb.UpdatePeerRequest\x1a\x1c.meshpb.AllowRoutingResponse\x12\x45\n\x0b\x44\x65nyRouting\x12\x19.meshpb.UpdatePeerRequest\x1a\x1b.meshpb.DenyRoutingResponse\x12I\n\rAllowIncoming\x12\x19.meshpb.UpdatePeerRequest\x1a\x1d.meshpb.AllowIncomingResponse\x12G\n\x0c\x44\x65nyIncoming\x12\x19.meshpb.UpdatePeerRequest\x1a\x1c.meshpb.DenyIncomingResponse\x12Q\n\x11\x41llowLocalNetwork\x12\x19.meshpb.UpdatePeerRequest\x1a!.meshpb.AllowLocalNetworkResponse\x12O\n\x10\x44\x65nyLocalNetwork\x12\x19.meshpb.UpdatePeerRequest\x1a .meshpb.DenyLocalNetworkResponse\x12K\n\x0e\x41llowFileshare\x12\x19.meshpb.UpdatePeerRequest\x1a\x1e.meshpb.AllowFileshareResponse\x12I\n\rDenyFileshare\x12\x19.meshpb.UpdatePeerRequest\x1a\x1d.meshpb.DenyFileshareResponse\x12_\n\x18\x45nableAutomaticFileshare\x12\x19.meshpb.UpdatePeerRequest\x1a(.meshpb.EnableAutomaticFileshareResponse\x12\x61\n\x19\x44isableAutomaticFileshare\x12\x19.meshpb.UpdatePeerRequest\x1a).meshpb.DisableAutomaticFileshareResponse\x12=\n\x07\x43onnect\x12\x19.meshpb.UpdatePeerRequest\x1a\x17.meshpb.ConnectResponse\x12\x43\n\rConnectCancel\x12\x19.meshpb.UpdatePeerRequest\x1a\x17.meshpb.ConnectResponse\x12W\n\x11NotifyNewTransfer\x12\x1f.meshpb.NewTransferNotification\x1a!.meshpb.NotifyNewTransferResponse\x12:\n\rGetPrivateKey\x12\r.meshpb.Empty\x1a\x1a.meshpb.PrivateKeyResponseB2Z0github.com/NordSecurity/nordvpn-linux/meshnet/pbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'service_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'Z0github.com/NordSecurity/nordvpn-linux/meshnet/pb'
  _globals['_MESHNET']._serialized_start=105
  _globals['_MESHNET']._serialized_end=2069
# @@protoc_insertion_point(module_scope)
