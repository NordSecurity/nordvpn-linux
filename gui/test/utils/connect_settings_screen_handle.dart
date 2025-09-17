import 'package:nordvpn/widgets/advanced_list_tile.dart';
import 'package:nordvpn/widgets/on_off_switch.dart';

import 'finders.dart';
import 'screen_handle.dart';

final class ConnectionSettingsScreenHandle extends ScreenHandle {
  ConnectionSettingsScreenHandle(super.app);

  bool isAutoConnectSwitchOn() {
    final widget = app.tester.widget<OnOffSwitch>(autoConnectSwitch());
    return widget.value;
  }

  bool isKillSwitchOn() {
    final widget = app.tester.widget<OnOffSwitch>(killSwitchToggle());
    return widget.value;
  }

  Future<void> clickAutoConnectSwitch() async {
    await app.tester.tap(autoConnectSwitch());
    await app.refreshAppState();
  }

  bool isAutoConnectTileEnabled() {
    final widget = app.tester.widget<AdvancedListTile>(autoConnectTile());
    return widget.enabled;
  }
}
