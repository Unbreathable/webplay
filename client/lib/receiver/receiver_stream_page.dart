import 'dart:async';

import 'package:client/lobby_page.dart';
import 'package:client/vertical_spacing.dart';
import 'package:client/web.dart';
import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_webrtc/flutter_webrtc.dart';
import 'package:signals/signals_flutter.dart';

class ReceiverStreamPage extends StatefulWidget {
  final String token;

  const ReceiverStreamPage({super.key, required this.token});

  @override
  State<ReceiverStreamPage> createState() => _ReceiverStreamPageState();
}

class _ReceiverStreamPageState extends State<ReceiverStreamPage> with SignalsMixin {
  RTCPeerConnection? _peer;
  final _remoteRenderer = RTCVideoRenderer();
  final connected = signal(false);

  @override
  void initState() {
    super.initState();
    _remoteRenderer.initialize();
  }

  @override
  void dispose() {
    _remoteRenderer.dispose();
    _peer?.close();
    _peer?.dispose();
    super.dispose();
  }

  void createConnection() async {
    // Wait quickly
    await Future.delayed(500.ms);

    // Create a new webrtc connection
    _peer = await createPeerConnection({
      "iceServers": [
        {
          "urls": [
            "stun:stun.l.google.com:19302",
          ]
        }
      ],
    });

    // Add connection listener
    _peer!.onConnectionState = (state) {
      sendLog("new connection state: $state");
    };

    _peer!.onIceConnectionState = (state) {
      sendLog("ICE connection state: $state");
    };

    // Add track listener
    _peer!.onTrack = (RTCTrackEvent event) {
      sendLog("received track");
      if (event.track.kind == "video") {
        sendLog("received video track ${event.streams.length} ${event.track.label}");
        _remoteRenderer.srcObject = event.streams[0];
        connected.value = true;
      }
    };

    // Add the transceiver
    await _peer!.addTransceiver(
      kind: RTCRtpMediaType.RTCRtpMediaTypeVideo,
      init: RTCRtpTransceiverInit(
        direction: TransceiverDirection.RecvOnly,
      ),
    );

    // Create the offer
    final offer = await _peer!.createOffer({});
    await _peer!.setLocalDescription(offer);
    final completer = Completer<bool>();
    _peer!.onIceCandidate = (candidate) {
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
    final (res, error) = await postRq("/receiver/connect", {
      "token": widget.token,
      "offer": (await _peer!.getLocalDescription())!.toMap(),
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
    await _peer!.setRemoteDescription(RTCSessionDescription(json["sdp"], json["type"]));

    sendLog("receiver: webrtc connection *technically* finished");
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Watch.builder(builder: (context) {
        if (connected.value) {
          return Stack(
            children: [
              RTCVideoView(_remoteRenderer),
              Align(
                alignment: Alignment.bottomCenter,
                child: Padding(
                  padding: const EdgeInsets.all(defaultSpacing),
                  child: Row(
                    children: [
                      ElevatedButton(
                        onPressed: () => popAllAndPush(context, MaterialPageRoute(builder: (context) => LobbyPage())),
                        child: Text("Return to Lobby"),
                      ),
                    ],
                  ),
                ),
              )
            ],
          );
        }

        return Center(
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
        ));
      }),
    );
  }
}
