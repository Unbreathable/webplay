import 'dart:math';

import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';

const noTextHeight = TextHeightBehavior(
  applyHeightToFirstAscent: false,
  applyHeightToLastDescent: false,
);

Widget verticalSpacing(double height) {
  return SizedBox(height: height);
}

Widget horizontalSpacing(double width) {
  return SizedBox(width: width);
}

String getRandomString(int length) {
  const chars = 'AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz1234567890';
  final random = Random();
  return String.fromCharCodes(List.generate(length, (index) => chars.codeUnitAt(random.nextInt(chars.length))));
}

void popAllAndPush<T>(BuildContext context, Route<T> route) {
  final nav = Navigator.of(context);
  nav.popUntil((_) => false);
  nav.push(route);
}

const defaultSpacing = 8.0;
const elementSpacing = defaultSpacing * 0.5;
const elementSpacing2 = elementSpacing * 1.5;
const sectionSpacing = defaultSpacing * 2;
const modelBorderRadius = defaultSpacing * 1.5;
const modelPadding = defaultSpacing * 2;
const dialogBorderRadius = defaultSpacing * 1.5;
const dialogPadding = defaultSpacing * 1.5;
const scaleAnimationCurve = ElasticOutCurve(1.1);

/// The most advanced logging framework in the world
void sendLog(Object? object) {
  print(object);
}

class ExpandEffect extends CustomEffect {
  ExpandEffect({super.curve, super.duration, Axis? axis, Alignment? alignment, double? customHeightFactor, super.delay})
      : super(builder: (context, value, child) {
          return ClipRect(
            child: Align(
              alignment: alignment ?? Alignment.topCenter,
              heightFactor: customHeightFactor ?? (axis == Axis.vertical ? max(value, 0.0) : null),
              widthFactor: axis == Axis.horizontal ? max(value, 0.0) : null,
              child: child,
            ),
          );
        });
}
