import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'QA Test Manager',
      theme: ThemeData(
        primarySwatch: Colors.blue,
      ),
      home: const MyHomePage(title: 'QA Test Manager'),
    );
  }
}

class MyHomePage extends StatefulWidget {
  const MyHomePage({Key? key, required this.title}) : super(key: key);

  final String title;

  @override
  _MyHomePageState createState() => _MyHomePageState();
}

class _MyHomePageState extends State<MyHomePage> {
  String _apiResponse = 'Press the button to ping the API';

  Future<void> _pingApi() async {
    // make API call
    // log function call
    print('Pinging API');
    // add try catch block
    http.Response response;
    try {
      response = await http.get(Uri.parse('http://localhost:3000/ping'));
    } on Exception catch (e) {
      print('Error: ${e.toString()}');
      // TODO
      return;
    }
    // check response status code
    // log response status code
    // print('Response status code: ${}');
    if (response.statusCode == 200) {
      // log success
      print('API call successful');
      final decodedResponse = jsonDecode(response.body);
      setState(() {
        _apiResponse = 'API Response: ${decodedResponse['message']}';
      });
    } else {
      // log error
      print('API call failed with status code ${response.statusCode}');
      setState(() {
        _apiResponse =
            'Error: API call failed with status code ${response.statusCode}';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.title),
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            Text(
              _apiResponse,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 20),
            ElevatedButton(
              onPressed: _pingApi,
              child: const Text('Ping API'),
            ),
          ],
        ),
      ),
    );
  }
}
