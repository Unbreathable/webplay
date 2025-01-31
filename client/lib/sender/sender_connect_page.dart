import 'package:flutter/material.dart';

class SenderConnectPage extends StatefulWidget {
  final String token;

  const SenderConnectPage({super.key, required this.token});

  @override
  State<SenderConnectPage> createState() => _SenderConnectPageState();
}

class _SenderConnectPageState extends State<SenderConnectPage> {
  @override
  void initState() {
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: Text("Select screen here"),
      ),
    );
  }
}
