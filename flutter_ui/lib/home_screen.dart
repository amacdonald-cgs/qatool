import 'package:flutter/material.dart';
import 'services/auth_service.dart'; // Import AuthService
import 'login_screen.dart';
import 'dart:convert'; // Add this import

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  String _protectedData = 'Loading...';

  @override
  void initState() {
    super.initState();
    _fetchProtectedData(); // Fetch data when the screen loads
  }

  Future<void> _fetchProtectedData() async {
    final response =
        await AuthService.authenticatedGet('/protected'); // Corrected call

    if (response.statusCode == 200) {
      final decodedResponse = jsonDecode(response.body); // Use jsonDecode
      setState(() {
        _protectedData =
            'Data: ${decodedResponse['message']}'; // Update UI with data
      });
    } else {
      setState(() {
        _protectedData = 'Error: Failed to fetch data';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Home Screen')),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(_protectedData),
            ElevatedButton(
                onPressed: () async {
                  await AuthService.logout();
                  if (!mounted) return;
                  Navigator.of(context).pushReplacement(
                    MaterialPageRoute(
                        builder: (context) => const LoginScreen()),
                  );
                },
                child: const Text("Logout"))
          ],
        ),
      ),
    );
  }
}
