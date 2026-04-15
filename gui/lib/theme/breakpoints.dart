final class AppBreakpoints {
  AppBreakpoints._();

  static const Breakpoint small = Breakpoint(
    beginWidth: 0,
    endWidth: 500,
    beginHeight: 532,
    endHeight: 532,
  );

  static const Breakpoint medium = Breakpoint(
    beginWidth: 500,
    endWidth: 700,
    beginHeight: 532,
    endHeight: 532,
  );

  static const Breakpoint large = Breakpoint(
    beginWidth: 700,
    endWidth: null,
    beginHeight: 532,
    endHeight: 532,
  );
}

final class Breakpoint {
  final double? beginWidth;

  final double? endWidth;

  final double? beginHeight;

  final double? endHeight;

  const Breakpoint({
    this.beginWidth,
    this.endWidth,
    this.beginHeight,
    this.endHeight,
  });
}
