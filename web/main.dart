import 'package:flutter/material.dart';
import 'dart:typed_data';

void main() {
  runApp(const AetherSovereignSurface());
}

/// The Conformed Aether Sovereign presentation surface using Material Design V3.
class AetherSovereignSurface extends StatelessWidget {
  const AetherSovereignSurface({super.key});

  @override
  Widget build(BuildContext context) {
    // 47 role-based M3 color scheme derived from primary Crimson, secondary Teal, and tertiary Onyx.
    final colorScheme = ColorScheme.fromSeed(
      seedColor: const Color(0xFFBF1E2D), // Crimson Seed
      primary: const Color(0xFFBF1E2D),   // Crimson
      secondary: const Color(0xFF00A2E8), // Teal Accent
      tertiary: const Color(0xFF0A0A0A),  // Onyx Base
      surface: const Color(0xFF0D0D11),   // Dark Obsidian
      brightness: Brightness.dark,
    );

    return MaterialApp(
      title: 'Aether Sovereign Studio',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        useMaterial3: true,
        colorScheme: colorScheme,
        fontFamily: 'monospace',
        scaffoldBackgroundColor: colorScheme.surface,
      ),
      home: const AetherDashboardPage(),
    );
  }
}

class AetherDashboardPage extends StatefulWidget {
  const AetherDashboardPage({super.key});

  @override
  State<AetherDashboardPage> createState() => _AetherDashboardPageState();
}

class _AetherDashboardPageState extends State<AetherDashboardPage> {
  final List<String> _telemetryLogs = [];
  late final MV4TelemetryVisualizer _telemetryVisualizer;

  // Shared Memory Offsets Display Data
  int _origAuth = 255;
  int _currAuth = 128;
  String _activeAgentUUID = 'AETH-SOV-092';
  String _latestPayload = 'SYS_INIT_SEQUENCE_OK';

  @override
  void initState() {
    super.initState();
    _telemetryVisualizer = MV4TelemetryVisualizer(
      onUpdate: (logMessage) {
        setState(() {
          _telemetryLogs.insert(0, logMessage);
          if (_telemetryLogs.length > 50) {
            _telemetryLogs.removeLast();
          }
          // Extract values to update interactive dashboard widgets dynamically
          if (logMessage.contains('Agent:')) {
            final parts = logMessage.split('|');
            for (var part in parts) {
              if (part.contains('Agent:')) {
                _activeAgentUUID = part.replaceAll('Agent:', '').trim();
              } else if (part.contains('Current=')) {
                final authPart = part.replaceAll('Monotonic Auth Invariant:', '').trim();
                final authVals = authPart.split('<=').map((e) => e.replaceAll(RegExp(r'[^\d]'), '')).toList();
                if (authVals.length >= 2) {
                  _currAuth = int.tryParse(authVals[0]) ?? _currAuth;
                  _origAuth = int.tryParse(authVals[1]) ?? _origAuth;
                }
              } else if (part.contains('Payload:')) {
                _latestPayload = part.replaceAll('Payload:', '').replaceAll('"', '').trim();
              }
            }
          }
        });
      },
    );

    // Bootstrap local mock server ticks to demonstrate active WASM-GC heap sharing
    _simulateTelemetryStream();
  }

  void _simulateTelemetryStream() {
    Future.doWhile(() async {
      await Future.delayed(const Duration(milliseconds: 1500));
      if (!mounted) return false;

      // Mock zero-copy WASM shared heap bytes configuration (slotSize = 64)
      final mockMemory = Uint8List(256);
      final byteData = ByteData.view(mockMemory.buffer);

      // Write UUID Offset (Bytes 0-63)
      final uuidBytes = 'AETH-SOV-092-MUTATION-${(100 + _telemetryLogs.length)}'.codeUnits;
      for (int i = 0; i < uuidBytes.length && i < 64; i++) {
        mockMemory[i] = uuidBytes[i];
      }

      // Write Original Authority Offset (Bytes 64-127)
      byteData.setUint8(64, 255);

      // Write Current Authority Offset (Bytes 128-191)
      final mockCurrAuth = 120 - (_telemetryLogs.length % 20);
      byteData.setUint8(128, mockCurrAuth);

      // Write Payload Offset (Bytes 192-255)
      final payloadBytes = 'SACP_DISPATCH_OK_CYCLE_${_telemetryLogs.length}'.codeUnits;
      for (int i = 0; i < payloadBytes.length && i < 64; i++) {
        mockMemory[192 + i] = payloadBytes[i];
      }

      // Ingest the simulated frame from WASM heap
      _telemetryVisualizer.processSACPFrame(mockMemory.buffer, 0);
      return true;
    });
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final isInvariantViolated = _currAuth > _origAuth;

    return Scaffold(
      appBar: AppBar(
        title: const Row(
          children: [
            Icon(Icons.hub_outlined, color: Color(0xFF00A2E8)),
            SizedBox(width: 12),
            Text(
              'AETHER SOVEREIGN ENGINE - MV4 DASHBOARD',
              style: TextStyle(fontWeight: FontWeight.bold, letterSpacing: 1.2),
            ),
          ],
        ),
        elevation: 0,
        backgroundColor: Colors.transparent,
        border: Border(bottom: BorderSide(color: colorScheme.outlineVariant.withOpacity(0.3))),
      ),
      body: Row(
        children: [
          // Sidebar: Shared WebAssembly Heap Map representation
          Container(
            width: 320,
            decoration: BoxDecoration(
              color: colorScheme.surfaceContainerLow.withOpacity(0.6),
              border: Border(right: BorderSide(color: colorScheme.outlineVariant.withOpacity(0.3))),
            ),
            child: Column(
              children: [
                const Padding(
                  padding: EdgeInsets.all(16.0),
                  child: Row(
                    children: [
                      Icon(Icons.memory, size: 18, color: Color(0xFF00A2E8)),
                      SizedBox(width: 8),
                      Text(
                        "WASM-GC HEAP SLOTS",
                        style: TextStyle(fontWeight: FontWeight.bold, letterSpacing: 1.1),
                      ),
                    ],
                  ),
                ),
                const Divider(height: 1),
                Expanded(
                  child: ListView(
                    padding: const EdgeInsets.all(12),
                    children: [
                      _buildHeapSlotWidget(
                        name: "uuidOffset",
                        range: "Bytes 0 - 63",
                        value: _activeAgentUUID,
                        color: colorScheme.primary,
                      ),
                      const SizedBox(height: 12),
                      _buildHeapSlotWidget(
                        name: "origAuthOffset",
                        range: "Bytes 64 - 127",
                        value: "0x${_origAuth.toRadixString(16).toUpperCase()} ($_origAuth)",
                        color: Colors.purpleAccent,
                      ),
                      const SizedBox(height: 12),
                      _buildHeapSlotWidget(
                        name: "currAuthOffset",
                        range: "Bytes 128 - 191",
                        value: "0x${_currAuth.toRadixString(16).toUpperCase()} ($_currAuth)",
                        color: Colors.deepOrangeAccent,
                      ),
                      const SizedBox(height: 12),
                      _buildHeapSlotWidget(
                        name: "payloadOffset",
                        range: "Bytes 192 - 255",
                        value: _latestPayload,
                        color: Colors.greenAccent,
                      ),
                    ],
                  ),
                ),
                // Dynamic SACP Invariant Visualizer
                _buildSACPInvariantWidget(isInvariantViolated),
              ],
            ),
          ),
          // Central Canvas: Live SACP Telemetry Feeds
          Expanded(
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      const Text(
                        "LIVE TELEMETRY BROADCASTS (120Hz)",
                        style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: Colors.grey),
                      ),
                      const Spacer(),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                        decoration: BoxDecoration(
                          color: colorScheme.secondary.withOpacity(0.15),
                          borderRadius: BorderRadius.circular(4),
                          border: Border.all(color: colorScheme.secondary),
                        ),
                        child: Row(
                          children: [
                            const SizedBox(
                              width: 8,
                              height: 8,
                              child: CircularProgressIndicator(
                                strokeWidth: 1.5,
                                color: Color(0xFF00A2E8),
                              ),
                            ),
                            const SizedBox(width: 8),
                            Text(
                              "MUTATION RESOLVER ACTIVE",
                              style: TextStyle(
                                fontSize: 9,
                                color: colorScheme.secondary,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  // Log Terminal Display Panel
                  Expanded(
                    child: Container(
                      width: double.infinity,
                      decoration: BoxDecoration(
                        color: const Color(0xFF050507),
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: colorScheme.outlineVariant.withOpacity(0.3)),
                      ),
                      padding: const EdgeInsets.all(12.0),
                      child: _telemetryLogs.isEmpty
                          ? const Center(
                              child: Text(
                                "WAITING FOR FIRST SACP TELEMETRY FRAME...",
                                style: TextStyle(color: Colors.grey, fontSize: 12),
                              ),
                            )
                          : ListView.builder(
                              itemCount: _telemetryLogs.length,
                              itemBuilder: (context, index) {
                                final isSystemLog = _telemetryLogs[index].contains("[PASS]");
                                return Padding(
                                  padding: const EdgeInsets.symmetric(vertical: 4.0),
                                  child: Row(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        "> ",
                                        style: TextStyle(
                                          color: isSystemLog ? Colors.greenAccent : Colors.redAccent,
                                          fontWeight: FontWeight.bold,
                                        ),
                                      ),
                                      Expanded(
                                        child: Text(
                                          _telemetryLogs[index],
                                          style: TextStyle(
                                            color: isSystemLog ? Colors.green[200] : Colors.grey[400],
                                            fontSize: 12,
                                            height: 1.4,
                                          ),
                                        ),
                                      ),
                                    ],
                                  ),
                                );
                              },
                            ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildHeapSlotWidget({
    required String name,
    required String range,
    required String value,
    required Color color,
  }) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.blueGrey.withOpacity(0.08),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: color.withOpacity(0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Text(
                name,
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: color),
              ),
              const Spacer(),
              Text(
                range,
                style: const TextStyle(fontSize: 10, color: Colors.grey),
              ),
            ],
          ),
          const SizedBox(height: 6),
          Text(
            value,
            style: const TextStyle(fontSize: 12, color: Colors.white, overflow: TextOverflow.ellipsis),
            maxLines: 1,
          ),
        ],
      ),
    );
  }

  Widget _buildSACPInvariantWidget(bool isViolated) {
    return Container(
      margin: const EdgeInsets.all(12),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: isViolated ? Colors.red.withOpacity(0.15) : Colors.cyan.withOpacity(0.08),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: isViolated ? Colors.red : Colors.cyan.withOpacity(0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(
                isViolated ? Icons.gpp_bad : Icons.shield_outlined,
                size: 16,
                color: isViolated ? Colors.redAccent : Colors.cyan,
              ),
              const SizedBox(width: 8),
              Text(
                "SACP INVARIANT STATUS",
                style: TextStyle(
                  fontSize: 10,
                  fontWeight: FontWeight.bold,
                  color: isViolated ? Colors.redAccent : Colors.cyan,
                  letterSpacing: 1.2,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            isViolated ? "INVARIANT BREACH DETECTED" : "SECURITY CONTRAINTS OK",
            style: const TextStyle(fontSize: 11, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 4),
          Text(
            "CurrentAuthority ($_currAuth) <= OriginalAuthority ($_origAuth)",
            style: TextStyle(
              fontSize: 10,
              color: isViolated ? Colors.red[200] : Colors.grey[400],
              fontStyle: FontStyle.italic,
            ),
          ),
        ],
      ),
    );
  }
}
