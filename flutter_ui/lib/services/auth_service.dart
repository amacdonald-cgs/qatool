import 'package:http/http.dart' as http;
import 'dart:convert';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class AuthService {
  static const String _baseUrl =
      'http://localhost:8000/api'; // Base URL for your Go backend API
  static const _storage = FlutterSecureStorage();

  // Function to login and store the JWT
  static Future<bool> login(String email, String password) async {
    final url = Uri.parse('$_baseUrl/login');
    final headers = {'Content-Type': 'application/json'};
    final body =
        jsonEncode({'email': email, 'password': password, 'token_2fa': ''});
    final response = await http.post(url, headers: headers, body: body);

    if (response.statusCode == 200) {
      final decodedResponse = jsonDecode(response.body);
      if (decodedResponse['status'] == true) {
        final token = decodedResponse['token'];
        await _storage.write(
            key: 'jwt_token', value: token); // Store the token securely
        return true;
      }
    }
    return false;
  }

  // Function to retrieve the stored JWT
  static Future<String?> getToken() async {
    return await _storage.read(key: 'jwt_token');
  }

  // Function to check if the user is logged in (token exists)
  static Future<bool> isLoggedIn() async {
    final token = await getToken();
    return token != null;
  }

  // Function to clear the JWT (logout)
  static Future<void> logout() async {
    await _storage.delete(key: 'jwt_token');
  }

  // Function to add the Authorization header to API requests
  static Future<Map<String, String>> getAuthHeaders() async {
    final token = await getToken();
    if (token != null) {
      return {'Authorization': 'Bearer $token'};
    } else {
      return {}; // Return empty headers if no token
    }
  }
}
