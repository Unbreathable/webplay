import 'dart:async';
import 'dart:io';

import 'package:client/lobby_page.dart';
import 'package:client/util/screen_select_dialog.dart';
import 'package:client/vertical_spacing.dart';
import 'package:client/web.dart';
import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_webrtc/flutter_webrtc.dart';

class SenderConnectPage extends StatefulWidget {
  final String token;

  const SenderConnectPage({super.key, required this.token});

  @override
  State<SenderConnectPage> createState() => _SenderConnectPageState();
}

class _SenderConnectPageState extends State<SenderConnectPage> {
  MediaStream? _localStream;

  @override
  void initState() {
    super.initState();
  }

  void createConnection() async {
    // Wait quickly
    await Future.delayed(500.ms);

    // Get the screen/app the user wants to stream
    DesktopCapturerSource? source;
    if (!Platform.isLinux || bool.fromEnvironment("LINUX")) {
      final c = context;
      if (c.mounted) {
        source = await showDialog<DesktopCapturerSource>(context: c, builder: (context) => ScreenSelectDialog());
        if (source == null) {
          return;
        }
      } else {
        return;
      }
    }

    // Create a new webrtc connection
    final peer = await createPeerConnection({
      "iceServers": [
        {
          "urls": [
            "stun:stun.l.google.com:19302",
          ]
        }
      ],
    });

    // Create the stream
    try {
      final stream = await navigator.mediaDevices.getDisplayMedia(<String, dynamic>{
        'video': {
          'deviceId': source != null ? {'exact': source.id} : null,
          'mandatory': {'frameRate': 30.0}
        }
      });

      _localStream = stream;
    } catch (e) {
      sendLog("error while creating stream: ${e.toString()}");
      return;
    }

    // Add the first track
    await peer.addTrack(_localStream!.getVideoTracks()[0], _localStream!);

    // Add connection listener
    peer.onConnectionState = (state) {
      sendLog("new connection state: $state");
    };

    // Create the offer
    final offer = await peer.createOffer({});
    await peer.setLocalDescription(offer);
    final completer = Completer<bool>();
    peer.onIceCandidate = (candidate) {
      if (candidate.candidate != null && !completer.isCompleted) {
        completer.complete(true);
      }
    };
    final success = await completer.future.timeout(
      Duration(seconds: 10),
      onTimeout: () => false,
    );
    if (!success) {
      sendLog("couldn't find ice candidates in time");
      return;
    }

    // Send the offer to the server
    final (res, error) = await postRq("/sender/connect", {
      "token": widget.token,
      "offer": (await peer.getLocalDescription())!.toMap(),
    });
    if (error != null) {
      sendLog("error during offer sending: $error");
      return;
    }
    if ((res!.statusCode ?? 404) != 200) {
      sendLog("error during offer received code ${res.statusCode} ${res.statusMessage}");
      return;
    }

    // Accept the answer from the server
    final json = res.data;
    await peer.setRemoteDescription(RTCSessionDescription(json["sdp"], json["type"]));

    sendLog("sender: webrtc connection *technically* finished");
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
          child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          ElevatedButton(
            onPressed: () => createConnection(),
            child: Text("Create connection"),
          ),
          verticalSpacing(defaultSpacing),
          ElevatedButton(
            onPressed: () => popAllAndPush(context, MaterialPageRoute(builder: (context) => LobbyPage())),
            child: Text("Return to Lobby"),
          ),
        ],
      )),
    );
  }
}
