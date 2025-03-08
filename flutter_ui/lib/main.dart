import 'package:flutter/material.dart';
import 'services/auth_service.dart';
import 'home_screen.dart';
import 'login_screen.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'QA Test Manager',
      theme: ThemeData(
        primarySwatch: Colors.blue,
      ),
      home: const AuthenticationCheck(),
    );
  }
}

class AuthenticationCheck extends StatefulWidget {
  const AuthenticationCheck({super.key});

  @override
  State<AuthenticationCheck> createState() => _AuthenticationCheckState();
}

class _AuthenticationCheckState extends State<AuthenticationCheck> {
  bool _isAuthenticated = false;
  bool _isCheckingAuth = true;

  @override
  void initState() {
    super.initState();
    _checkAuthentication();
  }

  Future<void> _checkAuthentication() async {
    final isValidToken = await AuthService.validateToken();
    setState(() {
      _isAuthenticated = isValidToken;
      _isCheckingAuth = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    if (_isCheckingAuth) {
      return const Scaffold(
        body: Center(child: CircularProgressIndicator()),
      );
    } else {
      return _isAuthenticated ? const HomeScreen() : const LoginScreen();
    }
  }
}

// Placeholder for your HomeScreen (create lib/home_screen.dart)
// class HomeScreen extends StatelessWidget {
//   const HomeScreen({super.key});
//   @override
//   Widget build(BuildContext context) {
//     return Scaffold(appBar: AppBar(title: const Text('Home Screen')), body: const Center(child: Text('Logged In!')));
//   }
// }

// // Placeholder for your LoginScreen (create lib/login_screen.dart) - we'll implement this next
// class LoginScreen extends StatelessWidget {
//   const LoginScreen({super.key});
//   @override
//   Widget build(BuildContext context) {
//     return Scaffold(appBar: AppBar(title: const Text('Login')), body: const Center(child: Text('Login Screen')));
//   }
// }