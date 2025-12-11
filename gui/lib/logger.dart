import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:logger/logger.dart';
import 'package:path/path.dart' as path;

import './constants.dart';

late final Logger logger;
const int _oneMB = 1024 * 1024;

Future<void> setupLogger() async {
  LogOutput loggerOutput;

  if (kIsWeb) {
    loggerOutput = ConsoleOutput();
  } else {
    final cacheDir = await _getLinuxCacheDirectoryPath();
    final logDirPath = "$cacheDir/nordvpn";
    final logDir = await _ensureExists(logDirPath);
    final file = File("${logDir.path}/$logFile");

    // Trim the log file if it's already bigger than 10MB
    await _trimLogFileIfNeeded(
      file,
      maxLogFileSizeBytes: 10 * _oneMB,
      trimSizeBytes: 5 * _oneMB,
    );

    final fileOutput = FileOutput(file: file);
    loggerOutput = (kDebugMode)
        ? MultiOutput([ConsoleOutput(), fileOutput])
        : fileOutput;
  }

  final logLevel = giveLogLevel();
  logger = Logger(
    output: loggerOutput,
    filter: ProductionFilter(),
    level: logLevel,
    printer: LevelPrefixPrinter(
      PrettyPrinter(
        colors: true,
        printEmojis: false,
        noBoxingByDefault: true,
        methodCount: 1,
        dateTimeFormat: DateTimeFormat.dateAndTime,
      ),
    ),
  );

  logger.i("starting with log level: $logLevel");
}

Future<Directory> _ensureExists(String logDir) async {
  final directory = Directory(logDir);
  if (!await directory.exists()) {
    await directory.create(recursive: true);
  }
  return directory;
}

Future<String> _getLinuxCacheDirectoryPath() async {
  String homeDirPath = Platform.environment['HOME']!;
  return path.join(homeDirPath, '.cache');
}

Future<void> _trimLogFileIfNeeded(
  File file, {
  int maxLogFileSizeBytes = 10 * _oneMB,
  int trimSizeBytes = 5 * _oneMB,
}) async {
  if (await file.exists() && maxLogFileSizeBytes >= trimSizeBytes) {
    final int size = await file.length();
    if (size > maxLogFileSizeBytes) {
      final RandomAccessFile raf = await file.open(mode: FileMode.read);

      final int start = size - trimSizeBytes;
      await raf.setPosition(start);

      final Uint8List remainingBytes = await raf.read(trimSizeBytes);
      await raf.close();

      await file.writeAsBytes(remainingBytes, mode: FileMode.write);
    }
  }
}

Level giveLogLevel() {
  if (kDebugMode) return Level.all;
  final env = Platform.environment;
  final logLevelStr = env["NORDVPN_GUI_LOG_LEVEL"];
  return parseLogLevel(logLevelStr);
}

Level parseLogLevel(String? value) {
  return switch (value?.toLowerCase()) {
    "all" => Level.all,
    "trace" => Level.trace,
    "debug" => Level.debug,
    "info" => Level.info,
    "warn" => Level.warning,
    "error" => Level.error,
    "fatal" => Level.fatal,
    _ => Level.info,
  };
}

final class LevelPrefixPrinter extends LogPrinter {
  final LogPrinter _inner;

  LevelPrefixPrinter(this._inner);

  @override
  List<String> log(LogEvent event) {
    final levelStr = event.level.toString().split('.').last.toUpperCase();

    final lines = _inner.log(event);
    return [
      for (final line in lines)
        line.trim().isEmpty ? line : "[$levelStr] $line",
    ];
  }
}
