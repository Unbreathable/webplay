import 'dart:async';

import 'package:client/lobby_page.dart';
import 'package:client/sender/code_enter_page.dart';
import 'package:client/vertical_spacing.dart';
import 'package:client/web.dart';
import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:signals/signals_flutter.dart';

class SenderCreatePage extends StatefulWidget {
  const SenderCreatePage({super.key});

  @override
  State<SenderCreatePage> createState() => _SenderCreatePageState();
}

class _SenderCreatePageState extends State<SenderCreatePage> with SignalsMixin {
  final showLoading = signal(false);
  final error = signal<String?>(null);
  bool success = false;

  @override
  void initState() {
    startReceiver();

    // Show the loading spinner after a second if there wasn't a code
    Timer(1.seconds, () {
      if (!success && error.peek() == null) {
        showLoading.value = true;
      }
    });

    super.initState();
  }

  void startReceiver() async {
    // Make a request to create the receiver
    final (res, err) = await postRq("/sender/create", {
      "name": "test_client",
    });
    if (err != null) {
      error.value = err.toString();
      showLoading.value = false;
      return;
    }
    if (res!.statusCode != null && res.statusCode != 200) {
      sendLog("error: ${res.statusCode}");
      error.value = "Received status code: ${res.statusCode} (${res.statusMessage})";
      showLoading.value = false;
      return;
    }

    // Get the output from the request and go to display page
    final json = res.data;
    popAllAndPush(context, MaterialPageRoute(builder: (context) => CodeEnterPage(token: json["token"])));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Watch(
              (context) => Animate(
                effects: [
                  ExpandEffect(
                    curve: Curves.ease,
                    alignment: Alignment.center,
                    axis: Axis.vertical,
                  ),
                ],
                target: showLoading.value ? 1 : 0,
                child: Padding(
                  padding: const EdgeInsets.only(bottom: sectionSpacing),
                  child: SizedBox(
                    height: 54,
                    width: 54,
                    child: Center(
                      child: SizedBox(
                        width: 50,
                        height: 50,
                        child: CircularProgressIndicator(
                          value: null,
                        ),
                      ),
                    ),
                  ),
                ),
              ),
            ),
            Watch(
              (context) => Text(error.value ?? "Waiting for token..", style: Theme.of(context).textTheme.bodyLarge!.copyWith()),
            ),
            verticalSpacing(sectionSpacing),
            ElevatedButton(
              onPressed: () => popAllAndPush(context, MaterialPageRoute(builder: (context) => LobbyPage())),
              child: Text("Return to Lobby"),
            ),
          ],
        ),
      ),
    );
  }
}
