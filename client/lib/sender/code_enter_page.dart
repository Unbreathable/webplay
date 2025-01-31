import 'package:client/lobby_page.dart';
import 'package:client/sender/sender_connect_page.dart';
import 'package:client/vertical_spacing.dart';
import 'package:client/web.dart';
import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:signals/signals_flutter.dart';

class CodeEnterPage extends StatefulWidget {
  final String token;

  const CodeEnterPage({super.key, required this.token});

  @override
  State<CodeEnterPage> createState() => _CodeEnterPageState();
}

class _CodeEnterPageState extends State<CodeEnterPage> with SignalsMixin {
  final TextEditingController _controller = TextEditingController();
  final error = signal<String?>(null);

  void checkCode() async {
    // Make a request to create the receiver
    final (res, err) = await postRq("/sender/attempt", {
      "token": widget.token,
      "code": _controller.text,
    });
    if (err != null) {
      error.value = err.toString();
      return;
    }
    if (res!.statusCode != null && res.statusCode != 200) {
      sendLog("error: ${res.statusCode}");
      error.value = "Received status code: ${res.statusCode} (${res.statusMessage})";
      return;
    }

    popAllAndPush(context, MaterialPageRoute(builder: (context) => SenderConnectPage(token: widget.token)));
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            SizedBox(
              width: 300,
              child: TextField(
                decoration: InputDecoration(
                  hintText: "123456",
                ),
                controller: _controller,
              ),
            ),
            verticalSpacing(defaultSpacing),
            Watch(
              (context) => Animate(
                effects: [
                  ExpandEffect(
                    curve: Curves.ease,
                    alignment: Alignment.center,
                    axis: Axis.vertical,
                  ),
                ],
                target: error.value != null ? 1 : 0,
                child: Padding(
                  padding: const EdgeInsets.only(bottom: defaultSpacing),
                  child: Text(error.value ?? "Nothing went wrong, yet.", style: Theme.of(context).textTheme.headlineMedium),
                ),
              ),
            ),
            ElevatedButton(
              onPressed: () => checkCode(),
              child: Text("Check code"),
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
