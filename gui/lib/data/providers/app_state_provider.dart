import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/repository/app_state_repository.dart';
import 'package:nordvpn/data/models/recent_connections.dart';
import 'package:nordvpn/data/repository/vpn_repository.dart';
import 'package:nordvpn/data/repository/vpn_settings_repository.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/pb/daemon/servers.pb.dart';
import 'package:nordvpn/pb/daemon/state.pb.dart';
import 'package:nordvpn/pb/daemon/status.pb.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:nordvpn/constants.dart';

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

abstract class RecentConnectionsListObserver {
  void onRecentConnectionsListChanged(List<RecentConnection> recentConnections);
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
  final Set<RecentConnectionsListObserver> _recentConnectionsListObservers = {};

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

  void addRecentConnectionsListObserver(
    RecentConnectionsListObserver observer,
  ) {
    _recentConnectionsListObservers.add(observer);
  }

  void removeRecentConnectionsListObserver(
    RecentConnectionsListObserver observer,
  ) {
    _recentConnectionsListObservers.remove(observer);
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
            case UpdateEvent.RECENTS_LIST_UPDATE:
              _notifyRecentConnectionsListChanged();
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

  void _notifySettingsChanged(SettingsUpdate update) {
    final newSettings = ApplicationSettings.fromSettings(update.settings);

    for (final observer in _settingsObservers) {
      observer.onSettingsChanged(newSettings);
    }

    // Check if connection lists need refresh
    if (_shouldRefreshConnectionLists(newSettings, update.isResetToDefaults)) {
      _notifyServersListChanged();
      _notifyRecentConnectionsListChanged();
    }

    _appSettings = newSettings;
  }

  bool _shouldRefreshConnectionLists(
    ApplicationSettings newSettings,
    bool settingsWereReset,
  ) {
    if (settingsWereReset) {
      return true;
    }

    // If no previous settings exist, no need to refresh
    final currentSettings = _appSettings;
    if (currentSettings == null) {
      return false;
    }

    return _hasConnectionListsAffectingSettingsChanged(
      currentSettings,
      newSettings,
    );
  }

  bool _hasConnectionListsAffectingSettingsChanged(
    ApplicationSettings current,
    ApplicationSettings incoming,
  ) {
    return current.obfuscatedServers != incoming.obfuscatedServers ||
        current.virtualServers != incoming.virtualServers ||
        current.protocol != incoming.protocol;
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

  void _notifyRecentConnectionsListChanged() async {
    if (_recentConnectionsListObservers.isEmpty) {
      return;
    }

    _vpnRepository.fetchRecentConnections(maxRecentConnections).then((
      response,
    ) {
      final connections = response.connections
          .map((pb) => RecentConnection.fromPb(pb))
          .toList();

      for (final obs in _recentConnectionsListObservers) {
        obs.onRecentConnectionsListChanged(connections);
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
