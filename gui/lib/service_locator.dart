import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:get_it/get_it.dart';
import 'package:nordvpn/config.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/grpc/error_handling_interceptor.dart';
import 'package:nordvpn/grpc/grpc_service.dart';
import 'package:nordvpn/i18n/country_names_service.dart';
import 'package:nordvpn/internal/assets_manager.dart';
import 'package:nordvpn/internal/images_manager.dart';
import 'package:nordvpn/vpn/server_list_item_factory.dart';
import 'package:shared_preferences/shared_preferences.dart';

final sl = GetIt.instance;

Future<void> initServiceLocator() async {
  assert(!sl.isRegistered<ImagesManager>(), "service locator must be empty");
  sl.registerSingleton<ImagesManager>(ImagesManager());
  sl.registerSingleton<ServerListItemFactory>(
    ServerListItemFactory(imagesManager: sl()),
  );

  final assetManager = await AssetManager.create();
  sl.registerSingleton<AssetManager>(assetManager);

  sl.registerSingleton<Config>(
    ConfigImpl(loginTimeoutDuration: loginTimeoutDuration),
  );

  sl.registerSingleton<SharedPreferencesAsync>(SharedPreferencesAsync());
  sl.registerSingleton(CountryNamesService());

  sl.registerSingleton(
    ProviderContainer(),
    dispose: (container) => container.dispose(),
  );
  sl.registerSingleton(
    createNewChannel(),
    dispose: (channel) async => await channel.terminate(),
  );
  sl.registerSingleton(ErrorHandlingInterceptor());
}
