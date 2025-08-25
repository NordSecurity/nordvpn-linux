import 'package:nordvpn/constants.dart';
import 'package:nordvpn/logger.dart';

DateTime? parseDate(String dateString) {
  try {
    return daemonDateFormat.parseStrict(dateString);
  } on Exception {
    logger.e("failed to parse date $dateString");
    return null;
  }
}
