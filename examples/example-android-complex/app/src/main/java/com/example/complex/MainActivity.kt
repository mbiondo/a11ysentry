package com.example.complex

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import com.example.complex.components.CustomHeader

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            // Testing cross-file import resolution
            CustomHeader(title = "Complex Example")
        }
    }
}
