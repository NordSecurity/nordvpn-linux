final class ConfigImpl implements Config {
  final Duration loginTimeoutDuration;

  ConfigImpl({required this.loginTimeoutDuration});

  @override
  Duration get loginTimeout => loginTimeoutDuration;
}

abstract class Config {
  Duration get loginTimeout;
}
