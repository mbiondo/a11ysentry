package com.example.complex.utils

import androidx.compose.runtime.Composable
import com.example.complex.MainActivity

@Composable
fun ThemeContext() {
    // Circular dependency for testing tree generator
    // This is a bad pattern in real life but essential for stress testing
}
