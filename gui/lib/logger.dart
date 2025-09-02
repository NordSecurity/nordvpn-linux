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

  logger = Logger(
    output: loggerOutput,
    filter: ProductionFilter(),
    level: kDebugMode ? Level.all : Level.info,
    printer: PrettyPrinter(
      colors: kDebugMode,
      printEmojis: kDebugMode,
      noBoxingByDefault: kDebugMode,
      methodCount: 1,
      dateTimeFormat: DateTimeFormat.dateAndTime,
    ),
  );
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
