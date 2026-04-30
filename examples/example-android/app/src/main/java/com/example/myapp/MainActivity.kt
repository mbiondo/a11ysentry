package com.example.myapp

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.material3.Text
import androidx.compose.material3.Button
import androidx.compose.foundation.Image
import androidx.compose.runtime.Composable
import androidx.compose.ui.res.painterResource

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            MyScreen()
        }
    }
}

@Composable
fun MyScreen() {
    // Accessibility Issue: Image without contentDescription
    Image(
        painter = painterResource(id = 123),
        contentDescription = null 
    )

    // Accessibility Issue: Button without proper label/text
    Button(onClick = { }) {
        // Empty button
    }

    Text("Hello A11ySentry!")
}
