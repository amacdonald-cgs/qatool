import 'package:http/http.dart' as http;
import 'dart:convert';

class AuthService {
  static const String _baseUrl =
      'http://localhost:8000/api'; // Base URL for your Go backend API

  // Function to call the token validation endpoint on your Go backend
  static Future<bool> validateToken() async {
    final url = Uri.parse('$_baseUrl/token');
    final response = await http
        .get(url); // No headers needed, cookies are handled by browser

    if (response.statusCode == 200) {
      final decodedResponse = jsonDecode(response.body);
      return decodedResponse['status'] == true;
    } else {
      return false;
    }
  }

  // Function to call the login endpoint on your Go backend
  static Future<bool> login(String email, String password) async {
    final url = Uri.parse('$_baseUrl/login');
    final headers = {'Content-Type': 'application/json'};
    final body =
        jsonEncode({'email': email, 'password': password, 'token_2fa': ''});
    final response = await http.post(url, headers: headers, body: body);

    if (response.statusCode == 200) {
      // Cookies are now handled by the Go backend setting Set-Cookie headers
      return jsonDecode(response.body)['status'] == true;
    } else {
      return false;
    }
  }
}
