import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/server_group_extension.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/data/providers/vpn_settings_controller.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/i18n/string_translation_extension.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/images_manager.dart';
import 'package:nordvpn/pb/daemon/config/group.pbenum.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/settings/settings_wrapper_widget.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/autoconnect_panel_theme.dart';
import 'package:nordvpn/vpn/servers_list_card.dart';
import 'package:nordvpn/widgets/custom_error_widget.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/loading_button.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';
import 'package:nordvpn/widgets/padded_circle_avatar.dart';

typedef AutoconnectLocation = ConnectArguments;

final class AutoconnectSettings extends ConsumerStatefulWidget {
  const AutoconnectSettings({super.key});

  @override
  ConsumerState<AutoconnectSettings> createState() =>
      AutoconnectSettingsState();
}

final class AutoconnectSettingsState
    extends ConsumerState<AutoconnectSettings> {
  bool _isAutoconnectSet = true;
  // clicked in GUI, not yet saved in daemon
  AutoconnectLocation? _clickedLocation;

  @override
  Widget build(BuildContext context) {
    final settingsProvider = ref.watch(vpnSettingsControllerProvider);
    return settingsProvider.when(
      loading: () => const LoadingIndicator(),
      error: (error, stackTrace) => CustomErrorWidget(message: "$error"),
      data: (settings) => _build(context, settings),
    );
  }

  Widget _build(BuildContext context, ApplicationSettings settings) {
    final appTheme = context.appTheme;
    return SingleChildSettingsWrapperWidget(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Padding(
            padding: EdgeInsets.only(
              left: appTheme.outerPadding,
              right: appTheme.outerPadding,
              bottom: appTheme.outerPadding,
            ),
            child: AutoconnectPanel(
              settings: settings,
              isUpdating: !_isAutoconnectSet,
              clickedLocation: _clickedLocation,
            ),
          ),
          Expanded(
            child: Padding(
              padding: EdgeInsets.only(
                left: appTheme.outerPadding,
                right: appTheme.outerPadding,
              ),
              child: ServersListCard(
                onSelected: (location) async {
                  if (_isLocationAlreadySet(settings, location)) {
                    return;
                  }
                  await _setAutoconnect(ref, location);
                },
                enabled: _isAutoconnectSet,
                allowServerNameSearch: false,
                withQuickConnectTile: true,
              ),
            ),
          ),
        ],
      ),
    );
  }

  bool _isLocationAlreadySet(
    ApplicationSettings settings,
    ConnectArguments args,
  ) {
    final autoConnectLocation = settings.autoConnectLocation;
    if (autoConnectLocation == null) return false;

    return autoConnectLocation.country == args.country &&
        autoConnectLocation.city == args.city &&
        autoConnectLocation.specialtyGroup == args.specialtyGroup;
  }

  Future<void> _setAutoconnect(
    WidgetRef ref,
    AutoconnectLocation selectedLocation,
  ) async {
    setState(() {
      _isAutoconnectSet = false;
      _clickedLocation = selectedLocation;
    });
    final vpnController = ref.read(vpnSettingsControllerProvider.notifier);
    await vpnController.setAutoConnect(true, selectedLocation);
    setState(() {
      _isAutoconnectSet = true;
      _clickedLocation = null;
    });
  }
}

final class AutoconnectPanel extends ConsumerWidget {
  final ApplicationSettings settings;
  final bool isUpdating;
  final AutoconnectLocation? clickedLocation;

  const AutoconnectPanel({
    super.key,
    required this.settings,
    this.isUpdating = false,
    this.clickedLocation,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final vpnStatusProvider = ref.watch(vpnStatusControllerProvider);
    return vpnStatusProvider.when(
      error: (error, _) => CustomErrorWidget(message: "$error"),
      loading: () => LoadingIndicator(),
      data: (vpnStatus) {
        final appTheme = context.appTheme;
        return Container(
          color: appTheme.area,
          child: Padding(
            padding: EdgeInsets.only(
              left: appTheme.outerPadding,
              right: appTheme.outerPadding,
              top: appTheme.padding,
              bottom: appTheme.padding,
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Expanded(
                  child: AutoconnectSelectionStatus(
                    vpnStatus: vpnStatus,
                    savedLocation: settings.autoConnectLocation,
                    clickedLocation: clickedLocation,
                    isUpdating: isUpdating,
                  ),
                ),
                _ConnectNowButton(
                  vpnStatus: vpnStatus,
                  savedLocation: settings.autoConnectLocation,
                  isUpdating: isUpdating,
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}

final class _ConnectNowButton extends ConsumerWidget {
  final VpnStatus vpnStatus;
  final AutoconnectLocation? savedLocation;
  final bool isUpdating;

  const _ConnectNowButton({
    required this.vpnStatus,
    required this.savedLocation,
    this.isUpdating = false,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final isEnabled =
        isUpdating || _isConnectedToSelectedLocation(vpnStatus, savedLocation);
    return LoadingElevatedButton(
      key: ValueKey(isEnabled),
      onPressed: isEnabled
          ? null
          : () => ref
                .read(vpnStatusControllerProvider.notifier)
                .connect(savedLocation),
      child: Text(t.ui.connectNow),
    );
  }
}

final class AutoconnectSelectionStatus extends StatelessWidget {
  final AutoconnectLocation? savedLocation;
  final VpnStatus vpnStatus;
  final AutoconnectLocation? clickedLocation;
  final bool isUpdating;

  const AutoconnectSelectionStatus({
    super.key,
    required this.vpnStatus,
    required this.savedLocation,
    this.clickedLocation,
    this.isUpdating = false,
  });

  @override
  Widget build(BuildContext context) {
    final theme = context.appTheme;
    return Row(
      children: [
        _VpnStatusIcon(
          vpnStatus: vpnStatus,
          savedLocation: savedLocation,
          isUpdating: isUpdating,
        ),
        SizedBox(width: theme.horizontalSpace),
        Expanded(
          child: AutoConnectServerInfo(
            vpnStatus: vpnStatus,
            savedLocation: savedLocation,
            isUpdating: isUpdating,
            clickedLocation: clickedLocation,
          ),
        ),
      ],
    );
  }
}

final class _VpnStatusIcon extends StatelessWidget {
  final ImagesManager imagesManager;
  final AutoconnectLocation? savedLocation;
  final VpnStatus vpnStatus;
  final bool isUpdating;

  _VpnStatusIcon({
    required this.savedLocation,
    required this.vpnStatus,
    this.isUpdating = false,
    ImagesManager? imagesManager,
  }) : imagesManager = imagesManager ?? sl();

  @override
  Widget build(BuildContext context) {
    final autoconnectPanelTheme = context.autoconnectPanelTheme;
    final appTheme = context.appTheme;

    if (isUpdating || vpnStatus.isConnecting()) {
      return Padding(
        padding: EdgeInsets.all(appTheme.flagsBorderSize),
        child: LoadingIndicator(size: autoconnectPanelTheme.loaderSize),
      );
    }

    return PaddedCircleAvatar(
      size: autoconnectPanelTheme.iconSize,
      borderColor: _isConnectedToSelectedLocation(vpnStatus, savedLocation)
          ? appTheme.successColor
          : Colors.transparent,
      borderSize: appTheme.flagsBorderSize,
      child: _selectedServerIcon(),
    );
  }

  Widget _selectedServerIcon() {
    // specific location selected
    if (savedLocation?.country != null) {
      return imagesManager.forCountry(savedLocation!.country!);
    }

    // only specialty group selected without location
    if (savedLocation?.specialtyGroup != null) {
      return imagesManager.forSpecialtyServer(savedLocation!.specialtyGroup!);
    }

    return DynamicThemeImage("fastest_server.svg");
  }
}

bool _isConnectedToSelectedLocation(
  VpnStatus vpnStatus,
  AutoconnectLocation? location,
) {
  if (!vpnStatus.isConnected()) return false;

  final connParams = vpnStatus.connectionParameters;
  // connected to default - Fastest server (Quick Connect)
  if (location == null) {
    return connParams.country.isEmpty &&
        connParams.city.isEmpty &&
        (connParams.group == ServerGroup.UNDEFINED ||
            connParams.group == ServerGroup.STANDARD_VPN_SERVERS);
  }

  // connected to something else

  // `?? ""` here is because `connParams.country` is empty string when not set
  // while `location.countryCode?.code` is `null` when not set, but both cases
  // mean "not set". Similarly for city name.
  final countryMatches = (location.country?.code ?? "") == connParams.country;
  final cityMatches = (location.city?.name ?? "") == connParams.city;
  final groupMatches =
      location.specialtyGroup == connParams.group.toSpecialtyType();

  return countryMatches && cityMatches && groupMatches;
}

final class AutoConnectServerInfo extends StatelessWidget {
  final VpnStatus vpnStatus;
  final AutoconnectLocation? savedLocation;
  final bool isUpdating;
  final AutoconnectLocation? clickedLocation;

  const AutoConnectServerInfo({
    super.key,
    required this.vpnStatus,
    required this.savedLocation,
    this.isUpdating = false,
    this.clickedLocation,
  });

  @override
  Widget build(BuildContext context) {
    final theme = context.autoconnectPanelTheme;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (isUpdating && clickedLocation != null)
          ..._updatingInProgressLabel(theme),
        if (!isUpdating && savedLocation != null)
          ..._selectedServerLabel(theme),
        if (!isUpdating && savedLocation == null)
          ..._defaultFastestServer(theme),
      ],
    );
  }

  List<Widget> _updatingInProgressLabel(AutoconnectPanelTheme theme) {
    assert(clickedLocation != null);

    var connectionTarget = "";
    if (clickedLocation?.specialtyGroup != null) {
      // connecting to specialty server
      final specialtyGroup = clickedLocation?.specialtyGroup!;
      final serverLabel = labelForServerType(specialtyGroup!);
      connectionTarget = serverLabel;
    } else if (clickedLocation?.country?.name != null) {
      // connecting to regular server
      connectionTarget = clickedLocation!.country!.name;
    } else {
      // quick connect
      connectionTarget = fastestServerLabel;
    }

    return [
      Text(
        t.ui.settingAutoconnectTo(target: connectionTarget),
        style: theme.primaryFont,
        overflow: TextOverflow.ellipsis,
      ),
    ];
  }

  List<Widget> _selectedServerLabel(AutoconnectPanelTheme theme) {
    if (savedLocation?.specialtyGroup == null &&
        savedLocation?.country != null) {
      // regular server selected - show:
      // ┌──────┐
      // │      │    Country name
      // │ flag │
      // │      │    City name
      // └──────┘
      return _countryNameAndCityName(theme);
    }

    if (savedLocation?.specialtyGroup != null &&
        savedLocation?.country != null) {
      // specialty server selected with specific location, city name optional - show:
      // ┌──────┐
      // │      │    Server group
      // │ flag │
      // │      │    Country name [- City name]
      // └──────┘
      return _serverGroupWithCountryAndCityName(theme);
    }

    // specialty server selected without specific location - show:
    // ┌──────┐
    // │      │    Server group
    // │ flag │
    // │      │    "Fastest Server"
    // └──────┘
    return _serverGroupAndFastestServer(theme);
  }

  List<Widget> _countryNameAndCityName(AutoconnectPanelTheme theme) {
    final countryName = savedLocation?.country?.localizedName ?? "";
    var cityName = savedLocation?.city?.localizedName ?? "";

    if (cityName.isNotEmpty) {
      cityName += savedLocation?.server?.isVirtual == true ? "- Virtual" : "";
    }

    return [
      Text(
        countryName,
        style: theme.primaryFont,
        overflow: TextOverflow.ellipsis,
      ),
      if (cityName.isNotEmpty)
        Text(
          cityName,
          style: theme.secondaryFont,
          overflow: TextOverflow.ellipsis,
        ),
    ];
  }

  List<Widget> _serverGroupWithCountryAndCityName(AutoconnectPanelTheme theme) {
    assert(savedLocation?.specialtyGroup != null);
    final countryName = savedLocation?.country?.localizedName ?? "";
    final cityName = savedLocation?.city?.localizedName ?? "";
    final location = countryName + (cityName.isNotEmpty ? " - $cityName" : "");

    return [
      Text(
        labelForServerType(savedLocation!.specialtyGroup!),
        style: theme.primaryFont,
        overflow: TextOverflow.ellipsis,
      ),
      Text(
        location,
        style: theme.secondaryFont,
        overflow: TextOverflow.ellipsis,
      ),
    ];
  }

  List<Widget> _serverGroupAndFastestServer(AutoconnectPanelTheme theme) {
    return [
      Text(
        labelForServerType(savedLocation!.specialtyGroup!),
        style: theme.primaryFont,
        overflow: TextOverflow.ellipsis,
      ),
      Text(
        t.ui.fastestServer,
        style: theme.secondaryFont,
        overflow: TextOverflow.ellipsis,
      ),
    ];
  }

  List<Widget> _defaultFastestServer(AutoconnectPanelTheme theme) {
    return [
      Text(
        fastestServerLabel,
        style: theme.primaryFont,
        overflow: TextOverflow.ellipsis,
      ),
    ];
  }
}
