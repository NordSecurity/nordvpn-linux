import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/models/server_info.dart';

void main() {
  test('Server information is parsed correctly', () async {
    final server = ServerInfo(
      id: 9999,
      hostname: "de1234.nordvpn.com",
      isVirtual: true,
    );
    expect(server.serverNumber, "1234");
    expect(server.serverName(), "de1234");
  });
}
