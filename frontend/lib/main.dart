import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

const String backendBaseUrl = 'http://104.197.178.51:8080';

void main() {
  runApp(const PokerApp());
}

class PokerApp extends StatelessWidget {
  const PokerApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Texas Hold\'em Calculator',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.deepPurple),
        useMaterial3: true,
      ),
      home: const PokerHomePage(),
    );
  }
}

/// Simple value object for a single card selection.
class CardSelection {
  final String? suit; // H, D, C, S
  final String? rank; // 2-9, T, J, Q, K, A

  const CardSelection({this.suit, this.rank});

  String? get code {
    if (suit == null || rank == null) return null;
    return '$suit$rank';
  }

  CardSelection copyWith({String? suit, String? rank}) {
    return CardSelection(
      suit: suit ?? this.suit,
      rank: rank ?? this.rank,
    );
  }
}

class PokerHomePage extends StatefulWidget {
  const PokerHomePage({super.key});

  @override
  State<PokerHomePage> createState() => _PokerHomePageState();
}

class _PokerHomePageState extends State<PokerHomePage> {
  // Card selections
  final List<CardSelection> _heroHole = [
    const CardSelection(),
    const CardSelection(),
  ];
  final List<CardSelection> _community = List.generate(5, (_) => const CardSelection());

  // Community card count (0,3,4,5)
  final List<int> _communityOptions = const [0, 3, 4, 5];
  int _communityCount = 5;

  // Simulation settings
  int _playerCount = 2; // 2–10
  final TextEditingController _trialsController =
      TextEditingController(text: '2000');

  // Evaluate Best Hand result
  String? _handCategory;
  List<String>? _handKickers;

  // Simulation result
  double? _heroWinPct;
  double? _villainWinPct;
  double? _tiePct;
  int? _trialsRun;

  bool _loadingEval = false;
  bool _loadingSim = false;

  @override
  void dispose() {
    _trialsController.dispose();
    super.dispose();
  }

  // ---------- Helpers ----------

  List<String> _collectHeroCodes() {
    return _heroHole.map((c) => c.code).whereType<String>().toList();
  }

  List<String> _collectCommunityCodes() {
    return _community
        .take(_communityCount)
        .map((c) => c.code)
        .whereType<String>()
        .toList();
  }

  bool _validateForEvaluate() {
    final messenger = ScaffoldMessenger.of(context);
    final hero = _collectHeroCodes();
    final community = _collectCommunityCodes();

    if (hero.length != 2) {
      messenger.showSnackBar(
        const SnackBar(content: Text('Please select exactly 2 hero hole cards.')),
      );
      return false;
    }

    if (_communityCount != 0 && community.length != _communityCount) {
      messenger.showSnackBar(
        SnackBar(
          content: Text('Please select exactly $_communityCount community cards.'),
        ),
      );
      return false;
    }

    if (hero.length + community.length < 5) {
      messenger.showSnackBar(
        const SnackBar(
          content: Text(
              'Need at least 5 total cards (2 hole + at least 3 community) to evaluate.'),
        ),
      );
      return false;
    }

    // For /api/evaluate we require 5 community cards (full board).
    if (community.length != 5) {
      messenger.showSnackBar(
        const SnackBar(
          content: Text(
              'Best-hand evaluation requires a full board (5 community cards).'),
        ),
      );
      return false;
    }

    // Simple duplicate check
    final all = [...hero, ...community];
    final set = all.toSet();
    if (set.length != all.length) {
      messenger.showSnackBar(
        const SnackBar(content: Text('Cards must be unique (no duplicates).')),
      );
      return false;
    }

    return true;
  }

  bool _validateForSimulation() {
    final messenger = ScaffoldMessenger.of(context);
    final hero = _collectHeroCodes();
    final community = _collectCommunityCodes();

    if (hero.length != 2) {
      messenger.showSnackBar(
        const SnackBar(content: Text('Please select exactly 2 hero hole cards.')),
      );
      return false;
    }

    if (!(_communityCount == 0 ||
        _communityCount == 3 ||
        _communityCount == 4 ||
        _communityCount == 5)) {
      messenger.showSnackBar(
        const SnackBar(
          content: Text('Community count must be 0, 3, 4, or 5 for simulation.'),
        ),
      );
      return false;
    }

    if (_communityCount != 0 && community.length != _communityCount) {
      messenger.showSnackBar(
        SnackBar(
          content: Text('Please select exactly $_communityCount community cards.'),
        ),
      );
      return false;
    }

    final all = [...hero, ...community];
    final set = all.toSet();
    if (set.length != all.length) {
      messenger.showSnackBar(
        const SnackBar(content: Text('Cards must be unique (no duplicates).')),
      );
      return false;
    }

    final trials = int.tryParse(_trialsController.text.trim());
    if (trials == null || trials <= 0) {
      messenger.showSnackBar(
        const SnackBar(
            content:
                Text('Please enter a valid positive number of simulations.')),
      );
      return false;
    }

    if (_playerCount < 2 || _playerCount > 10) {
      messenger.showSnackBar(
        const SnackBar(
            content: Text('Number of players must be between 2 and 10.')),
      );
      return false;
    }

    return true;
  }

  // ---------- API Calls ----------

  Future<void> _evaluateBestHand() async {
    if (!_validateForEvaluate()) return;

    final hero = _collectHeroCodes();
    final community = _collectCommunityCodes();

    setState(() {
      _loadingEval = true;
      _handCategory = null;
      _handKickers = null;
    });

    try {
      final resp = await http.post(
        Uri.parse('$backendBaseUrl/api/evaluate'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({
          'hole': hero,
          'community': community,
        }),
      );

      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        setState(() {
          _handCategory = data['category'] as String?;
          _handKickers =
              (data['kickers'] as List<dynamic>).map((e) => e.toString()).toList();
        });
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content:
                Text('Error ${resp.statusCode}: ${resp.body} (Evaluate Best Hand)'),
          ),
        );
      }
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Network error during evaluation: $e')),
      );
    } finally {
      if (mounted) {
        setState(() => _loadingEval = false);
      }
    }
  }

  Future<void> _simulateEquity() async {
    if (!_validateForSimulation()) return;

    final hero = _collectHeroCodes();
    final community = _collectCommunityCodes();
    final trials = int.parse(_trialsController.text.trim());
    final numOpponents = _playerCount - 1; // hero + N opponents

    setState(() {
      _loadingSim = true;
      _heroWinPct = null;
      _villainWinPct = null;
      _tiePct = null;
      _trialsRun = null;
    });

    try {
      final resp = await http.post(
        Uri.parse('$backendBaseUrl/api/simulate'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({
          'hole': hero,
          'community': community,
          'numOpponents': numOpponents,
          'trials': trials,
        }),
      );

      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        setState(() {
          _heroWinPct = (data['heroWinPct'] as num).toDouble();
          _villainWinPct = (data['villainWinPct'] as num).toDouble();
          _tiePct = (data['tiePct'] as num).toDouble();
          _trialsRun = data['trialsRun'] as int?;
        });
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content:
                Text('Error ${resp.statusCode}: ${resp.body} (Simulation)'),
          ),
        );
      }
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Network error during simulation: $e')),
      );
    } finally {
      if (mounted) {
        setState(() => _loadingSim = false);
      }
    }
  }

  // ---------- UI ----------

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Texas Hold\'em Evaluator'),
        centerTitle: false,
      ),
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 900),
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: ListView(
              children: [
                const Text(
                  'Build your hand and simulate equity against multiple opponents.',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500),
                ),
                const SizedBox(height: 16),
                _buildCardInputSection(),
                const SizedBox(height: 24),
                _buildSimulationSettings(),
                const SizedBox(height: 24),
                _buildActionButtons(),
                const SizedBox(height: 24),
                _buildResultsSection(),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildCardInputSection() {
    return Card(
      elevation: 1,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Cards',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            const Text(
              'Select suits (H, D, C, S) and ranks (2–9, T, J, Q, K, A). '
              'Each card must be unique.',
              style: TextStyle(fontSize: 13),
            ),
            const SizedBox(height: 16),
            const Text(
              'Hero Hole Cards (2)',
              style: TextStyle(fontWeight: FontWeight.w600),
            ),
            const SizedBox(height: 8),
            Wrap(
              spacing: 12,
              runSpacing: 8,
              children: List.generate(
                2,
                (index) => CardPicker(
                  label: 'Card ${index + 1}',
                  selection: _heroHole[index],
                  onChanged: (sel) {
                    setState(() => _heroHole[index] = sel);
                  },
                ),
              ),
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                const Text(
                  'Community Cards',
                  style: TextStyle(fontWeight: FontWeight.w600),
                ),
                const SizedBox(width: 16),
                const Text('Count:'),
                const SizedBox(width: 8),
                DropdownButton<int>(
                  value: _communityCount,
                  items: _communityOptions
                      .map(
                        (c) => DropdownMenuItem<int>(
                          value: c,
                          child: Text(c.toString()),
                        ),
                      )
                      .toList(),
                  onChanged: (val) {
                    if (val == null) return;
                    setState(() => _communityCount = val);
                  },
                ),
                const SizedBox(width: 8),
                const Text(
                  '(0, 3, 4, or 5 required for simulation; 5 required for full evaluation.)',
                  style: TextStyle(fontSize: 11),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Wrap(
              spacing: 12,
              runSpacing: 8,
              children: List.generate(
                5,
                (index) {
                  final enabled = index < _communityCount || _communityCount == 0;
                  return Opacity(
                    opacity: enabled ? 1.0 : 0.35,
                    child: IgnorePointer(
                      ignoring: !enabled,
                      child: CardPicker(
                        label: 'Board ${index + 1}',
                        selection: _community[index],
                        onChanged: (sel) {
                          setState(() => _community[index] = sel);
                        },
                      ),
                    ),
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSimulationSettings() {
    return Card(
      elevation: 1,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Simulation Settings',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                const Text('Number of Players'),
                const SizedBox(width: 16),
                Expanded(
                  child: Slider(
                    value: _playerCount.toDouble(),
                    min: 2,
                    max: 10,
                    divisions: 8,
                    label: '$_playerCount',
                    onChanged: (v) {
                      setState(() => _playerCount = v.round());
                    },
                  ),
                ),
                SizedBox(
                  width: 32,
                  child: Text(
                    '$_playerCount',
                    textAlign: TextAlign.center,
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                const Text('Monte Carlo Simulations'),
                const SizedBox(width: 16),
                SizedBox(
                  width: 120,
                  child: TextField(
                    controller: _trialsController,
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(
                      border: OutlineInputBorder(),
                      isDense: true,
                      contentPadding:
                          EdgeInsets.symmetric(horizontal: 8, vertical: 8),
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                const Text(
                  '(e.g. 2000–20000; higher = slower but more accurate)',
                  style: TextStyle(fontSize: 11),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildActionButtons() {
    return Row(
      children: [
        Expanded(
          child: FilledButton.tonal(
            onPressed: _loadingEval ? null : _evaluateBestHand,
            child: _loadingEval
                ? const SizedBox(
                    height: 18,
                    width: 18,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text('Evaluate Best Hand'),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: FilledButton(
            onPressed: _loadingSim ? null : _simulateEquity,
            child: _loadingSim
                ? const SizedBox(
                    height: 18,
                    width: 18,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      color: Colors.white,
                    ),
                  )
                : const Text('Simulate Equity'),
          ),
        ),
      ],
    );
  }

  Widget _buildResultsSection() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Card(
          elevation: 1,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Best Hand Evaluation',
                  style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                ),
                const SizedBox(height: 12),
                if (_loadingEval)
                  const Center(
                    child: CircularProgressIndicator(),
                  )
                else if (_handCategory == null)
                  const Text(
                    'No evaluation yet. Select cards and tap "Evaluate Best Hand".',
                    style: TextStyle(fontSize: 13),
                  )
                else
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        _handCategory ?? '',
                        style: const TextStyle(
                          fontSize: 20,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        'Kickers: ${_handKickers?.join(' ')}',
                        style: const TextStyle(fontSize: 14),
                      ),
                    ],
                  ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 16),
        Card(
          elevation: 1,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Monte Carlo Equity',
                  style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                ),
                const SizedBox(height: 12),
                if (_loadingSim)
                  Column(
                    children: const [
                      Center(child: CircularProgressIndicator()),
                      SizedBox(height: 8),
                      Text('Running simulations...'),
                    ],
                  )
                else if (_heroWinPct == null)
                  const Text(
                    'No simulation yet. Adjust players/trials and tap "Simulate Equity".',
                    style: TextStyle(fontSize: 13),
                  )
                else
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _buildEquityBar(
                        label: 'Hero Win',
                        value: _heroWinPct!,
                        color: Colors.green,
                      ),
                      const SizedBox(height: 8),
                      _buildEquityBar(
                        label: 'Villain Win (any opponent)',
                        value: _villainWinPct!,
                        color: Colors.redAccent,
                      ),
                      const SizedBox(height: 8),
                      _buildEquityBar(
                        label: 'Tie',
                        value: _tiePct!,
                        color: Colors.orange,
                      ),
                      const SizedBox(height: 8),
                      Text(
                        'Trials run: ${_trialsRun ?? 0}',
                        style: const TextStyle(fontSize: 12),
                      ),
                    ],
                  ),
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildEquityBar({
    required String label,
    required double value,
    required Color color,
  }) {
    final v = (value / 100).clamp(0.0, 1.0);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Expanded(child: Text(label)),
            Text('${value.toStringAsFixed(1)} %'),
          ],
        ),
        const SizedBox(height: 4),
        ClipRRect(
          borderRadius: BorderRadius.circular(6),
          child: LinearProgressIndicator(
            value: v,
            minHeight: 8,
            backgroundColor: Colors.grey.shade300,
            valueColor: AlwaysStoppedAnimation<Color>(color),
          ),
        ),
      ],
    );
  }
}

/// Reusable widget for selecting a single card (suit + rank).
class CardPicker extends StatelessWidget {
  const CardPicker({
    super.key,
    required this.label,
    required this.selection,
    required this.onChanged,
  });

  final String label;
  final CardSelection selection;
  final ValueChanged<CardSelection> onChanged;

  static const _suits = ['H', 'D', 'C', 'S'];
  static const _suitLabels = {
    'H': '♥ H',
    'D': '♦ D',
    'C': '♣ C',
    'S': '♠ S',
  };
  static const _ranks = ['2', '3', '4', '5', '6', '7', '8', '9', 'T', 'J', 'Q', 'K', 'A'];

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 170,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: const TextStyle(fontSize: 12)),
          const SizedBox(height: 4),
          Row(
            children: [
              Expanded(
                flex: 1,
                child: DropdownButtonFormField<String>(
                  value: selection.suit,
                  isExpanded: true,
                  decoration: const InputDecoration(
                    labelText: 'Suit',
                    isDense: true,
                    border: OutlineInputBorder(),
                  ),
                  items: _suits
                      .map(
                        (s) => DropdownMenuItem<String>(
                          value: s,
                          child: Text(_suitLabels[s] ?? s),
                        ),
                      )
                      .toList(),
                  onChanged: (val) {
                    onChanged(selection.copyWith(suit: val));
                  },
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                flex: 1,
                child: DropdownButtonFormField<String>(
                  value: selection.rank,
                  isExpanded: true,
                  decoration: const InputDecoration(
                    labelText: 'Rank',
                    isDense: true,
                    border: OutlineInputBorder(),
                  ),
                  items: _ranks
                      .map(
                        (r) => DropdownMenuItem<String>(
                          value: r,
                          child: Text(r),
                        ),
                      )
                      .toList(),
                  onChanged: (val) {
                    onChanged(selection.copyWith(rank: val));
                  },
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}