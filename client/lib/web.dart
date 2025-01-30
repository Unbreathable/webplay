import 'dart:convert';

import 'package:dio/dio.dart';

final dio = Dio();

final basePath = "http://localhost:3000";

/// Send a post request to the server.
///
/// Returns response or error.
Future<(Response?, Object?)> postRq(String path, Map<String, dynamic> body) async {
  try {
    return (
      await dio.post(
        "$basePath$path",
        data: jsonEncode(body),
        options: Options(
          headers: {
            "Content-Type": "application/json",
          },
          validateStatus: (status) => true,
        ),
      ),
      null,
    );
  } catch (e) {
    return (null, e);
  }
}
