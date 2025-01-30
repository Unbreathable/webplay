import 'package:client/receiver/receiver_create_page.dart';
import 'package:client/sender/sender_create_page.dart';
import 'package:client/vertical_spacing.dart';
import 'package:flutter/material.dart';

class LobbyPage extends StatefulWidget {
  const LobbyPage({super.key});

  @override
  State<LobbyPage> createState() => _LobbyPageState();
}

class _LobbyPageState extends State<LobbyPage> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ElevatedButton(
              onPressed: () => popAllAndPush(context, MaterialPageRoute(builder: (context) => ReceiverCreatePage())),
              child: Text("Start receiver"),
            ),
            verticalSpacing(defaultSpacing),
            ElevatedButton(
              onPressed: () => popAllAndPush(context, MaterialPageRoute(builder: (context) => SenderCreatePage())),
              child: Text("Start sender"),
            ),
          ],
        ),
      ),
    );
  }
}
