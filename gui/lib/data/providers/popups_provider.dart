import 'package:nordvpn/data/models/popup_metadata.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/internal/popups_metadata.dart';
import 'package:nordvpn/logger.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'popups_provider.g.dart';

@Riverpod(keepAlive: true)
final class Popups extends _$Popups {
  final List<PopupMetadata> _popups = [];

  @override
  PopupMetadata? build() => null;

  void show(int id) {
    if (_shouldIgnore(id)) {
      logger.d("ignoring popup with id: $id");
      return;
    }
    logger.i("showing popup for id: $id");

    final metadata = givePopupMetadata(id);
    _showWithMetadata(metadata);
  }

  bool _shouldIgnore(int code) {
    return code <= 2000 ||
        code ==
            3045 // CodeAllowlistSubnetNoop should be ignored by GUI
            ||
        code == 3047; // CodeAllowlistPortNoop should be ignored by GUI
  }

  void _showWithMetadata(PopupMetadata metadata) {
    if (metadata.id == DaemonStatusCode.success) {
      return;
    }

    if (state == null) {
      state = metadata;
    } else if (state! != metadata && !_popups.contains(metadata)) {
      _popups.add(metadata);
    }
  }

  void pop() {
    state = _popups.firstOrNull;
    if (_popups.isNotEmpty) {
      _popups.removeAt(0);
    }
  }
}
