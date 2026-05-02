import 'package:flutter/material.dart';
import 'package:example_flutter_complex/components/header.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      home: Scaffold(
        appBar: const PreferredSize(
          preferredSize: Size.fromHeight(60),
          child: CustomHeader(title: "Complex Flutter A11y"),
        ),
        body: SingleChildScrollView(
          child: Column(
            children: [
              // Rule: WCAG 1.1.1 (Image without semanticLabel)
              Image.asset(
                'assets/logo.png',
                // Missing semanticLabel - Error
              ),

              const SizedBox(height: 20),

              // Rule: WCAG 1.1.1 (Image with semanticLabel) - OK
              Image.asset(
                'assets/banner.png',
                semanticLabel: "A11ySentry Banner",
              ),

              const SizedBox(height: 20),

              // Rule: WCAG 4.1.2 (Button with label via child Text) - OK
              ElevatedButton(
                onPressed: () {},
                child: const Text("Save Changes"),
              ),

              const SizedBox(height: 20),

              // Rule: WCAG 4.1.2 (Button without label) - Error
              IconButton(
                icon: const Icon(Icons.add),
                onPressed: () {},
                // Missing tooltip or Semantics label
              ),

              const SizedBox(height: 20),

              // Rule: Semantics wrapper - OK
              Semantics(
                label: "Profile Picture",
                child: Container(
                  width: 100,
                  height: 100,
                  decoration: const BoxDecoration(
                    image: DecorationImage(image: AssetImage('profile.png')),
                  ),
                ),
              ),

              const SizedBox(height: 20),

              // Rule: Hidden Focus equivalent (Focusable in hidden Semantics)
              Semantics(
                excludeSemantics: true, // Hides from screen readers
                child: TextButton(
                  onPressed: () {},
                  child: const Text("I am hidden from screen readers"),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
