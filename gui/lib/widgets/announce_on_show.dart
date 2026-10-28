import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';

// AnnounceOnShow is a wrapping widget that makes a one-time announcement when its child widget appears.
//
// Announces the provided announcement once, when this widget is first mounted.
// If its route is still transitioning in, the announcement waits for the transition
// to complete. It is skipped if the route never completes, is no longer the
// current route, or the platform doesn't support announcements.
final class AnnounceOnShow extends StatefulWidget {
  final ValueGetter<String> announcement;
  final Widget child;

  const AnnounceOnShow({
    super.key,
    required this.announcement,
    required this.child,
  });

  @override
  State<AnnounceOnShow> createState() => _AnnounceOnShowState();
}

final class _AnnounceOnShowState extends State<AnnounceOnShow> {
  Animation<double>? _transition;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback(
      (_) => _scheduleAnnouncement(),
    );
  }

  // Wait for the route to finish appearing: announcing while the route is still
  // animating competes with the speech for the route change itself.
  void _scheduleAnnouncement() {
    if (!mounted) return;

    // No route to wait for when built directly, as in tests.
    final transition = ModalRoute.of(context)?.animation;
    if (transition == null || transition.isCompleted) {
      _announce();
      return;
    }

    _transition = transition..addStatusListener(_onTransitionStatus);
  }

  void _onTransitionStatus(AnimationStatus status) {
    if (status != AnimationStatus.completed) return;

    _stopListening();
    _announce();
  }

  void _stopListening() {
    _transition?.removeStatusListener(_onTransitionStatus);
    _transition = null;
  }

  void _announce() {
    if (!mounted || !MediaQuery.supportsAnnounceOf(context)) return;
    if (ModalRoute.isCurrentOf(context) == false) return;

    SemanticsService.sendAnnouncement(
      View.of(context),
      widget.announcement(),
      Directionality.of(context),
    );
  }

  @override
  void dispose() {
    _stopListening();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => widget.child;
}
