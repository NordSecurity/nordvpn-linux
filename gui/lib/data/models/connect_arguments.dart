import 'package:nordvpn/data/models/city.dart';
import 'package:nordvpn/data/models/country.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/pb/daemon/connect.pb.dart';

// Used to store the arguments used at connect and autoconnect.
final class ConnectArguments {
  final Country? country;
  final City? city;
  final ServerType? _specialtyGroup;
  final ServerInfo? server;

  ConnectArguments({
    this.country,
    this.city,
    ServerType? specialtyGroup,
    this.server,
  }) : _specialtyGroup = specialtyGroup;

  ServerType? get specialtyGroup {
    if (_specialtyGroup == null) {
      return null;
    } else if (_specialtyGroup == ServerType.standardVpn) {
      // for standard VPN servers return null because it is not specialty group
      return null;
    }
    return _specialtyGroup;
  }

  ConnectArguments copyWith({
    Country? country,
    City? city,
    ServerType? specialtyGroup,
    ServerInfo? server,
  }) {
    return ConnectArguments(
      country: country ?? this.country,
      city: city ?? this.city,
      specialtyGroup: specialtyGroup ?? this.specialtyGroup,
      server: server ?? this.server,
    );
  }

  // Convert to the class needed to communicate with the daemon
  ConnectRequest toConnectRequest() {
    final connectRequest = ConnectRequest();
    if (server != null) {
      // if server is specified pass the server name
      connectRequest.serverTag = server!.serverName();
    } else {
      if (country != null) {
        connectRequest.serverTag = country!.code.toLowerCase();
      }

      if (city != null) {
        connectRequest.serverTag += " ${city!.sanitizedName}";
      }

      if ((specialtyGroup != null) &&
          (specialtyGroup != ServerType.standardVpn)) {
        connectRequest.serverGroup = specialtyGroup!.backendName!;
      }
    }

    return connectRequest;
  }

  @override
  String toString() {
    return "ConnectArguments(countryCode: $country, city: $city, specialty: $specialtyGroup, server: $server)";
  }
}
