import 'package:nordvpn/data/repository/vpn_repository.dart';
import 'package:nordvpn/pb/daemon/servers.pb.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'recommended_server_provider.g.dart';

@Riverpod(keepAlive: true)
class RecommendedServer extends _$RecommendedServer {
  @override
  FutureOr<RecommendedServerLocation> build() async {
    return await ref
        .read(vpnRepositoryProvider)
        .fetchRecommendedServerLocation();
  }
}
