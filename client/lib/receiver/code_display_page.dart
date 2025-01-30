import 'dart:async';

import 'package:client/lobby_page.dart';
import 'package:client/vertical_spacing.dart';
import 'package:client/web.dart';
import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:signals/signals_flutter.dart';

class CodeDisplayPage extends StatefulWidget {
  final String token;

  const CodeDisplayPage({super.key, required this.token});

  @override
  State<CodeDisplayPage> createState() => _CodeDisplayPageState();
}

class _CodeDisplayPageState extends State<CodeDisplayPage> with SignalsMixin {
  final showLoading = signal(false);
  final name = signal<String?>(null);
  final code = signal<String?>(null);
  bool success = false;
  Timer? _timer;

  @override
  void initState() {
    startCodeChecker();

    // Show the loading spinner after a second if there wasn't a code
    Timer(1.seconds, () {
      if (!success && name.peek() == null) {
        showLoading.value = true;
      }
    });

    super.initState();
  }

  void startCodeChecker() async {
    // Make a request to the server every 2 seconds to check for the code and if it has been entered correctly
    _timer = Timer.periodic(2.seconds, (timer) async {
      final (res, err) = await postRq("/receiver/check_state", {
        "token": widget.token,
      });
      if (err != null) {
        name.value = err.toString();
        showLoading.value = false;
        return;
      }
      if (res!.statusCode != null && res.statusCode != 200) {
        sendLog("error: ${res.statusCode}");
        name.value = "Received status code: ${res.statusCode} (${res.statusMessage})";
        showLoading.value = false;
        return;
      }

      // Get the output from the request
      final json = res.data;
      if (json["exists"]) {
        if (json["completed"]) {
          sendLog("challenge completed");
          return;
        }

        // Update all the state in one batch
        batch(() {
          showLoading.value = false;
          name.value = json["name"];
          code.value = json["code"];
        });
      } else {
        sendLog("waiting..");
      }
    });
  }

  @override
  void dispose() {
    sendLog("cancelling..");
    _timer?.cancel();
    super.dispose();
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
              (context) => Animate(
                effects: [
                  ExpandEffect(
                    curve: Curves.ease,
                    alignment: Alignment.center,
                    axis: Axis.vertical,
                  ),
                ],
                target: code.value != null ? 1 : 0,
                child: Padding(
                  padding: const EdgeInsets.only(bottom: sectionSpacing),
                  child: Text(code.value ?? "??????", style: Theme.of(context).textTheme.headlineMedium),
                ),
              ),
            ),
            Watch(
              (context) => Text(name.value ?? "Waiting for a connection..", style: Theme.of(context).textTheme.bodyLarge!.copyWith()),
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
