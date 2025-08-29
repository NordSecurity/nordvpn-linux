import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/repository/app_state_repository.dart';
import 'package:nordvpn/data/repository/vpn_repository.dart';
import 'package:nordvpn/data/repository/vpn_settings_repository.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/pb/daemon/servers.pb.dart';
import 'package:nordvpn/pb/daemon/settings.pb.dart';
import 'package:nordvpn/pb/daemon/state.pb.dart';
import 'package:nordvpn/pb/daemon/status.pb.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'app_state_provider.g.dart';

// Observers for Account changes need to implement this
abstract class AccountObserver {
  void onAccountChanged(LoginEventType type);
  void onAccountModified(AccountModification accountModification);
}

// Observers for VPN status changes need to implement this
abstract class VpnStatusObserver {
  void onVpnStatusChanged(StatusResponse status);
}

// Observer of the application settings must implement this
abstract class VpnSettingsObserver {
  void onSettingsChanged(ApplicationSettings settings);
}

abstract class ServersListObserver {
  void onServersListChanged(ServersResponse servers);
}

// This will observe the daemon changes and notify the observers
// At the moment this will pull for status, in future the daemon will notify it instead
class AppStateChange {
  final AppStateRepository _appStateRepository;
  final VpnRepository _vpnRepository;
  ApplicationSettings? _appSettings;
  StreamSubscription<AppState>? _stateSubscription;

  final Set<AccountObserver> _accountObservers = {};
  final Set<VpnStatusObserver> _vpnObservers = {};
  final Set<VpnSettingsObserver> _settingsObservers = {};
  final Set<ServersListObserver> _serversListObservers = {};

  AppStateChange(
    AppStateRepository repository,
    VpnRepository vpnRepository,
    VpnSettingsRepository settingsRepository,
  ) : _appStateRepository = repository,
      _vpnRepository = vpnRepository {
    // It initializes the `_appSettings` so that when we are getting the event
    // with changed settings, we can compare it to the initial ones set here and
    // figure out what changed.
    // `_appSettings` will be overwritten anyway after we got updates
    // from daemon, so this is useful only for the first time.
    _maybeInitSettings(settingsRepository);

    _startEventsListener();
  }

  void _maybeInitSettings(VpnSettingsRepository settingsRepository) {
    settingsRepository
        .fetchSettings()
        .then((settings) => _appSettings = settings)
        .ignore();
  }

  void dispose() => _stateSubscription?.cancel();

  void addAccountObserver(AccountObserver observer) {
    _accountObservers.add(observer);
  }

  void removeAccountObserver(AccountObserver observer) {
    _accountObservers.remove(observer);
  }

  void addVpnStatusObserver(VpnStatusObserver observer) {
    _vpnObservers.add(observer);
  }

  void removeVpnStatusObserver(VpnStatusObserver observer) {
    _vpnObservers.remove(observer);
  }

  void addSettingsObserver(VpnSettingsObserver observer) {
    _settingsObservers.add(observer);
  }

  void removeSettingsObserver(VpnSettingsObserver observer) {
    _settingsObservers.remove(observer);
  }

  void addServersListObserver(ServersListObserver observer) {
    _serversListObservers.add(observer);
  }

  void removeServersListObserver(ServersListObserver observer) {
    _serversListObservers.remove(observer);
  }

  void _startEventsListener() {
    _stateSubscription?.cancel();
    _stateSubscription = _appStateRepository.stream.listen(
      (value) {
        logger.d("app state received $value");
        if (value.hasSettingsChange()) {
          _notifySettingsChanged(value.settingsChange);
        } else if (value.hasConnectionStatus()) {
          _notifyVpnStatusChanged(value.connectionStatus);
        } else if (value.hasLoginEvent()) {
          _notifyAccountChanged(value.loginEvent);
        } else if (value.hasAccountModification()) {
          _notifyAccountModified(value.accountModification);
        } else if (value.hasUpdateEvent()) {
          switch (value.updateEvent) {
            case UpdateEvent.SERVERS_LIST_UPDATE:
              _notifyServersListChanged();
              break;
          }
        }
      },
      onError: (error) {
        logger.e("app state listener ended error: $error");
      },
      onDone: () {
        logger.f("app state listener closed");
        // TODO: check if there is a safer way instead of recursive call
        Future.delayed(Duration(seconds: 3), () {
          _startEventsListener();
        });
      },
    );
  }

  void _notifyAccountChanged(LoginEvent event) {
    if (_accountObservers.isEmpty) {
      return;
    }
    for (final observer in _accountObservers) {
      observer.onAccountChanged(event.type);
    }
  }

  void _notifyAccountModified(AccountModification modification) {
    if (_accountObservers.isEmpty) {
      return;
    }
    for (final observer in _accountObservers) {
      observer.onAccountModified(modification);
    }
  }

  void _notifySettingsChanged(Settings settings) {
    final appSettings = ApplicationSettings.fromSettings(settings);

    for (final observer in _settingsObservers) {
      observer.onSettingsChanged(appSettings);
    }
    // for some user changes refresh servers list
    if (_shouldRefreshServersList(appSettings)) {
      _notifyServersListChanged();
    }
    _appSettings = appSettings;
  }

  bool _shouldRefreshServersList(ApplicationSettings appSettings) {
    return (_appSettings?.obfuscatedServers != appSettings.obfuscatedServers) ||
        (_appSettings?.virtualServers != appSettings.virtualServers) ||
        (_appSettings?.protocol != appSettings.protocol);
  }

  void _notifyVpnStatusChanged(StatusResponse state) async {
    if (_vpnObservers.isEmpty) {
      return;
    }

    for (final observer in _vpnObservers) {
      observer.onVpnStatusChanged(state);
    }
  }

  void _notifyServersListChanged() async {
    if (_serversListObservers.isEmpty) {
      return;
    }

    _vpnRepository.fetchServers().then((servers) {
      for (final observer in _serversListObservers) {
        observer.onServersListChanged(servers);
      }
    });
  }
}

@Riverpod(keepAlive: true)
AppStateChange appState(Ref ref) {
  final service = AppStateChange(
    ref.read(appStateRepositoryProvider),
    ref.read(vpnRepositoryProvider),
    ref.read(vpnSettingsProvider),
  );
  ref.onDispose(service.dispose);

  return service;
}
