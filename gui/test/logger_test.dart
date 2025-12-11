import 'package:flutter_test/flutter_test.dart';
import 'package:logger/logger.dart';
import 'package:nordvpn/logger.dart';

void main() {
  group('parseLogLevel', () {
    test('returns all for "all"', () {
      expect(parseLogLevel('all'), Level.all);
    });

    test('returns trace for "trace"', () {
      expect(parseLogLevel('trace'), Level.trace);
    });

    test('returns debug for "debug"', () {
      expect(parseLogLevel('debug'), Level.debug);
    });

    test('returns info for "info"', () {
      expect(parseLogLevel('info'), Level.info);
    });

    test('returns warn for "warn"', () {
      expect(parseLogLevel('warn'), Level.warning);
    });

    test('returns error for "error"', () {
      expect(parseLogLevel('error'), Level.error);
    });

    test('returns fatal for "fatal"', () {
      expect(parseLogLevel('fatal'), Level.fatal);
    });

    test('is case-insensitive', () {
      expect(parseLogLevel('DEBUG'), Level.debug);
      expect(parseLogLevel('Info'), Level.info);
      expect(parseLogLevel('ErRoR'), Level.error);
    });

    test('returns default info for null', () {
      expect(parseLogLevel(null), Level.info);
    });

    test('returns default info for unknown value', () {
      expect(parseLogLevel('verbose'), Level.info);
      expect(parseLogLevel(''), Level.info);
      expect(parseLogLevel('   '), Level.info);
    });
  });
}
