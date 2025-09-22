// TODO: fix later
void main() async {
  // setUpAll(() async => await initServiceLocator());
  //   testWidgets("VpnStatusCard check UI elements", (WidgetTester tester) async {
  //     await tester.mockAppAndProviders(VpnStatusCard());
  //     await tester.pumpUntilFound(find.text(t.ui.notConnected),
  //         timeout: Duration(seconds: 120));

  //     checkDisconnectedState() {
  //       expect(tester.widget<Text>(find.text(t.ui.notConnected)).style!.color,
  //           UXColors.textCaution);
  //       expect(find.text(t.ui.notConnected), findsOneWidget);
  //       expect(find.text(t.ui.connectToVpn), findsOneWidget);
  //       expect(find.text(t.ui.quickConnect), findsOneWidget);
  //       expect(tester.findSvgWithPath("vpn_not_connected.svg"), findsOneWidget);
  //     }

  //     // initial not connected state
  //     checkDisconnectedState();

  //     // connecting state
  //     await tester.tap(find.text(t.ui.quickConnect));
  //     await tester.pumpUntilFound(find.text(t.ui.cancel));
  //     expect(find.text(t.ui.findingServer), findsOneWidget);
  //     expect(find.byType(CircularProgressIndicator), findsOneWidget);
  //     expect(
  //       tester.widget<Text>(find.textContaining(t.ui.connecting)).style!.color,
  //       UXColors.textCaution,
  //     );

  //     // connected state
  //     await tester.pumpUntilFound(find.text(t.ui.disconnect));
  //     expect(tester.widget<Text>(find.text(t.ui.connected)).style!.color,
  //         UXColors.textSuccess);
  //     expect(tester.findSvgWithPath("flags/fr.svg"), findsOneWidget);
  //     expect(tester.findSvgWithPath("reconnect.svg"), findsOneWidget);

  //     // switch to disconnected
  //     await tester.tap(find.text(t.ui.disconnect));
  //     await tester.pumpUntilFound(find.text(t.ui.quickConnect));
  //     checkDisconnectedState();
  //     await tester.pumpAndSettleWithTimeout(duration: Duration(seconds: 1));
  //   });
}
