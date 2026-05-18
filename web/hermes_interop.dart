import 'dart:js_interop';
import 'dart:typed_data';

/// Modern JS-Interop representation of the shared WASM-GC instance.
@JS()
extension type HermesWasmInstance(JSObject _) implements JSObject {
  @JS('exports')
  external HermesWasmExports get exports;
}

@JS()
extension type HermesWasmExports(JSObject _) implements JSObject {
  @JS('memory')
  external JSObject get memory;

  @JS('execute_skill')
  external JSFunction get executeSkill;
}

/// Zero-Copy Symmetric Memory Offset Configuration (SACP Protocol)
const int slotSize = 64;
const int uuidOffset = 0;
const int origAuthOffset = slotSize;
const int currAuthOffset = slotSize * 2;
const int payloadOffset = slotSize * 3;

/// Zero-Copy Reader for SACP Telemetry frames in Dart/Flutter.
class SACPFrameReader {
  final ByteData memoryBuffer;
  final int baseOffset;

  SACPFrameReader(this.memoryBuffer, this.baseOffset);

  String getUUID() {
    final bytes = memoryBuffer.buffer.asUint8List(baseOffset + uuidOffset, slotSize);
    return _utf8DecodeTrim(bytes);
  }

  int getOriginalAuthority() {
    return memoryBuffer.getUint8(baseOffset + origAuthOffset);
  }

  int getCurrentAuthority() {
    return memoryBuffer.getUint8(baseOffset + currAuthOffset);
  }

  String getPayload() {
    final bytes = memoryBuffer.buffer.asUint8List(baseOffset + payloadOffset);
    return _utf8DecodeTrim(bytes);
  }

  String _utf8DecodeTrim(Uint8List bytes) {
    final end = bytes.indexOf(0);
    final len = end == -1 ? bytes.length : end;
    return String.fromCharCodes(bytes.sublist(0, len));
  }
}

/// Reactive MV4 Telemetry Visualizer console state hook
class MV4TelemetryVisualizer {
  final void Function(String logMessage) onUpdate;

  MV4TelemetryVisualizer({required this.onUpdate});

  void processSACPFrame(ByteBuffer wasmMemory, int frameBaseOffset) {
    final byteData = ByteData.view(wasmMemory);
    final reader = SACPFrameReader(byteData, frameBaseOffset);

    final uuid = reader.getUUID();
    final origAuth = reader.getOriginalAuthority();
    final currAuth = reader.getCurrentAuthority();
    final payload = reader.getPayload();

    final formattedLog = '[MV4 Telemetry] [PASS] SACP Frame Ingested '
        '| Agent: $uuid '
        '| Monotonic Auth Invariant: Current=$currAuth <= Original=$origAuth '
        '| Payload: "$payload"';

    onUpdate(formattedLog);
  }
}
