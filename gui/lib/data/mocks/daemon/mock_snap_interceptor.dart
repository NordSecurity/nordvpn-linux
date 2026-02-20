import 'dart:convert';

import 'package:grpc/grpc.dart';
import 'package:nordvpn/data/mocks/daemon/mock_error_interceptor.dart';
import 'package:nordvpn/pb/snapconf/snapconf.pb.dart';

const mockedMissingConnections = [
  "network",
  "network-bind",
  "network-control",
  "firewall-control",
];

class MockSnapErrorInterceptor extends MockErrorInterceptor {
  MockSnapErrorInterceptor() : super(error: null);

  void setEnabled(bool enabled) {
    if (enabled) {
      super.setError(_buildSnapError());
    } else {
      super.setError(null);
    }
  }

  static GrpcError _buildSnapError() {
    final detail = ErrMissingConnections(
      missingConnections: mockedMissingConnections,
    );

    // Dart GRPC implementation does not send the details through the wire
    // We have to force it using this hack
    // We encode the error the way it is encoded by the daemon using Status structure
    // And add it to trailers forcing the dart GRPC server to send it through the wire
    final statusBytes = _buildStatusBinary(
      code: 7,
      message: 'Snap permissions required',
      detail: detail,
    );
    final trailers = {'grpc-status-details-bin': base64Encode(statusBytes)};

    final err = GrpcError.custom(
      7,
      'Snap permissions required',
      null,
      null,
      trailers,
    );

    return err;
  }
}

List<int> _buildStatusBinary({
  required int code,
  required String message,
  required ErrMissingConnections detail,
}) {
  final statusBytes = <int>[];

  // field 1: code (varint)
  statusBytes.addAll(_encodeTag(1, 0));
  statusBytes.addAll(_encodeVarint(code));

  // field 2: message (string)
  statusBytes.addAll(_encodeTag(2, 2));
  statusBytes.addAll(_encodeLengthDelimited(utf8.encode(message)));

  // field 3: details (Any)
  final detailBytes = detail.writeToBuffer();

  final anyBytes = _buildAny(
    'type.googleapis.com/snappb.ErrMissingConnections',
    detailBytes,
  );

  statusBytes.addAll(_encodeTag(3, 2));
  statusBytes.addAll(_encodeLengthDelimited(anyBytes));

  return statusBytes;
}

List<int> _buildAny(String typeUrl, List<int> valueBytes) {
  final bytes = <int>[];

  // field 1: type_url (string)
  bytes.addAll(_encodeTag(1, 2));
  bytes.addAll(_encodeLengthDelimited(utf8.encode(typeUrl)));

  // field 2: value (bytes)
  bytes.addAll(_encodeTag(2, 2));
  bytes.addAll(_encodeLengthDelimited(valueBytes));

  return bytes;
}

List<int> _encodeVarint(int value) {
  final bytes = <int>[];

  while (true) {
    if ((value & ~0x7F) == 0) {
      bytes.add(value);
      break;
    } else {
      bytes.add((value & 0x7F) | 0x80);
      value >>= 7;
    }
  }

  return bytes;
}

List<int> _encodeTag(int field, int wireType) {
  return _encodeVarint((field << 3) | wireType);
}

List<int> _encodeLengthDelimited(List<int> data) {
  return [..._encodeVarint(data.length), ...data];
}
