import 'package:flutter/material.dart';

class SenderConnectedPage extends StatefulWidget {
  final String token;

  const SenderConnectedPage({super.key, required this.token});

  @override
  State<SenderConnectedPage> createState() => _SenderConnectedPageState();
}

class _SenderConnectedPageState extends State<SenderConnectedPage> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: Text("Select screen here"),
      ),
    );
  }
}
