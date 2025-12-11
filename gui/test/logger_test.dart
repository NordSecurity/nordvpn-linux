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

  group('LevelPrefixPrinter', () {
    test('prefixes single non-empty line with uppercased level', () {
      final inner = _FakePrinter(['hello world']);
      final printer = LevelPrefixPrinter(inner);

      final event = LogEvent(Level.info, 'hello world');
      final result = printer.log(event);

      expect(result, ['[INFO] hello world']);
    });

    test('prefixes multiple non-empty lines', () {
      final inner = _FakePrinter(['line 1', 'line 2']);
      final printer = LevelPrefixPrinter(inner);

      final event = LogEvent(Level.debug, 'multi');
      final result = printer.log(event);

      expect(result, ['[DEBUG] line 1', '[DEBUG] line 2']);
    });

    test('does not modify empty or whitespace-only lines', () {
      final inner = _FakePrinter(['', '   ', '\t', 'msg']);
      final printer = LevelPrefixPrinter(inner);

      final event = LogEvent(Level.warning, 'test');
      final result = printer.log(event);

      expect(result, ['', '   ', '\t', '[WARNING] msg']);
    });

    test('handles different levels correctly', () {
      final inner = _FakePrinter(['x']);
      final printer = LevelPrefixPrinter(inner);

      expect(printer.log(LogEvent(Level.all, 'x')), ['[ALL] x']);
      expect(printer.log(LogEvent(Level.trace, 'x')), ['[TRACE] x']);
      expect(printer.log(LogEvent(Level.debug, 'x')), ['[DEBUG] x']);
      expect(printer.log(LogEvent(Level.info, 'x')), ['[INFO] x']);
      expect(printer.log(LogEvent(Level.warning, 'x')), ['[WARNING] x']);
      expect(printer.log(LogEvent(Level.error, 'x')), ['[ERROR] x']);
      expect(printer.log(LogEvent(Level.fatal, 'x')), ['[FATAL] x']);
    });
  });
}

final class _FakePrinter extends LogPrinter {
  _FakePrinter(this._lines);

  final List<String> _lines;

  @override
  List<String> log(LogEvent event) => _lines;
}
