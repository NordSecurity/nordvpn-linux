import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/models/servers_list.dart';
import 'package:nordvpn/data/providers/account_controller.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/providers/servers_list_controller.dart';
import 'package:nordvpn/data/providers/vpn_settings_controller.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/images_manager.dart';
import 'package:nordvpn/internal/popup_codes.dart';
import 'package:nordvpn/router/routes.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/i18n/string_translation_extension.dart';
import 'package:nordvpn/theme/servers_list_theme.dart';
import 'package:nordvpn/vpn/server_list_item_factory.dart';
import 'package:nordvpn/widgets/custom_error_widget.dart';
import 'package:nordvpn/widgets/custom_expansion_tile.dart';
import 'package:nordvpn/widgets/dialog_factory.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';
import 'package:nordvpn/widgets/searchable_servers_list.dart';
import 'package:nordvpn/vpn/recent_connections_list.dart';

// ServersListCard displays the list of servers from the VPN screen
final class ServersListCard extends StatefulWidget {
  final ImagesManager imagesManager;
  final ServerListItemFactory itemFactory;
  final Future<void> Function(ConnectArguments) onSelected;
  final bool enabled;
  final bool allowServerNameSearch;
  final bool withQuickConnectTile;
  final bool withRecentConnectionsWidget;

  ServersListCard({
    super.key,
    required this.onSelected,
    ImagesManager? imagesManager,
    ServerListItemFactory? itemFactory,
    this.enabled = true,
    this.allowServerNameSearch = true,
    this.withQuickConnectTile = false,
    this.withRecentConnectionsWidget = false,
  }) : imagesManager = imagesManager ?? sl(),
       itemFactory = itemFactory ?? sl();

  @override
  State<ServersListCard> createState() => _ServersListCardState();
}

final class ServersListKeys {
  ServersListKeys._();
  static const searchKey = ValueKey('servers-list-search-key');
  static const countriesServersListKey = ValueKey('countries-servers-list-key');
  static const specialtyServerKey = PageStorageKey('specialty');
}

final class _ServersListCardState extends State<ServersListCard> {
  bool _showSearchView = false;

  // controls the content for the search bar from the servers list.
  // declaring it here ensures that the searched text is lost if the servers
  // list is updated from notification
  final _searchTextController = TextEditingController();

  @override
  Widget build(BuildContext context) {
    return Consumer(
      builder: (context, ref, _) {
        return ref
            .watch(serversListControllerProvider)
            .when(
              loading: () => const LoadingIndicator(),
              error: (_, _) => _buildError(context, ref),
              data: (serversList) {
                return Opacity(
                  opacity: widget.enabled ? 1.0 : 0.5,
                  child: _buildCurrentView(context, serversList, ref),
                );
              },
            );
      },
    );
  }

  @override
  void dispose() {
    super.dispose();
    _searchTextController.dispose();
  }

  Widget _buildError(BuildContext context, WidgetRef ref) {
    final message = t.ui.failedToLoadService;
    return CustomErrorWidget(
      message: message,
      buttonText: t.ui.retry,
      onPressed: () async {
        await ref.read(serversListControllerProvider.notifier).refetch();
      },
    );
  }

  Widget _buildCurrentView(
    BuildContext context,
    ServersList serversList,
    WidgetRef ref,
  ) {
    final isObfuscationEnabled =
        serversList.standardServersList.isEmpty &&
        serversList.obfuscatedServersList.isNotEmpty;

    return (!_showSearchView)
        ? _buildTabBarView(context, serversList, ref, isObfuscationEnabled)
        : _buildSearchList(context, serversList, ref, isObfuscationEnabled);
  }

  DefaultTabController _buildTabBarView(
    BuildContext context,
    ServersList serversList,
    WidgetRef ref,
    bool isObfuscationEnabled,
  ) {
    return DefaultTabController(
      length: 2,
      child: Column(
        spacing: context.appTheme.verticalSpaceSmall,
        children: [
          if (isObfuscationEnabled)
            _showObfuscatedMessage(context, t.ui.turnOffObfuscationLocations),
          if (widget.withRecentConnectionsWidget)
            RecentConnectionsList(onSelected: widget.onSelected),
          Expanded(
            child: NestedScrollView(
              headerSliverBuilder:
                  (BuildContext context, bool innerBoxIsScrolled) {
                    final handle =
                        NestedScrollView.sliverOverlapAbsorberHandleFor(
                          context,
                        );
                    return [
                      SliverOverlapAbsorber(
                        handle: handle,
                        sliver: SliverPersistentHeader(
                          delegate: _SliverAppBarDelegate(
                            TabBar(
                              isScrollable: true,
                              tabs: [
                                Tab(text: t.ui.countries),
                                Tab(text: t.ui.specialServers),
                              ],
                            ),
                            context.appTheme,
                            () => setState(() => _showSearchView = true),
                          ),
                          pinned: true,
                          floating: true,
                        ),
                      ),
                    ];
                  },
              body: _buildTabsWithServers(
                serversList,
                ref,
                isObfuscationEnabled,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTabsWithServers(
    ServersList serversList,
    WidgetRef ref,
    bool isObfuscationEnabled,
  ) {
    return TabBarView(
      children: [
        _buildServersList(serversList, ref, isObfuscationEnabled),
        _buildSpecialtyServersList(serversList, ref, isObfuscationEnabled),
      ],
    );
  }

  // Builds the countries servers list from the tabbar
  Widget _buildServersList(
    ServersList serversList,
    WidgetRef ref,
    bool isObfuscationEnabled,
  ) {
    final servers = isObfuscationEnabled
        ? serversList.obfuscatedServersList
        : serversList.standardServersList;

    // count additional quick connect tile if specified
    final itemsCount = servers.length + (widget.withQuickConnectTile ? 1 : 0);
    final appTheme = context.appTheme;

    return Builder(
      builder: (context) {
        return CustomScrollView(
          primary: false,
          key: ServersListKeys.countriesServersListKey,
          slivers: [
            SliverOverlapInjector(
              handle: NestedScrollView.sliverOverlapAbsorberHandleFor(context),
            ),
            SliverList(
              delegate: SliverChildBuilderDelegate((context, index) {
                // show additional quick connect tile if specified
                if (widget.withQuickConnectTile && index == 0) {
                  return CustomExpansionTile(
                    leading: DynamicThemeImage("fastest_server.svg"),
                    title: Text(fastestServerLabel, style: appTheme.body),
                    onTap: () async => await widget.onSelected(
                      ConnectArguments(
                        specialtyGroup: isObfuscationEnabled
                            ? ServerType.obfuscated
                            : null,
                      ),
                    ),
                  );
                }
                // adjust for additional quick connect tile if specified
                final idx = index - (widget.withQuickConnectTile ? 1 : 0);
                return widget.itemFactory.forCountry(
                  context: context,
                  country: servers[idx],
                  onTap: (args) async => await widget.onSelected(args),
                  enabled: widget.enabled,
                  specialtyGroup: isObfuscationEnabled
                      ? ServerType.obfuscated
                      : null,
                );
              }, childCount: itemsCount),
            ),
          ],
        );
      },
    );
  }

  // Show the specialty servers in the tabbar
  Widget _buildSpecialtyServersList(
    ServersList serversList,
    WidgetRef ref,
    bool isObfuscatedOn,
  ) {
    final specialtyServersOrder = [
      (type: ServerType.dedicatedIP, description: t.ui.getYourDip),
      (type: ServerType.doubleVpn, description: t.ui.doubleVpnDesc),
      (type: ServerType.onionOverVpn, description: t.ui.onionOverVpnDesc),
      (type: ServerType.p2p, description: t.ui.p2pDesc),
      (type: ServerType.europe, description: t.ui.europe),
      (type: ServerType.asiaPacific, description: t.ui.asiaPacific),
      (type: ServerType.theAmericas, description: t.ui.theAmericas),
      (
        type: ServerType.africaTheMiddleEastAndIndia,
        description: t.ui.africaTheMiddleEastAndIndia,
      ),
    ];
    return Builder(
      builder: (context) {
        return CustomScrollView(
          primary: false,
          key: ServersListKeys.specialtyServerKey,
          slivers: [
            SliverOverlapInjector(
              handle: NestedScrollView.sliverOverlapAbsorberHandleFor(context),
            ),
            SliverList(
              delegate: SliverChildBuilderDelegate((context, index) {
                final group = specialtyServersOrder[index];
                final type = group.type;
                final description = group.description;
                if (type == ServerType.dedicatedIP) {
                  return _buildDipListItem(ref, serversList, isObfuscatedOn);
                }
                final servers = serversList.specialtyServersList(type);
                return widget.itemFactory.forSpecialtyServer(
                  context: context,
                  type: type,
                  enabled:
                      (servers.isNotEmpty || group.type.isRegion) &&
                      !isObfuscatedOn,
                  servers: servers,
                  subtitle: description,
                  onTap: (args) => widget.onSelected(args),
                  showDetails: () => _showDetailsForSpecialtyServer(
                    context: context,
                    ref: ref,
                    type: type,
                    servers: servers,
                  ),
                );
              }, childCount: specialtyServersOrder.length),
            ),
          ],
        );
      },
    );
  }

  // Build the list item for dedicated IP
  Widget _buildDipListItem(
    WidgetRef ref,
    ServersList serversList,
    bool isObfuscatedOn,
  ) {
    return Consumer(
      builder: (context, ref, child) {
        final accountProvider = ref.watch(accountControllerProvider);

        if (accountProvider case AsyncData(:final value) when value != null) {
          final accountInfo = value;
          final settings = ref.watch(vpnSettingsControllerProvider);

          if (settings case AsyncData(:final value)) {
            final settings = value;
            var subtitle = t.ui.getYourDip;
            if (accountInfo.hasDipSubscription) {
              final dipServers = accountInfo.dedicatedIpServers ?? [];
              if (dipServers.isNotEmpty) {
                // construct the subtitle
                int count = 0;
                subtitle = "";
                for (final countryGroup in accountInfo.dedicatedIpServers!) {
                  for (final city in countryGroup.cities) {
                    if (count > 0) {
                      subtitle += ", ";
                    }
                    subtitle +=
                        "${countryGroup.countryName} - ${city.localizedName}";
                    count += 1;
                  }
                }
              } else {
                subtitle = t.ui.selectServerForDip;
              }
            }

            return widget.itemFactory.forSpecialtyServer(
              context: context,
              enabled: settings.areDipServersSupported(),
              type: ServerType.dedicatedIP,
              servers: accountInfo.dedicatedIpServers ?? [],
              subtitle: subtitle,
              onTap: (args) {
                if (!accountInfo.hasDipSubscription) {
                  ref
                      .read(popupsProvider.notifier)
                      .show(PopupCodes.getDedicatedIp);
                  return;
                }
                if (!accountInfo.hasDipServers) {
                  ref.read(popupsProvider.notifier).show(PopupCodes.chooseDip);
                  return;
                }
                widget.onSelected(args);
              },
              showDetails: () => _showDetailsForSpecialtyServer(
                context: context,
                ref: ref,
                type: ServerType.dedicatedIP,
                servers: accountInfo.dedicatedIpServers!,
              ),
            );
          }
        }

        return const SizedBox.shrink();
      },
    );
  }

  // shows the popup with the servers when ... button is pressed,
  // for a specialty group
  void _showDetailsForSpecialtyServer({
    required BuildContext context,
    required WidgetRef ref,
    required ServerType type,
    required List<CountryServersGroup> servers,
  }) {
    // If there are more than 2 cities or more than 1 city in a country,
    // then show the view with the search.
    // Because if a country has more than 1 server it can be expanded,
    // changing the view height => change of size
    var numberOfCities = 0;
    for (final item in servers) {
      numberOfCities += item.cities.length;
      if (numberOfCities > 2 || item.cities.length > 1) {
        // set some big value to be sure it will display the search list view
        // and the Quick Connect button
        numberOfCities = 100;
        break;
      }
    }

    final shouldShowStaticList = (numberOfCities <= 2);

    DialogFactory.showPopover(
      context: context,
      icon: widget.imagesManager.forSpecialtyServer(type),
      title: labelForServerType(type),
      showDivider: shouldShowStaticList,
      buttonTitle: shouldShowStaticList || widget.withQuickConnectTile
          ? ""
          : t.ui.quickConnect,
      stretchButton: true,
      child: SearchableServersList(
        servers: servers,
        allowServerNameSearch: true,
        specialtyServer: type,
        onTap: (args) {
          DialogFactory.close(context);
          widget.onSelected(args);
        },
        withQuickConnectTile: widget.withQuickConnectTile,
      ),
      onButtonClicked: () {
        // quick connect is pressed
        widget.onSelected(ConnectArguments(specialtyGroup: type));
      },
    );
  }

  // Build the list when searching all the servers
  Widget _buildSearchList(
    BuildContext context,
    ServersList serversList,
    WidgetRef ref,
    bool isObfuscationEnabled,
  ) {
    return SearchableServersList.forServersList(
      leadingWidget: IconButton(
        icon: DynamicThemeImage("back_arrow.svg"),
        onPressed: () {
          setState(() {
            _showSearchView = false;
            _searchTextController.clear();
          });
        },
      ),
      serversList: serversList,
      searchTextController: _searchTextController,
      specialtyServer: isObfuscationEnabled ? ServerType.obfuscated : null,
      onTap: (args) => widget.onSelected(args),
      allowServerNameSearch: widget.allowServerNameSearch,
    );
  }

  Widget _showObfuscatedMessage(BuildContext context, String message) {
    final appTheme = context.appTheme;
    final serversListTheme = context.serversListTheme;
    return Container(
      padding: EdgeInsets.symmetric(
        horizontal: appTheme.horizontalSpace,
        vertical: appTheme.verticalSpaceVerySmall,
      ),
      color: serversListTheme.obfuscatedItemBackgroundColor,
      child: Row(
        spacing: appTheme.horizontalSpace,
        children: [
          Expanded(child: Text(message, style: appTheme.body)),
          TextButton(
            onPressed: () =>
                context.navigateToRoute(AppRoute.settingsSecurityAndPrivacy),
            style: ButtonStyle(
              padding: WidgetStateProperty.all(const EdgeInsets.only(right: 4)),
            ),
            child: Text(t.ui.goToSettings),
          ),
        ],
      ),
    );
  }
}

class _SliverAppBarDelegate extends SliverPersistentHeaderDelegate {
  _SliverAppBarDelegate(this._tabBar, this._theme, this._onSearchPressed);

  final TabBar _tabBar;
  final AppTheme _theme;
  final VoidCallback _onSearchPressed;

  @override
  double get minExtent => kToolbarHeight;
  @override
  double get maxExtent => kToolbarHeight;

  @override
  Widget build(
    BuildContext context,
    double shrinkOffset,
    bool overlapsContent,
  ) {
    return Container(
      color: _theme.backgroundColor,
      child: Column(
        children: [
          Row(
            children: [
              Expanded(child: _tabBar),
              Padding(
                padding: EdgeInsets.only(right: _theme.padding),
                child: IconButton(
                  key: ServersListKeys.searchKey,
                  tooltip: 'Search',
                  icon: DynamicThemeImage("search.svg"),
                  onPressed: _onSearchPressed,
                ),
              ),
            ],
          ),
          Divider(height: 1, thickness: 1, color: _theme.dividerColor),
        ],
      ),
    );
  }

  @override
  bool shouldRebuild(_SliverAppBarDelegate oldDelegate) {
    return false;
  }
}
