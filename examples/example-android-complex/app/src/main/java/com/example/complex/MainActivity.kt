package com.example.complex

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.*
import androidx.compose.material.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.example.complex.components.CustomHeader
import com.example.complex.components.SharedIcon

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            Column(modifier = Modifier.padding(16.dp)) {
                SharedIcon()
                // Testing cross-file import resolution
                CustomHeader(title = "Complex Example")

                Spacer(modifier = Modifier.height(20.dp))

                // Rule: WCAG 4.1.2 (Button without label) - Error
                Button(onClick = { /* TODO */ }) {
                    // No text child - Error
                }

                Spacer(modifier = Modifier.height(20.dp))

                // Rule: WCAG 4.1.2 (Button with explicit contentDescription) - OK
                Button(
                    onClick = { /* TODO */ },
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Text("Submit Form")
                }

                Spacer(modifier = Modifier.height(20.dp))

                // Rule: Hidden Focus equivalent (invisibleToUser)
                // In Compose, we can use Modifier.semantics { invisibleToUser() }
                // or just clear the semantics.
                Surface(
                    modifier = Modifier.padding(8.dp),
                    // This hides the entire surface and its children from screen readers
                ) {
                    IconButton(onClick = { /* TODO */ }) {
                        // This icon button might be hidden but still focusable if not handled
                    }
                }
            }
        }
    }
}
